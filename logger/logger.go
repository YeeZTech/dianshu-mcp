// Package logger provides structured logging for dianshu-mcp.
// It wraps logrus with a consistent key-value interface.
//
// Author: zhyyao
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)
}

// SetLevel sets the minimum log level.
func SetLevel(level logrus.Level) { log.SetLevel(level) }

// Info logs an informational message.
func Info(msg string, keysAndValues ...interface{}) {
	log.WithFields(toFields(keysAndValues...)).Info(msg)
}

// Warn logs a warning message.
func Warn(msg string, keysAndValues ...interface{}) {
	log.WithFields(toFields(keysAndValues...)).Warn(msg)
}

// Error logs an error message.
func Error(msg string, keysAndValues ...interface{}) {
	log.WithFields(toFields(keysAndValues...)).Error(msg)
}

// Debug logs a debug message.
func Debug(msg string, keysAndValues ...interface{}) {
	log.WithFields(toFields(keysAndValues...)).Debug(msg)
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, keysAndValues ...interface{}) {
	log.WithFields(toFields(keysAndValues...)).Fatal(msg)
}

// With returns a logger with preset fields.
func With(keysAndValues ...interface{}) *logrus.Entry {
	return log.WithFields(toFields(keysAndValues...))
}

// toFields converts key-value pairs to logrus.Fields.
func toFields(keysAndValues ...interface{}) logrus.Fields {
	f := logrus.Fields{}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		f[key] = keysAndValues[i+1]
	}
	return f
}
