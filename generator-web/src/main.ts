import JSZip from "jszip";
import { TEMPLATE_FILE_PATHS } from "./templateManifest";

type DatabaseConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password: string;
};

type ServiceInput = {
  name: string;
  httpPort: string;
  grpcPort: string;
  db: DatabaseConfig;
};

type ServiceSpec = {
  name: string;
  httpPort: number;
  grpcPort: number;
  db: DatabaseConfig;
};

type GenerationRequest = {
  projectName: string;
  traefikPort: number;
  authHttpPort: number;
  authGrpcPort: number;
  authDb: DatabaseConfig;
  services: ServiceSpec[];
};

const RAW_BASE =
  "https://raw.githubusercontent.com/jiwomdf/go-fiber-microservice-generator/main/";
const EXCLUDED_TOP_LEVEL = new Set([".git", ".DS_Store", "generator-web"]);
const servicesEl = document.getElementById("services") as HTMLDivElement;
const countEl = document.getElementById("service-count") as HTMLInputElement;
const formEl = document.getElementById("generator-form") as HTMLFormElement;
const statusEl = document.getElementById("status") as HTMLDivElement;
const buttonEl = document.getElementById("generate-button") as HTMLButtonElement;
const applyServicesButtonEl = document.getElementById(
  "apply-services-button",
) as HTMLButtonElement;

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function pascalCase(value: string): string {
  return value
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join("");
}

function replaceAllLiteral(value: string, from: string, to: string): string {
  return value.split(from).join(to);
}

function setStatus(message: string, isError = false): void {
  statusEl.textContent = message;
  statusEl.className = isError ? "status error" : "status";
}

function renderServiceInputs(): void {
  const count = Math.max(0, Math.min(20, Number(countEl.value || 0)));
  servicesEl.innerHTML = "";

  for (let i = 0; i < count; i += 1) {
    const card = document.createElement("div");
    card.className = "service-card extra-service-card";
    card.innerHTML = `
      <strong>Service ${i + 1}</strong>

      <div class="service-subsection non-db-section">
        <h4 class="subsection-title">Service Config</h4>
        <div class="grid">
          <label>
            Service Name
            <input name="service-name" placeholder="inventory" required>
          </label>
          <label>
            HTTP Port
            <input name="http-port" type="number" min="7000" max="7999" placeholder="auto">
          </label>
          <label>
            gRPC Port
            <input name="grpc-port" type="number" min="57000" max="59999" placeholder="auto">
          </label>
        </div>
      </div>

      <div class="service-subsection db-section">
        <h4 class="subsection-title">Database Config</h4>
        <div class="grid">
          <label>
            DB Host
            <input name="db-host" placeholder="localhost" required>
          </label>
          <label>
            DB Port
            <input name="db-port" placeholder="5432" required>
          </label>
          <label>
            DB Name
            <input name="db-name" placeholder="somedatabase" required>
          </label>
          <label>
            DB Username
            <input name="db-user" placeholder="postgres" required>
          </label>
          <label>
            DB Password
            <input name="db-password" placeholder="your-password">
          </label>
        </div>
      </div>
    `;
    servicesEl.appendChild(card);
  }
}

function applyServiceInputs(): void {
  const count = Number(countEl.value || 0);
  if (!Number.isInteger(count) || count < 0 || count > 20) {
    servicesEl.innerHTML = "";
    setStatus("Number Of New Services must be between 0 and 20.", true);
    return;
  }

  renderServiceInputs();
  if (count === 0) {
    setStatus("Auth-only project is ready to generate.");
    return;
  }

  setStatus(`${count} service input${count > 1 ? "s are" : " is"} ready to fill.`);
}

function readValue(id: string): string {
  const element = document.getElementById(id) as HTMLInputElement | null;
  return element?.value.trim() ?? "";
}

function readDatabase(prefix: string): DatabaseConfig {
  return {
    host: readValue(`${prefix}-db-host`),
    port: readValue(`${prefix}-db-port`),
    database: readValue(`${prefix}-db-name`),
    username: readValue(`${prefix}-db-user`),
    password: (document.getElementById(`${prefix}-db-password`) as HTMLInputElement)
      .value,
  };
}

