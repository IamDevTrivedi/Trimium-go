package logger

import (
	"log/slog"
	"os"
)

var instance *slog.Logger

func Get() *slog.Logger {
	if instance == nil {
		instance = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return instance
}
