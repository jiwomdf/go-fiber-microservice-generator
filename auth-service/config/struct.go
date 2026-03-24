package config

import (
	"time"
)

// Config is main config structure
type Config struct {
	Server            Server     `yaml:"server"`
	Postgres          Postgres   `yaml:"postgres"`
	SMTP              SMTP       `yaml:"smtp"`
	DefaultLimitQuery int64      `yaml:"default_limit_query"`
	ClientID          string     `yaml:"client_id"`
	GRPCClient        GRPCClient `yaml:"grpc_client"`
	Tracer            Tracer     `yaml:"tracer"`
}

// Server is server related config
type Server struct {
	// BasePath is router base path
	BasePath string `yaml:"base_path"`

	// LogType is log type, available value: text, json
	LogType string `yaml:"log_type"`

	// LogLevel is log level, available value: error, warning, info, debug
	LogLevel string `yaml:"log_level"`

	// HTTP is HTTP server config
	HTTP HTTPServer `yaml:"http"`

	// GRPC Port is port for grpc server
	GRPCPort string `yaml:"grpc_port"`
}

// HTTPServer is HTTP server related config
type HTTPServer struct {
	// Port is the local machine TCP Port to bind the HTTP Server to
	Port string `yaml:"port"`

	// Prefork will spawn multiple Go processes listening on the same port
	Prefork bool `yaml:"prefork"`

	// StrictRouting
	// When enabled, the router treats /foo and /foo/ as different.
	// Otherwise, the router treats /foo and /foo/ as the same.
	StrictRouting bool `yaml:"strict_routing"`

	// CaseSensitive
	// When enabled, /Foo and /foo are different routes.
	// When disabled, /Foo and /foo are treated the same.
	CaseSensitive bool `yaml:"case_sensitive"`

	// BodyLimit
	// Sets the maximum allowed size for a request body, if the size exceeds
	// the configured limit, it sends 413 - Request Entity Too Large response.
	BodyLimit int `yaml:"body_limit"`

	// Concurrency maximum number of concurrent connections
	Concurrency int `yaml:"concurrency"`

	// Timeout is HTTP server timeout
	Timeout Timeout `yaml:"timeout"`

	// AllowOrigins
	// Put a list of origins that are allowed to access the resource,
	// separated by comma
	AllowOrigins string `yaml:"allow_origins"`

	// AllowMethods
	AllowMethods string `yaml:"allow_methods"`

	// AllowHeaders
	AllowHeaders string `yaml:"allow_headers"`

	// ExposeHeaders
	ExposeHeaders string `yaml:"expose_headers"`

	FiberMicroServicePath string `yaml:"fiber_micro_service_path"`

	BaseURL string `yaml:"base_url"`

	// CacheStaticTTL defines how long (in seconds) static files should be cached.
	// This value is used to set the Cache-Control header.
	CacheStaticTTL int `yaml:"cache_static_ttl"`
}

// Timeout is server timeout related config
type Timeout struct {
	// Read is the amount of time to wait until an HTTP server
	// read operation is cancelled
	Read time.Duration `yaml:"read"`

	// Write is the amount of time to wait until an HTTP server
	// write opperation is cancelled
	Write time.Duration `yaml:"write"`

	// Read is the amount of time to wait
	// until an IDLE HTTP session is closed
	Idle time.Duration `yaml:"idle"`
}

// Postgres is PostgreSQL database related config
type Postgres struct {
	// Host is the PostgreSQL IP Address to connect to
	Host string `yaml:"host,omitempty"`

	// Port is the PostgreSQL Port to connect to
	Port string `yaml:"port,omitempty"`

	// Database is PostgreSQL database name
	Database string `yaml:"database"`

	// User is PostgreSQL username
	User string `yaml:"user"`

	// Password is PostgreSQL password
	Password string `yaml:"password"`

	// PathMigrate is directory for migration file
	PathMigrate string `yaml:"path_migrate"`

	// Timezone is PostgreSQL timezone
	Timezone string `yaml:"timezone"`

	// SSLMode is PostgreSQL sslmode
	SSLMode string `yaml:"sslmode"`

	// SetMaxOpenConns is maximum number of open connections to the database
	SetMaxOpenConns int `yaml:"set_max_open_conns"`

	// SetMaxIdleConns is maximum number of connections in the idle connection
	// pool
	SetMaxIdleConns int `yaml:"set_max_idle_conns"`

	// SetConnMaxIdleTime is maximum amount of time a connection may be idle
	SetConnMaxIdleTime time.Duration `yaml:"set_conn_max_idle_time"`

	// SetConnMaxLifetime is maximum amount of time a connection may be
	// reused
	SetConnMaxLifetime time.Duration `yaml:"set_conn_max_lifetime"`
}