function readServiceInputs(): ServiceInput[] {
  return Array.from(document.querySelectorAll(".service-card"))
    .filter((card) => card.querySelector('input[name="service-name"]'))
    .map((card) => ({
      name: (
        card.querySelector('input[name="service-name"]') as HTMLInputElement
      ).value.trim(),
      httpPort: (
        card.querySelector('input[name="http-port"]') as HTMLInputElement
      ).value.trim(),
      grpcPort: (
        card.querySelector('input[name="grpc-port"]') as HTMLInputElement
      ).value.trim(),
      db: {
        host: (
          card.querySelector('input[name="db-host"]') as HTMLInputElement
        ).value.trim(),
        port: (
          card.querySelector('input[name="db-port"]') as HTMLInputElement
        ).value.trim(),
        database: (
          card.querySelector('input[name="db-name"]') as HTMLInputElement
        ).value.trim(),
        username: (
          card.querySelector('input[name="db-user"]') as HTMLInputElement
        ).value.trim(),
        password: (
          card.querySelector('input[name="db-password"]') as HTMLInputElement
        ).value,
      },
    }))
    .filter((service) => service.name);
}

function normalizeDbConfig(label: string, db: DatabaseConfig): DatabaseConfig {
  const normalized = {
    host: db.host.trim(),
    port: db.port.trim(),
    database: db.database.trim(),
    username: db.username.trim(),
    password: db.password,
  };

  if (!normalized.host) {
    throw new Error(`${label} database host is required`);
  }
  if (!normalized.port) {
    throw new Error(`${label} database port is required`);
  }
  if (!/^\d+$/.test(normalized.port)) {
    throw new Error(`${label} database port must be numeric`);
  }
  if (!normalized.database) {
    throw new Error(`${label} database name is required`);
  }
  if (!normalized.username) {
    throw new Error(`${label} database username is required`);
  }

  return normalized;
}

function parsePort(raw: string, min: number, max: number, label: string): number {
  if (!raw.trim()) {
    return 0;
  }

  if (!/^\d+$/.test(raw.trim())) {
    throw new Error(`${label} must be a number`);
  }

  const port = Number(raw.trim());
  if (port < min || port > max) {
    throw new Error(`${label} must be between ${min} and ${max}`);
  }

  return port;
}

function nextUnusedPort(start: number, used: Set<number>): number {
  let port = start;
  while (used.has(port)) {
    port += 1;
  }
  return port;
}

function normalizeRequest(
  projectNameRaw: string,
  traefikPortRaw: string,
  authHttpPortRaw: string,
  authGrpcPortRaw: string,
  authDbRaw: DatabaseConfig,
  servicesRaw: ServiceInput[],
): GenerationRequest {
  const projectName = slugify(projectNameRaw);
  if (!projectName) {
    throw new Error("Project name is required.");
  }
  if (servicesRaw.length > 20) {
    throw new Error("Too many services requested.");
  }

  const authDb = normalizeDbConfig("auth-service", authDbRaw);
  const traefikPort = parsePort(traefikPortRaw, 1, 65535, "Traefik port") || 8000;
  const authHttpPort =
    parsePort(authHttpPortRaw, 7000, 7999, "auth-service http port") || 7704;
  const authGrpcPort =
    parsePort(authGrpcPortRaw, 57000, 59999, "auth-service grpc port") || 57704;

  if (traefikPort === authHttpPort) {
    throw new Error(`Traefik port already in use: ${traefikPort}`);
  }

  const usedHttp = new Set<number>([7704, 7705, authHttpPort, traefikPort]);
  const usedGrpc = new Set<number>([57704, 57705, authGrpcPort]);
  let nextHttp = 7704;
  let nextGrpc = 57704;

  const services = servicesRaw.map((raw) => {
    const serviceName = slugify(raw.name);
    if (!serviceName) {
      throw new Error("Service names must contain letters or numbers.");
    }
    if (serviceName === "auth") {
      throw new Error('"auth" is reserved because the base project already contains auth-service.');
    }

    const db = normalizeDbConfig(`${serviceName}-service`, raw.db);
    let httpPort = parsePort(raw.httpPort, 7000, 7999, `${serviceName} http port`);
    if (httpPort === 0) {
      nextHttp = nextUnusedPort(nextHttp + 1, usedHttp);
      httpPort = nextHttp;
    }
    if (usedHttp.has(httpPort)) {
      throw new Error(`HTTP port already in use: ${httpPort}`);
    }

    let grpcPort = parsePort(raw.grpcPort, 57000, 59999, `${serviceName} grpc port`);
    if (grpcPort === 0) {
      nextGrpc = nextUnusedPort(nextGrpc + 1, usedGrpc);
      grpcPort = nextGrpc;
    }
    if (usedGrpc.has(grpcPort)) {
      throw new Error(`gRPC port already in use: ${grpcPort}`);
    }

    usedHttp.add(httpPort);
    usedGrpc.add(grpcPort);
    return { name: serviceName, httpPort, grpcPort, db };
  });

  const uniqueNames = new Set(services.map((service) => service.name));
  if (uniqueNames.size !== services.length) {
    throw new Error("Duplicate service name in request.");
  }

  return {
    projectName,
    traefikPort,
    authHttpPort,
    authGrpcPort,
    authDb,
    services,
  };
}

