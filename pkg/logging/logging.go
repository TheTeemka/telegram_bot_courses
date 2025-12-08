package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	StageDev  = "dev"
	StageProd = "prod"
)

func SetSlog(stage string) {
	logFile, err := os.OpenFile("data/logging.json", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	var l slog.Level
	var w io.Writer

	switch stage {
	case StageDev:
		w = io.MultiWriter(os.Stdout, logFile)
		l = slog.LevelDebug
	case StageProd:
		w = logFile
		l = slog.LevelInfo
	default:
		panic("Unknown stage")
	}

	log.SetOutput(w)
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     l,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}

			switch v := a.Value.Any().(type) {
			case *slog.Source:
				if v != nil {
					short := filepath.Base(v.File)
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
				}
			case slog.Source:
				short := filepath.Base(v.File)
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
			default:
				// Fallback: shorten the string representation
				s := a.Value.String()
				if s != "" {
					a.Value = slog.StringValue(filepath.Base(s))
				}
			}
			return a
		},
	})

	slog.SetDefault(slog.New(h))
}
