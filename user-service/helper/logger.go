package helper

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

var Logger *slog.Logger

func InitLogger(env string) {
	var handler slog.Handler

	var lvl slog.Level
	switch viper.GetString(SERVER_LOG_LEVEL) {
	case "debug":
		lvl = slog.LevelDebug
	case "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: true,
	})

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}