async function runWithConcurrency<T>(
  items: readonly T[],
  limit: number,
  worker: (item: T, index: number) => Promise<void>,
): Promise<void> {
  let cursor = 0;
  const runners = Array.from({ length: Math.min(limit, items.length) }, async () => {
    while (cursor < items.length) {
      const current = cursor;
      cursor += 1;
      await worker(items[current], current);
    }
  });
  await Promise.all(runners);
}

function rawFileUrl(path: string): string {
  const encoded = path.split("/").map(encodeURIComponent).join("/");
  return `${RAW_BASE}${encoded}`;
}

async function fetchTemplateFiles(
  updateStatus: (message: string) => void,
): Promise<Map<string, string>> {
  const files = new Map<string, string>();
  let completed = 0;

  await runWithConcurrency(TEMPLATE_FILE_PATHS, 8, async (path) => {
    const response = await fetch(rawFileUrl(path), { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`Failed to fetch template file: ${path}`);
    }
    files.set(path, await response.text());
    completed += 1;
    updateStatus(`Fetching template files from GitHub (${completed}/${TEMPLATE_FILE_PATHS.length})...`);
  });

  return files;
}

function getFile(files: Map<string, string>, path: string): string {
  const content = files.get(path);
  if (content === undefined) {
    throw new Error(`Missing template file: ${path}`);
  }
  return content;
}

function setFile(files: Map<string, string>, path: string, content: string): void {
  files.set(path, content);
}

function removeServiceBlock(text: string, serviceName: string): string {
  const lines = text.split("\n");
  const output: string[] = [];
  let skipping = false;

  for (const line of lines) {
    if (new RegExp(`^  ${serviceName}:\\s*$`).test(line)) {
      skipping = true;
      continue;
    }
    if (skipping && /^  [a-z0-9-]+:\s*$/.test(line)) {
      skipping = false;
    }
    if (!skipping) {
      output.push(line);
    }
  }

  return output.join("\n").replace(/\n{3,}/g, "\n\n");
}

function patchScaffoldScript(files: Map<string, string>): void {
  const path = "scripts/scaffold-service.sh";
  let content = getFile(files, path);
  if (!content.includes('source_dir="$repo_root/templates/service-template"')) {
    content = replaceAllLiteral(
      content,
      'source_dir="$repo_root/user-service"',
      'source_dir="$repo_root/templates/service-template"\nif [[ ! -d "$source_dir" ]]; then\n  source_dir="$repo_root/user-service"\nfi',
    );
  }
  setFile(files, path, content);
}

