package logging

import (
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

	l = slog.LevelDebug

	log.SetOutput(w)
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: l,
		// AddSource: stage == StageDev,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
				}
			}
			return a
		},
	})

	slog.SetDefault(slog.New(h))
}
