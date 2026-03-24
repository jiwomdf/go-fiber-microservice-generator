package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"generator-web/domain"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

//go:embed static/*
var StaticFS embed.FS

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		handleGenerate(w, r, repoRoot)
	})

	addr := ":8090"
	log.Printf("generator-web listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	content, err := StaticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "failed to load page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(content)
}

func handleGenerate(w http.ResponseWriter, r *http.Request, repoRoot string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	projectName, authGRPCPort, userGRPCPort, authDB, userDB, services, err := normalizeRequest(req, repoRoot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	archiveName, zipData, err := buildProjectArchive(repoRoot, projectName, authGRPCPort, userGRPCPort, authDB, userDB, services)
	if err != nil {
		log.Printf("generate failed: %v", err)
		http.Error(w, "failed to generate project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", archiveName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
	_, _ = w.Write(zipData)
}

type serviceSpec struct {
	Name     string
	HTTPPort int
	GRPCPort int
	DB       domain.DatabaseConfig
}

func normalizeRequest(req domain.GenerateRequest, repoRoot string) (string, int, int, domain.DatabaseConfig, domain.DatabaseConfig, []serviceSpec, error) {
	projectName := slugify(req.ProjectName)
	if projectName == "" {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, errors.New("project name is required")
	}

	if len(req.Services) == 0 {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, errors.New("at least one service is required")
	}
	if len(req.Services) > 20 {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, errors.New("too many services requested")
	}

	authDB, err := normalizeDBConfig("auth-service", req.AuthDB)
	if err != nil {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, err
	}

	userDB, err := normalizeDBConfig("user-service", req.UserDB)
	if err != nil {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, err
	}

	existingHTTPPorts, existingGRPCPorts, err := existingPorts(repoRoot)
	if err != nil {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, err
	}

	authGRPCPort, err := parsePort(req.AuthGRPCPort, 57000, 59999)
	if err != nil {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("auth-service grpc port: %w", err)
	}
	if authGRPCPort == 0 {
		authGRPCPort = 57704
	}

	userGRPCPort, err := parsePort(req.UserGRPCPort, 57000, 59999)
	if err != nil {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("user-service grpc port: %w", err)
	}
	if userGRPCPort == 0 {
		userGRPCPort = 57705
	}

	if authGRPCPort == userGRPCPort {
		return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, errors.New("auth-service and user-service grpc ports must be different")
	}

	nextHTTPPort := nextPort(existingHTTPPorts, 7703)
	nextGRPCPort := nextPort(existingGRPCPorts, 57703)

	seenNames := map[string]struct{}{}
	seenHTTP := map[int]struct{}{}
	seenGRPC := map[int]struct{}{}
	services := make([]serviceSpec, 0, len(req.Services))
	for _, raw := range req.Services {
		service := slugify(raw.Name)
		if service == "" {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, errors.New("service names must contain letters or numbers")
		}
		if service == "auth" || service == "user" {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("%q is reserved because the base project already contains %s-service", service, service)
		}
		if _, ok := seenNames[service]; ok {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("duplicate service name: %s", service)
		}

		httpPort, err := parsePort(raw.HTTPPort, 7000, 7999)
		if err != nil {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("%s http port: %w", service, err)
		}
		if httpPort == 0 {
			httpPort = nextUnusedPort(nextHTTPPort, existingHTTPPorts, seenHTTP)
			nextHTTPPort = httpPort + 1
		} else {
			if _, ok := existingHTTPPorts[httpPort]; ok {
				return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("http port already in use: %d", httpPort)
			}
			if _, ok := seenHTTP[httpPort]; ok {
				return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("duplicate http port in request: %d", httpPort)
			}
		}

		grpcPort, err := parsePort(raw.GRPCPort, 57000, 59999)
		if err != nil {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("%s grpc port: %w", service, err)
		}
		if grpcPort == 0 {
			grpcPort = nextUnusedPort(nextGRPCPort, existingGRPCPorts, seenGRPC)
			nextGRPCPort = grpcPort + 1
		} else {
			if _, ok := existingGRPCPorts[grpcPort]; ok {
				return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("grpc port already in use: %d", grpcPort)
			}
			if _, ok := seenGRPC[grpcPort]; ok {
				return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("duplicate grpc port in request: %d", grpcPort)
			}
		}

		dbConfig, err := normalizeDBConfig(service+"-service", raw.DB)
		if err != nil {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, err
		}

		if grpcPort == authGRPCPort || grpcPort == userGRPCPort {
			return "", 0, 0, domain.DatabaseConfig{}, domain.DatabaseConfig{}, nil, fmt.Errorf("grpc port already in use: %d", grpcPort)
		}

		seenNames[service] = struct{}{}
		seenHTTP[httpPort] = struct{}{}
		seenGRPC[grpcPort] = struct{}{}
		services = append(services, serviceSpec{
			Name:     service,
			HTTPPort: httpPort,
			GRPCPort: grpcPort,
			DB:       dbConfig,
		})
	}

	return projectName, authGRPCPort, userGRPCPort, authDB, userDB, services, nil
}

func buildProjectArchive(repoRoot, projectName string, authGRPCPort, userGRPCPort int, authDB, userDB domain.DatabaseConfig, services []serviceSpec) (string, []byte, error) {
	tempRoot, err := os.MkdirTemp("", "service-generator-*")
	if err != nil {
		return "", nil, err
	}
	defer os.RemoveAll(tempRoot)

	projectDir := filepath.Join(tempRoot, projectName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return "", nil, err
	}

	if err := copyTemplate(repoRoot, projectDir); err != nil {
		return "", nil, err
	}

	if err := updateServiceConfig(projectDir, "auth-service", authGRPCPort, authDB); err != nil {
		return "", nil, err
	}
	if err := updateServiceConfig(projectDir, "user-service", userGRPCPort, userDB); err != nil {
		return "", nil, err
	}

	for _, service := range services {
		cmd := exec.Command(
			"bash",
			"scripts/scaffold-service.sh",
			"--http-port", strconv.Itoa(service.HTTPPort),
			"--grpc-port", strconv.Itoa(service.GRPCPort),
			service.Name,
		)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", nil, fmt.Errorf("scaffold %s failed: %w: %s", service.Name, err, strings.TrimSpace(string(output)))
		}

		if err := updateServiceConfig(projectDir, service.Name+"-service", service.GRPCPort, service.DB); err != nil {
			return "", nil, err
		}
	}

	zipData, err := zipDirectory(projectDir, projectName)
	if err != nil {
		return "", nil, err
	}

	return projectName + ".zip", zipData, nil
}

func normalizeDBConfig(label string, cfg domain.DatabaseConfig) (domain.DatabaseConfig, error) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.Port = strings.TrimSpace(cfg.Port)
	cfg.Database = strings.TrimSpace(cfg.Database)
	cfg.Username = strings.TrimSpace(cfg.Username)

	if cfg.Host == "" {
		return domain.DatabaseConfig{}, fmt.Errorf("%s database host is required", label)
	}
	if cfg.Port == "" {
		return domain.DatabaseConfig{}, fmt.Errorf("%s database port is required", label)
	}
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return domain.DatabaseConfig{}, fmt.Errorf("%s database port must be numeric", label)
	}
	if cfg.Database == "" {
		return domain.DatabaseConfig{}, fmt.Errorf("%s database name is required", label)
	}
	if cfg.Username == "" {
		return domain.DatabaseConfig{}, fmt.Errorf("%s database username is required", label)
	}

	return cfg, nil
}

