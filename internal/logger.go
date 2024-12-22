package internal

import (
	"log/slog"
	"os"
)

func SetLogger() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// as we don't exactly need timestamp it will be handler by systemd
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr { //nolint
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(h))
}