function prepareServiceTemplate(files: Map<string, string>): void {
  const templatePrefix = "templates/service-template/";
  const hasTemplate = Array.from(files.keys()).some((path) =>
    path.startsWith(templatePrefix),
  );
  if (hasTemplate) {
    return;
  }

  for (const [path, content] of Array.from(files.entries())) {
    if (!path.startsWith("user-service/")) {
      continue;
    }
    files.set(path.replace(/^user-service\//, templatePrefix), content);
  }
}

function removeGeneratedUserService(files: Map<string, string>): void {
  for (const path of Array.from(files.keys())) {
    if (path.startsWith("user-service/")) {
      files.delete(path);
    }
  }

  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "user-service");
  compose = compose.replace(/\n\s+- user-service/g, "");
  setFile(files, "docker-compose.yml", compose);

  let routes = getFile(files, "traefik/dynamic/routes.yml");
  routes = routes.replace(
    "\n    user-api:\n      entryPoints:\n        - web\n      rule: PathPrefix(`/api/v1/user`)\n      middlewares:\n        - user-strip-v1\n        - user-addprefix\n        - protected-common\n      service: user-service\n",
    "",
  );
  routes = routes.replace(
    "\n    user-service:\n      loadBalancer:\n        servers:\n          - url: http://user-service:7705\n",
    "",
  );
  routes = routes.replace(
    "\n    user-strip-v1:\n      stripPrefix:\n        prefixes:\n          - /api/v1\n\n    user-addprefix:\n      addPrefix:\n        prefix: /api/user-service/v1\n",
    "",
  );
  setFile(files, "traefik/dynamic/routes.yml", routes);
}

function removeGeneratorWeb(files: Map<string, string>): void {
  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "generator-web");
  compose = compose.replace(/\n\s+- generator-web/g, "");
  setFile(files, "docker-compose.yml", compose);
}

function updateConfigSettings(
  files: Map<string, string>,
  path: string,
  httpPort: number,
  grpcPort: number,
  db: DatabaseConfig,
): void {
  const lines = getFile(files, path).split("\n");
  let inPostgres = false;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const trimmed = line.trim();

    if (trimmed.startsWith("port:") && line.startsWith("    ")) {
      lines[index] = `    port: "${httpPort}"`;
    }
    if (trimmed.startsWith("grpc_port:")) {
      lines[index] = `  grpc_port: "${grpcPort}"`;
    }
    if (trimmed === "postgres:") {
      inPostgres = true;
      continue;
    }
    if (inPostgres && !line.startsWith("  ") && trimmed !== "") {
      inPostgres = false;
    }
    if (!inPostgres) {
      continue;
    }

    if (trimmed.startsWith("host:")) {
      lines[index] = `  host: "${db.host}"`;
    } else if (trimmed.startsWith("port:")) {
      lines[index] = `  port: "${db.port}"`;
    } else if (trimmed.startsWith("database:")) {
      lines[index] = `  database: "${db.database}"`;
    } else if (trimmed.startsWith("user:")) {
      lines[index] = `  user: "${db.username}"`;
    } else if (trimmed.startsWith("password:")) {
      lines[index] = `  password: "${db.password}"`;
    }
  }

  setFile(files, path, lines.join("\n"));
}

function updateComposeServiceEnv(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  db: DatabaseConfig,
): void {
  const path = "docker-compose.yml";
  const lines = getFile(files, path).split("\n");
  let currentService = "";
  let inEnvironment = false;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const match = line.match(/^  ([a-z0-9-]+):\s*$/);
    if (match) {
      currentService = match[1];
      inEnvironment = false;
      continue;
    }

    if (currentService !== serviceName) {
      continue;
    }

    const trimmed = line.trim();
    if (line.startsWith("    environment:")) {
      inEnvironment = true;
      continue;
    }
    if (inEnvironment && line.startsWith("    ") && !line.startsWith("      ")) {
      inEnvironment = false;
    }
    if (!inEnvironment) {
      if (line.startsWith("      - \"")) {
        lines[index] = `      - "${httpPort}:${httpPort}"`;
        break;
      }
      continue;
    }

    if (trimmed.startsWith("POSTGRES_HOST:")) {
      lines[index] = `      POSTGRES_HOST: "${db.host}"`;
    } else if (trimmed.startsWith("POSTGRES_PORT:")) {
      lines[index] = `      POSTGRES_PORT: "${db.port}"`;
    } else if (trimmed.startsWith("POSTGRES_DATABASE:")) {
      lines[index] = `      POSTGRES_DATABASE: "${db.database}"`;
    } else if (trimmed.startsWith("POSTGRES_USER:")) {
      lines[index] = `      POSTGRES_USER: "${db.username}"`;
    } else if (trimmed.startsWith("POSTGRES_PASSWORD:")) {
      lines[index] = `      POSTGRES_PASSWORD: "${db.password}"`;
    } else if (trimmed.startsWith("SERVER_HTTP_PORT:")) {
      lines[index] = `      SERVER_HTTP_PORT: "${httpPort}"`;
    }
  }

  setFile(files, path, lines.join("\n"));
}

