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
	Services    []GenerateService `json:"services"`
}

type GenerateService struct {
	Name     string `json:"name"`
	HTTPPort string `json:"httpPort"`
	GRPCPort string `json:"grpcPort"`
}