type HostPort struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type SMTP struct {
	Mail        string `yaml:"mail"`
	Port        int64  `yaml:"port"`
	StartTls    int64  `yaml:"start_tls"`
	TslOrSsl    int64  `yaml:"tsl_or_ssl"`
	EmailSender string `yaml:"email_sender"`
	Password    string `yaml:"password"`
}

type GRPCClient struct {
	AdministrativeDivisionService HostPort      `yaml:"administrative_division_service"`
	ConfigurationService          HostPort      `yaml:"configuration_service"`
	IdentityService               HostPort      `yaml:"identity_service"`
	BillingService                HostPort      `yaml:"billing_service"`
	Init                          int           `yaml:"init"`
	Capacity                      int           `yaml:"capacity"`
	IdleDuration                  time.Duration `yaml:"idle_duration"`
	MaxLifeDuration               time.Duration `yaml:"max_life_duration"`
}

type Tracer struct {
	// ServiceName is the name of the service for tracing
	ServiceName string `yaml:"service_name"`

	// CollectorEndpointGRPC is the gRPC endpoint for the OpenTelemetry collector (e.g., Grafana Tempo)
	CollectorEndpointGRPC string `yaml:"collector_endpoint_grpc"`

	// EnableExporter enables the OpenTelemetry exporter
	// if true, the OpenTelemetry exporter to Grafana Tempo will be enabled
	EnableExporter bool `yaml:"enable_exporter"`
}

type MinIO struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
}

// Default config
var DefaultConfig = &Config{
	Server: Server{
		BasePath: "/api",
		LogType:  "text",
		LogLevel: "debug",
		HTTP: HTTPServer{
			Port:          "7704",
			Prefork:       false,
			StrictRouting: false,
			CaseSensitive: false,
			BodyLimit:     104 * 1024 * 1024,
			Concurrency:   256 * 1024,
			Timeout: Timeout{
				Read:  5,
				Write: 10,
				Idle:  120,
			},
			AllowOrigins:   "*",
			AllowMethods:   "GET, POST, PUT, DELETE, PATCH, OPTIONS",
			AllowHeaders:   "Origin, Content-Type, Accept, Authorization",
			CacheStaticTTL: 24 * 60 * 60,
		},
		GRPCPort: "57704",
	},
	Postgres: Postgres{
		Host:               "localhost",
		Port:               "5432",
		Database:           "somedb",
		User:               "someuser",
		Password:           "kmzway87aa",
		PathMigrate:        "file://migration",
		Timezone:           "UTC",
		SSLMode:            "disable",
		SetMaxOpenConns:    0,
		SetMaxIdleConns:    2,
		SetConnMaxIdleTime: 60 * time.Second,
		SetConnMaxLifetime: 5 * time.Minute,
	},
	SMTP: SMTP{
		Mail:        "smtp.gmail.com",
		Port:        25,
		StartTls:    587,
		TslOrSsl:    465,
		EmailSender: "",
		Password:    "",
	},
	DefaultLimitQuery: 100,
	ClientID:          "auth",
	GRPCClient: GRPCClient{
		Init:            5,
		Capacity:        50,
		IdleDuration:    60,
		MaxLifeDuration: 60,
	},
	Tracer: Tracer{
		ServiceName:           "auth-service",
		CollectorEndpointGRPC: "localhost:4317",
		EnableExporter:        false,
	},
}
