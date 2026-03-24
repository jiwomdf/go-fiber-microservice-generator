package helper

const (

	// Database
	DB_HOST     = "db.host"
	DB_PORT     = "db.port"
	DB_USER     = "db.user"
	DB_PASSWORD = "db.password"
	DB_NAME     = "db.name"

	// Server
	SERVER_HTTP_PORT        = "server.http.port"
	SERVER_HTTP_PREFORK     = "server.http.prefork"
	SERVER_HTTP_STRICT      = "server.http.strict_routing"
	SERVER_HTTP_CASE        = "server.http.case_sensitive"
	SERVER_HTTP_BODY        = "server.http.body_limit"
	SERVER_LOG_TYPE         = "server.log_type"
	SERVER_LOG_LEVEL        = "server.log_level"
	SERVER_STATIC_ROUTING   = "server.http.strict_routing"
	SERVER_CASE_SENSITIVE   = "server.http.case_sensitive"
	SERVER_HTTP_BODY_LIMIT  = "server.http.body_limit"
	SERVER_HTTP_ALOW_ORIGIN = "server.http.allow_origins"

	// POSTGRES
	POSTGES_HOST                    = "postgres.host"
	POSTGES_USER                    = "postgres.user"
	POSTGES_PASSWORD                = "postgres.password"
	POSTGES_DATABASE                = "postgres.database"
	POSTGES_PORT                    = "postgres.port"
	POSTGES_TIMEZONE                = "postgres.timezone"
	POSTGRES_SCHEMA                 = "postgres.schema"
	POSTGRES_PATH_MIGRATE           = "postgres.path_migrate"
	POSTGRES_SET_MAX_OPEN_CONNS     = "postgres.set_max_open_conns"
	POSTGRES_SET_MAX_IDLE_CONNS     = "postgres.set_max_idle_conns"
	POSTGRES_SET_CONN_MAX_IDLE_TIME = "postgres.set_conn_max_idle_time"
	POSTGRES_SET_CONN_MAX_LIFETIME  = "postgres.set_conn_max_lifetime"
	POSTGRES_SSL_MODE               = "postgres.sslmode"

	// JWT
	JWT_EXPARATION      = "jwt.expiration"
	JWT_SIGNING_METHOD  = "jwt.signing_method"
	JWT_SIGNATURE_KEY   = "jwt.signature_key"
	JWT_EXPIRATION_TEMP = "jwt.expiration_temp"
	JWT_EXPIRATION_MCH  = "jwt.expiration_machine"
)
