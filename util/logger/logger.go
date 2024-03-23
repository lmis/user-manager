package logger

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
)

func NewLogger(logJson bool) *slog.Logger {
	if !logJson {
		return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource: true,
		}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.MessageKey {
				return slog.Attr{Key: "message", Value: a.Value}
			}
			if a.Key == slog.LevelKey {
				return slog.Attr{Key: "severity", Value: a.Value}
			}
			return a
		},
	}))

}
