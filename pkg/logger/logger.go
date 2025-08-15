package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

type logrusLogger struct {
	*logrus.Logger
}

func New(service string) Logger {
	logger := logrus.New()
	
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	// Set log level from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if lvl, err := logrus.ParseLevel(level); err == nil {
			logger.SetLevel(lvl)
		}
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.WithField("service", service)
	
	return &logrusLogger{logger}
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}