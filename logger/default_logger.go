package logger

import (
	"io"
	"log"
)

// DefaultLogger default logger
type DefaultLogger struct {
	level LogLevel
}

// LogLevel logger level
type LogLevel int

// Log level
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Option config logger function
type Option func(l *DefaultLogger)

// NewLogger create a logger
func NewLogger(opts ...Option) *DefaultLogger {
	l := &DefaultLogger{}

	for _, o := range opts {
		o(l)
	}

	return l
}

// SetLevel set log level
func SetLevel(level LogLevel) Option {
	return func(l *DefaultLogger) {
		l.level = level
	}
}

// SetOutput set logger output
func SetOutput(w io.Writer) Option {
	return func(l *DefaultLogger) {
		log.SetOutput(w)
	}
}

// Debugf log debug level message
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		log.Printf("[DEBG] "+format, v...)
	}
}

// Infof log info level message
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warnf log warning level message
func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		log.Printf("[WARN] "+format, v...)
	}
}

// Errorf log error level message
func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		log.Printf("[ERRR] "+format, v...)
	}
}
