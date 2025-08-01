package logger

import (
	"io"
	"os"
	"path/filepath"

	"user-svc/config"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// InitLogger initializes the logger with configuration
func InitLogger(cfg *config.Config) error {
	log = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	// Set log format
	switch cfg.Logging.Format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Set output
	switch cfg.Logging.Output {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		// Check if it's a file path
		if cfg.Logging.File.Enabled {
			// Ensure directory exists
			dir := filepath.Dir(cfg.Logging.File.Path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(cfg.Logging.File.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return err
			}

			// Use multi-writer to log to both file and stdout
			mw := io.MultiWriter(os.Stdout, file)
			log.SetOutput(mw)
		} else {
			log.SetOutput(os.Stdout)
		}
	}

	// Add fields for application context
	log.WithFields(logrus.Fields{
		"app_name":    cfg.App.Name,
		"app_version": cfg.App.Version,
		"environment": cfg.App.Environment,
	}).Info("Logger initialized")

	return nil
}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		// Fallback to default logger if not initialized
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.InfoLevel)
	}
	return log
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields adds multiple fields to the logger
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError adds an error field to the logger
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}
