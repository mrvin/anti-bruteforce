package logger

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Conf struct {
	FilePath string `yaml:"filepath"`
	Level    string `yaml:"level"`
}

func Init(conf *Conf) (*os.File, error) {
	var level slog.Level

	if err := level.UnmarshalText([]byte(conf.Level)); err != nil {
		return nil, fmt.Errorf("getting level from text: %w", err)
	}

	var err error
	var logFile *os.File
	if conf.FilePath == "" {
		logFile = os.Stdout
	} else {
		logFile, err = os.OpenFile(conf.FilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed open log file: %w", err)
		}
	}

	replaceAttr := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			t := a.Value.Any().(time.Time) //nolint:forcetypeassert
			a.Value = slog.StringValue(t.Format(time.StampNano))
		}
		return a
	}

	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replaceAttr,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logFile, nil
}
