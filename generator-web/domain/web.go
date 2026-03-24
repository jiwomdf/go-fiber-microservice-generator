package domain

import (
	"regexp"
)

var (
	ValidNamePattern = regexp.MustCompile(`[^a-z0-9]+`)
	ExcludedEntries  = []string{".git", ".DS_Store", "generator-web"}
)

type GenerateRequest struct {
	ProjectName string            `json:"projectName"`
	AuthDB      DatabaseConfig    `json:"authDb"`
	UserDB      DatabaseConfig    `json:"userDb"`
	Services    []GenerateService `json:"services"`
}

type GenerateService struct {
	Name     string         `json:"name"`
	HTTPPort string         `json:"httpPort"`
	GRPCPort string         `json:"grpcPort"`
	DB       DatabaseConfig `json:"db"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}
