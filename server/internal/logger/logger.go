package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	instance *slog.Logger
	once     sync.Once
)

func Get() *slog.Logger {
	once.Do(func() {
		instance = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	})
	return instance
}
