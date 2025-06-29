package logging

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	StageDev  = "dev"
	StageProd = "prod"
)

func SetSlog(stage string) {
	var l slog.Level
	switch stage {
	case StageDev:
		l = slog.LevelDebug
	case StageProd:
		l = slog.LevelInfo
	default:
		panic("Unknown stage")
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     l,
		AddSource: stage == StageDev,
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