function updateServiceConfig(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  grpcPort: number,
  db: DatabaseConfig,
): void {
  updateConfigSettings(files, `${serviceName}/config.dev.yaml`, httpPort, grpcPort, db);
  updateConfigSettings(files, `${serviceName}/config.local.yaml`, httpPort, grpcPort, db);
  updateComposeServiceEnv(files, serviceName, httpPort, db);
}

function updateTraefikBasePorts(files: Map<string, string>, authHttpPort: number): void {
  let routes = getFile(files, "traefik/dynamic/routes.yml");
  routes = replaceAllLiteral(routes, "http://auth-service:7704", `http://auth-service:${authHttpPort}`);
  setFile(files, "traefik/dynamic/routes.yml", routes);

  let middlewares = getFile(files, "traefik/dynamic/middlewares.yml");
  middlewares = replaceAllLiteral(
    middlewares,
    "http://auth-service:7704",
    `http://auth-service:${authHttpPort}`,
  );
  setFile(files, "traefik/dynamic/middlewares.yml", middlewares);
}

function updateTraefikPort(files: Map<string, string>, traefikPort: number): void {
  let compose = getFile(files, "docker-compose.yml");
  compose = replaceAllLiteral(compose, '- "8000:8000"', `- "${traefikPort}:${traefikPort}"`);
  setFile(files, "docker-compose.yml", compose);

  let traefikConfig = getFile(files, "traefik/traefik.yml");
  traefikConfig = replaceAllLiteral(traefikConfig, 'address: ":8000"', `address: ":${traefikPort}"`);
  setFile(files, "traefik/traefik.yml", traefikConfig);
}

function appendComposeService(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  db: DatabaseConfig,
): void {
  const block = `
  ${serviceName}:
    build:
      context: ./${serviceName}
    container_name: ${serviceName}
    environment:
      POSTGRES_HOST: "${db.host}"
      POSTGRES_PORT: "${db.port}"
      POSTGRES_DATABASE: "${db.database}"
      POSTGRES_USER: "${db.username}"
      POSTGRES_PASSWORD: "${db.password}"
      POSTGRES_SCHEMA: public
      SERVER_HTTP_PORT: "${httpPort}"
    ports:
      - "${httpPort}:${httpPort}"
`;
  const compose = getFile(files, "docker-compose.yml").replace("\n  traefik:\n", `${block}\n  traefik:\n`);
  setFile(files, "docker-compose.yml", compose);
}

function appendTraefikRoute(
  files: Map<string, string>,
  entity: string,
  serviceName: string,
  httpPort: number,
): void {
  let routes = getFile(files, "traefik/dynamic/routes.yml");
  const routerBlock = `
    ${entity}-api:
      entryPoints:
        - web
      rule: PathPrefix(\`/api/v1/${entity}\`)
      middlewares:
        - ${entity}-strip-v1
        - ${entity}-addprefix
        - protected-common
      service: ${serviceName}
`;
  const serviceBlock = `
    ${serviceName}:
      loadBalancer:
        servers:
          - url: http://${serviceName}:${httpPort}
`;
  const middlewareBlock = `
    ${entity}-strip-v1:
      stripPrefix:
        prefixes:
          - /api/v1

    ${entity}-addprefix:
      addPrefix:
        prefix: /api/${serviceName}/v1
`;
  routes = routes.replace("\n  services:\n", `${routerBlock}\n  services:\n`);
  routes = routes.replace("\n  middlewares:\n", `${serviceBlock}\n  middlewares:\n`);
  routes = `${routes.trimEnd()}\n${middlewareBlock}`;
  setFile(files, "traefik/dynamic/routes.yml", routes);
}