func parsePort(raw string, min, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("must be a number")
	}
	if port < min || port > max {
		return 0, fmt.Errorf("must be between %d and %d", min, max)
	}

	return port, nil
}

func updateServiceConfig(projectDir, serviceName string, grpcPort int, db domain.DatabaseConfig) error {
	serviceDir := filepath.Join(projectDir, serviceName)
	for _, configFile := range []string{
		filepath.Join(serviceDir, "config.dev.yaml"),
		filepath.Join(serviceDir, "config.local.yaml"),
	} {
		if err := updateConfigSettings(configFile, grpcPort, db); err != nil {
			return err
		}
	}

	return updateComposeDatabaseEnv(filepath.Join(projectDir, "docker-compose.yml"), serviceName, db)
}

func updateConfigSettings(path string, grpcPort int, db domain.DatabaseConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	inPostgres := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "grpc_port:") {
			lines[i] = `  grpc_port: "` + strconv.Itoa(grpcPort) + `"`
		}
		if trimmed == "postgres:" {
			inPostgres = true
			continue
		}
		if inPostgres && !strings.HasPrefix(line, "  ") && trimmed != "" {
			inPostgres = false
		}
		if !inPostgres {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "host:"):
			lines[i] = `  host: "` + db.Host + `"`
		case strings.HasPrefix(trimmed, "port:"):
			lines[i] = `  port: "` + db.Port + `"`
		case strings.HasPrefix(trimmed, "database:"):
			lines[i] = `  database: "` + db.Database + `"`
		case strings.HasPrefix(trimmed, "user:"):
			lines[i] = `  user: "` + db.Username + `"`
		case strings.HasPrefix(trimmed, "password:"):
			lines[i] = `  password: "` + db.Password + `"`
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func updateComposeDatabaseEnv(path, serviceName string, db domain.DatabaseConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	currentService := ""
	inEnvironment := false
	serviceLinePattern := regexp.MustCompile(`^  ([a-z0-9-]+):\s*$`)

	for i, line := range lines {
		if match := serviceLinePattern.FindStringSubmatch(line); match != nil {
			currentService = match[1]
			inEnvironment = false
			continue
		}

		if currentService != serviceName {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(line, "    environment:") {
			inEnvironment = true
			continue
		}
		if inEnvironment && strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
			inEnvironment = false
		}
		if !inEnvironment {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "POSTGRES_HOST:"):
			lines[i] = `      POSTGRES_HOST: "` + db.Host + `"`
		case strings.HasPrefix(trimmed, "POSTGRES_PORT:"):
			lines[i] = `      POSTGRES_PORT: "` + db.Port + `"`
		case strings.HasPrefix(trimmed, "POSTGRES_DATABASE:"):
			lines[i] = `      POSTGRES_DATABASE: "` + db.Database + `"`
		case strings.HasPrefix(trimmed, "POSTGRES_USER:"):
			lines[i] = `      POSTGRES_USER: "` + db.Username + `"`
		case strings.HasPrefix(trimmed, "POSTGRES_PASSWORD:"):
			lines[i] = `      POSTGRES_PASSWORD: "` + db.Password + `"`
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func existingPorts(repoRoot string) (map[int]struct{}, map[int]struct{}, error) {
	httpPorts := map[int]struct{}{}
	grpcPorts := map[int]struct{}{}
	httpPattern := regexp.MustCompile(`^\s*port:\s*"(\d+)"\s*$`)
	grpcPattern := regexp.MustCompile(`^\s*grpc_port:\s*"(\d+)"\s*$`)

	matches, err := filepath.Glob(filepath.Join(repoRoot, "*-service", "config.dev.yaml"))
	if err != nil {
		return nil, nil, err
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if match := httpPattern.FindStringSubmatch(line); match != nil {
				port, _ := strconv.Atoi(match[1])
				httpPorts[port] = struct{}{}
			}
			if match := grpcPattern.FindStringSubmatch(line); match != nil {
				port, _ := strconv.Atoi(match[1])
				grpcPorts[port] = struct{}{}
			}
		}
	}

	return httpPorts, grpcPorts, nil
}

func nextPort(existing map[int]struct{}, base int) int {
	port := base + 1
	for {
		if _, ok := existing[port]; !ok {
			return port
		}
		port++
	}
}

func nextUnusedPort(start int, existing, local map[int]struct{}) int {
	port := start
	for {
		if _, ok := existing[port]; ok {
			port++
			continue
		}
		if _, ok := local[port]; ok {
			port++
			continue
		}
		return port
	}
}

func copyTemplate(repoRoot, target string) error {
	return filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		topLevel := strings.Split(relPath, string(filepath.Separator))[0]
		if d.Name() == ".DS_Store" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if slices.Contains(domain.ExcludedEntries, topLevel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(target, relPath)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func zipDirectory(sourceDir, rootName string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		archivePath := filepath.ToSlash(filepath.Join(rootName, relPath))
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = archivePath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		closeErr := file.Close()
		if err != nil {
			return err
		}

		return closeErr
	})
	if err != nil {
		_ = zipWriter.Close()
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = domain.ValidNamePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func findRepoRoot() (string, error) {
	if repoRoot := strings.TrimSpace(os.Getenv("REPO_ROOT")); repoRoot != "" {
		if _, err := os.Stat(filepath.Join(repoRoot, "scripts", "scaffold-service.sh")); err == nil {
			return repoRoot, nil
		}
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("could not determine source path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), ".."))
	if _, err := os.Stat(filepath.Join(repoRoot, "scripts", "scaffold-service.sh")); err != nil {
		return "", errors.New("could not locate repo root from generator-web")
	}

	return repoRoot, nil
}
