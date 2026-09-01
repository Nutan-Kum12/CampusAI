package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	Log  *slog.Logger
	once sync.Once
)

func Init(env string) {
	once.Do(func() {
		var handler slog.HandlerOptions
		handler.AddSource = true
		handler.Level = slog.LevelInfo
		if env == "production" {
			handler.AddSource = false
			handler.Level = slog.LevelInfo
		}
		logHandler := slog.NewJSONHandler(os.Stdout, &handler)
		Log = slog.New(logHandler)
	})
	if Log == nil {
		slog.Error("logger is not initialized")
	}
}

func WithRequestID(requestID string) *slog.Logger {
	return Log.With("request_id", requestID)
}

func WithUserID(userID string) *slog.Logger {
	return Log.With("user_id", userID)
}

func WithFields(args ...any) *slog.Logger {
	return Log.With(args...)
}

func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