function scaffoldService(
  files: Map<string, string>,
  spec: ServiceSpec,
): void {
  const entity = spec.name;
  const serviceName = `${entity}-service`;
  const entityPascal = pascalCase(entity);
  const entityPlural = `${entity}s`;
  const entityPluralPascal = `${entityPascal}s`;
  const entityUpper = entity.toUpperCase();
  const sourcePrefix = "templates/service-template/";
  const targetPrefix = `${serviceName}/`;

  const pathRenames: Record<string, string> = {
    "domain/user.go": `domain/${entity}.go`,
    "data/usecase/user.go": `data/usecase/${entity}.go`,
    "data/repository/postgres/user.go": `data/repository/postgres/${entity}.go`,
    "data/delivery/http/handler/user.go": `data/delivery/http/handler/${entity}.go`,
    "data/delivery/grpc/proto/user.proto": `data/delivery/grpc/proto/${entity}.proto`,
  };

  const copiedPaths: string[] = [];
  for (const [path, content] of Array.from(files.entries())) {
    if (!path.startsWith(sourcePrefix)) {
      continue;
    }
    if (path.includes("/.vscode/")) {
      continue;
    }

    const relativePath = path.slice(sourcePrefix.length);
    const renamedRelativePath = pathRenames[relativePath] ?? relativePath;
    const targetPath = `${targetPrefix}${renamedRelativePath}`;
    files.set(targetPath, content);
    copiedPaths.push(targetPath);
  }

  const specificReplacements: Array<[string, string]> = [
    ["user-service", serviceName],
    ["CreateUserRequest", `Create${entityPascal}Request`],
    ["UpdateUserRequest", `Update${entityPascal}Request`],
    ["GetUserByIdRequest", `Get${entityPascal}ByIdRequest`],
    ["DeleteUserRequest", `Delete${entityPascal}Request`],
    ["DeleteUserResponse", `Delete${entityPascal}Response`],
    ["UserListResponse", `${entityPascal}ListResponse`],
    ["UserResponse", `${entityPascal}Response`],
    ["UserHandler", `${entityPascal}Handler`],
    ["UserUsecase", `${entityPascal}Usecase`],
    ["UserRepository", `${entityPascal}Repository`],
    ["UserRepo", `${entityPascal}Repo`],
    ["UserService", `${entityPascal}Service`],
    ["NewUserHandler", `New${entityPascal}Handler`],
    ["NewUserUsecase", `New${entityPascal}Usecase`],
    ["NewUserRepo", `New${entityPascal}Repo`],
    ["GetAllUsers", `GetAll${entityPluralPascal}`],
    ["GetUserById", `Get${entityPascal}ById`],
    ["CreateUser", `Create${entityPascal}`],
    ["UpdateUser", `Update${entityPascal}`],
    ["DeleteUser", `Delete${entityPascal}`],
    ["/user", `/${entity}`],
    ['"users"', `"${entityPlural}"`],
    [" public.users ", ` public.${entityPlural} `],
    ["user endpoints", `${entity} endpoints`],
    ["Create user", `Create ${entity}`],
    ["Get all users", `Get all ${entityPlural}`],
    ["Get user by ID", `Get ${entity} by ID`],
    ["Update user by ID", `Update ${entity} by ID`],
    ["Delete user by ID", `Delete ${entity} by ID`],
    ["User Routes", `${entityPascal} Routes`],
  ];

  const entityFiles = new Set([
    `${targetPrefix}domain/errors.go`,
    `${targetPrefix}domain/${entity}.go`,
    `${targetPrefix}data/usecase/${entity}.go`,
    `${targetPrefix}data/repository/postgres/${entity}.go`,
    `${targetPrefix}data/delivery/http/handler/${entity}.go`,
    `${targetPrefix}data/delivery/grpc/proto/${entity}.proto`,
    `${targetPrefix}openapi.yaml`,
  ]);

  for (const path of copiedPaths) {
    let content = getFile(files, path);
    for (const [from, to] of specificReplacements) {
      content = replaceAllLiteral(content, from, to);
    }
    if (entityFiles.has(path)) {
      content = replaceAllLiteral(content, "Users", entityPluralPascal);
      content = replaceAllLiteral(content, "User", entityPascal);
    }
    if (path === `${targetPrefix}config/struct.go`) {
      content = replaceAllLiteral(
        content,
        'ClientID:          "user",',
        `ClientID:          "${entity}",`,
      );
    }
    if (path === `${targetPrefix}Dockerfile`) {
      content = replaceAllLiteral(content, "EXPOSE 7705", `EXPOSE ${spec.httpPort}`);
    }
    if (path === `${targetPrefix}openapi.yaml`) {
      content = replaceAllLiteral(
        content,
        "url: http://localhost:7704/api/",
        `url: http://localhost:${spec.httpPort}/api/`,
      );
    }
    if (path === `${targetPrefix}domain/errors.go`) {
      content = replaceAllLiteral(
        content,
        'const StatusCodePrefix = "USER"',
        `const StatusCodePrefix = "${entityUpper}"`,
      );
    }
    setFile(files, path, content);
  }

  appendComposeService(files, serviceName, spec.httpPort, spec.db);
  appendTraefikRoute(files, entity, serviceName, spec.httpPort);
  updateServiceConfig(files, serviceName, spec.httpPort, spec.grpcPort, spec.db);
}

