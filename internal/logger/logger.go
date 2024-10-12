package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path"
)

var logger *slog.Logger

func init() {
	logPath := path.Join(os.Getenv("HOME"), ".backuper.log")
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	mw := io.MultiWriter(file, os.Stdout)
	logger = slog.New(slog.NewTextHandler(mw, nil))
}

func Info(msg string, v ...interface{}) {
	logger.Info(msg, v...)
}

func Error(err string, v ...interface{}) {
	logger.Error(err, v...)
}

func Fatal(msg string, v ...interface{}) {
	Error(msg, v...)
	os.Exit(1)
}
