package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"
	"user-service/config"
	"user-service/helper"

	_DeliveryHTTP "user-service/data/delivery/http"
	_Repo "user-service/data/repository/postgres"
	_Usecase "user-service/data/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	configFile := flag.String("c", "config.yaml", "Config file")
	flag.Parse()
	config.ReadConfig(*configFile)
	initLogger()

	dsn := buildPostgresDSN()
	helper.RunMigrations(dsn)

	db := setupDatabase(dsn)
	app := setupFiber(db)

	/* Start server */
	go func() {
		port := viper.GetString(helper.SERVER_HTTP_PORT)
		if err := app.Listen(":" + port); err != nil {
			slog.Error("Failed to listen", "port", port)
		}
	}()

	/* Wait for shutdown */
	gracefulShutdown(app, db)
}

func initLogger() {
	env := "development"
	if viper.GetString(helper.SERVER_LOG_TYPE) == "json" {
		env = "production"
	}

	helper.InitLogger(env)
}

func buildPostgresDSN() string {
	query := url.Values{}
	query.Set("sslmode", viper.GetString(helper.POSTGRES_SSL_MODE))
	query.Set("TimeZone", viper.GetString(helper.POSTGES_TIMEZONE))

	dbAuth := url.User(viper.GetString(helper.POSTGES_USER))
	if password := viper.GetString(helper.POSTGES_PASSWORD); password != "" {
		dbAuth = url.UserPassword(viper.GetString(helper.POSTGES_USER), password)
	}

	host := fmt.Sprintf("%s:%s", viper.GetString(helper.POSTGES_HOST), strconv.Itoa(viper.GetInt(helper.POSTGES_PORT)))
	path := "/" + viper.GetString(helper.POSTGES_DATABASE)

	return (&url.URL{
		Scheme:   "postgresql",
		User:     dbAuth,
		Host:     host,
		Path:     path,
		RawQuery: query.Encode(),
	}).String()
}

func setupDatabase(dsn string) *gorm.DB {
	log.Printf(
		"Connecting to Postgres host=%s db=%s sslmode=%s (using pgx)",
		viper.GetString(helper.POSTGES_HOST),
		viper.GetString(helper.POSTGES_DATABASE),
		viper.GetString(helper.POSTGRES_SSL_MODE),
	)

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	loc, err := time.LoadLocation(viper.GetString(helper.POSTGES_TIMEZONE))
	if err != nil {
		log.Printf("timezone load failed, falling back to UTC: %v", err)
		loc = time.UTC
	}

	schemaName := viper.GetString(helper.POSTGRES_SCHEMA)
	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatal("parse pgx config:", err)
	}
	pgxCfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pgxCfg.StatementCacheCapacity = 0
	sqlDB := stdlib.OpenDB(*pgxCfg)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   schemaName + ".",
			SingularTable: false,
		},
		NowFunc: func() time.Time {
			return time.Now().In(loc)
		},
	})
	if err != nil {
		log.Fatal("error opening DB:", err)
	}

	sqlDB.SetMaxOpenConns(viper.GetInt(helper.POSTGRES_SET_MAX_OPEN_CONNS))
	sqlDB.SetMaxIdleConns(viper.GetInt(helper.POSTGRES_SET_MAX_IDLE_CONNS))
	sqlDB.SetConnMaxIdleTime(viper.GetDuration(helper.POSTGRES_SET_CONN_MAX_IDLE_TIME))
	sqlDB.SetConnMaxLifetime(viper.GetDuration(helper.POSTGRES_SET_CONN_MAX_LIFETIME))

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("failed to ping DB:", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	return db
}

func setupFiber(db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:       viper.GetBool(helper.SERVER_HTTP_PREFORK),
		StrictRouting: viper.GetBool(helper.SERVER_STATIC_ROUTING),
		CaseSensitive: viper.GetBool(helper.SERVER_CASE_SENSITIVE),
		BodyLimit:     viper.GetInt(helper.SERVER_HTTP_BODY_LIMIT),
	})

	/* Healthcheck */
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:     func(c *fiber.Ctx) bool { return true },
		LivenessEndpoint:  "/live",
		ReadinessProbe:    func(c *fiber.Ctx) bool { return true },
		ReadinessEndpoint: "/ready",
	}))
	app.Get("/healthy", func(c *fiber.Ctx) error {
		return c.SendString("i'm alive")
	})

	/* Logging middleware */
	logOpt := &slog.HandlerOptions{Level: &slog.LevelVar{}}
	textLog := slog.NewTextHandler(os.Stdout, logOpt)
	jsonLog := slog.NewJSONHandler(os.Stdout, logOpt)
	webLogger := slog.New(textLog)
	if viper.GetString(helper.SERVER_LOG_TYPE) == "json" {
		webLogger = slog.New(jsonLog)
		helper.Logger.Error(fmt.Sprintf("json log: %v", jsonLog))
	}
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		webLogger.Info(
			"http request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency", time.Since(start).String(),
		)

		return err
	})

	/* Recovery & CORS */
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			stack := string(debug.Stack())
			red := "\033[31m"
			reset := "\033[0m"

			fmt.Printf("%s[RECOVER] Panic: %v\n%s%s\n", red, e, stack, reset)
			helper.Logger.Error(fmt.Sprintf("panic recovered: %v", e))
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: viper.GetString(helper.SERVER_HTTP_ALOW_ORIGIN),
	}))

	/* Add version header */
	app.Use(func(c *fiber.Ctx) error {
		server := "hs/und"
		if version := os.Getenv("SERVER_VERSION"); version != "" {
			server = "hs/" + version
		}
		c.Set("Server", server)
		slog.Debug("respheader", "respheader", c.GetRespHeaders())
		return c.Next()
	})

	/* Routes */
	registerRoutes(app, db)

	return app
}

func registerRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World")
	})
	repo := _Repo.NewUserRepo(db)
	usecase := _Usecase.NewUserUsecase(repo)

	_DeliveryHTTP.RouterAPI(app, usecase)
}

func gracefulShutdown(app *fiber.App, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Gracefully shutting down...")
	_ = app.Shutdown()

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