async function buildProjectZip(
  request: GenerationRequest,
  updateStatus: (message: string) => void,
): Promise<Blob> {
  const files = await fetchTemplateFiles(updateStatus);

  patchScaffoldScript(files);
  prepareServiceTemplate(files);
  removeGeneratedUserService(files);
  removeGeneratorWeb(files);
  updateServiceConfig(
    files,
    "auth-service",
    request.authHttpPort,
    request.authGrpcPort,
    request.authDb,
  );
  updateTraefikBasePorts(files, request.authHttpPort);
  updateTraefikPort(files, request.traefikPort);

  for (const service of request.services) {
    updateStatus(`Scaffolding ${service.name}-service...`);
    scaffoldService(files, service);
  }

  updateStatus("Building zip archive...");
  const zip = new JSZip();
  for (const [path, content] of Array.from(files.entries()).sort(([a], [b]) =>
    a.localeCompare(b),
  )) {
    const topLevel = path.split("/")[0];
    if (EXCLUDED_TOP_LEVEL.has(topLevel)) {
      continue;
    }

    zip.file(`${request.projectName}/${path}`, content, {
      unixPermissions: path.endsWith(".sh") ? 0o755 : 0o644,
    });
  }

  return zip.generateAsync({ type: "blob", compression: "DEFLATE" });
}

function downloadBlob(filename: string, blob: Blob): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

async function generateProject(event: SubmitEvent): Promise<void> {
  event.preventDefault();
  setStatus("");
  buttonEl.disabled = true;

  try {
    const request = normalizeRequest(
      readValue("project-name"),
      readValue("traefik-port"),
      readValue("auth-http-port"),
      readValue("auth-grpc-port"),
      readDatabase("auth"),
      readServiceInputs(),
    );

    setStatus("Preparing template...");
    const zipBlob = await buildProjectZip(request, setStatus);
    downloadBlob(`${request.projectName}.zip`, zipBlob);
    setStatus("Project generated. Download should start automatically.");
  } catch (error) {
    const message = error instanceof Error ? error.message : "Generation failed.";
    setStatus(message, true);
  } finally {
    buttonEl.disabled = false;
  }
}

applyServicesButtonEl.addEventListener("click", applyServiceInputs);
formEl.addEventListener("submit", (event) => {
  void generateProject(event as SubmitEvent);
});
servicesEl.innerHTML = "";
