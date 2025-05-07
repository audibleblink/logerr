package logerr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// LogLevel represents the level of logging verbosity
type LogLevel int

// Log level constants
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String labels for each log level
var labels = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
}

// Color configurations for each log level
var labelColors = map[LogLevel]*color.Color{
	LogLevelDebug: color.New(color.FgCyan),
	LogLevelInfo:  color.New(color.FgGreen),
	LogLevelWarn:  color.New(color.FgYellow),
	LogLevelError: color.New(color.FgRed),
	LogLevelFatal: color.New(color.BgRed),
}

// Global default logger instance
var G = DefaultLogger()

// Logger provides structured logging capabilities
type Logger struct {
	// Level dictates the minimum LogLevel that will be output
	Level LogLevel

	// Output destination for log messages
	Output io.Writer

	// Exclusive dictates whether _only_ the configured loglevel messages are shown
	// Defaults to false, which prints everything at the configured LogLevel or higher
	Exclusive bool

	// LogWrappedErrors, when enabled, will print the error text,
	// according to level and context, before returning the error
	LogWrappedErrors bool

	// Additional prefix text to add context to log messages
	context []string

	// NoColor disables colored output when true
	NoColor bool
}

// DefaultLogger creates a new logger with default settings
func DefaultLogger() *Logger {
	logger := &Logger{
		Level:   LogLevelError,
		Output:  os.Stderr,
		NoColor: true,
	}
	color.NoColor = logger.NoColor
	return logger.SetContext("")
}

// SetAsGlobal sets this logger as the global default logger
func (l Logger) SetAsGlobal() {
	G = &l
}

// Context returns the current context string
func (l Logger) Context() string {
	return strings.Join(l.context, " | ")
}

// ClearContext removes all context from the logger
func (l *Logger) ClearContext() {
	l.context = make([]string, 0)
}

// SetContext sets a single context value, replacing any existing context
func (l Logger) SetContext(s string) *Logger {
	l.context = []string{s}
	return &l
}

// Add returns a copy of the logger with additional context
// Useful for loggers that can be used in a specific scope
func (l *Logger) Add(context string) Logger {
	dup := *l
	dup.context = append(dup.context, context)
	return dup
}

// EnableColors enables colored output
func (l *Logger) EnableColors() *Logger {
	l.NoColor = false
	color.NoColor = l.NoColor
	return l
}

// DisableColors disables colored output
func (l *Logger) DisableColors() *Logger {
	l.NoColor = true
	color.NoColor = l.NoColor
	return l
}

// Wrap wraps an error with the current context
// If a string is provided, it will be converted to an error
func (l Logger) Wrap(val any) error {
	var err error
	switch v := val.(type) {
	case error:
		err = v
	case string:
		err = errors.New(v)
	default:
		err = fmt.Errorf("%v", v)
	}
	
	if l.LogWrappedErrors {
		l.Error(err)
	}
	return fmt.Errorf("%s | %w", l.Context(), err)
}

// shouldLog determines if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	return (l.Level == level && l.Exclusive) || (l.Level <= level && !l.Exclusive)
}

// formatLogMessage creates a formatted log message with the level and context
func (l *Logger) formatLogMessage(level LogLevel, msg string) string {
	prefix := formatLabel(level, l.NoColor)
	ctx := l.Context()

	if ctx == "" {
		return fmt.Sprintf("%s | %s", prefix, msg)
	}

	return fmt.Sprintf("%s %s | %s", prefix, ctx, msg)
}

// messageToString converts a message (string or error) to string
func messageToString(message any) string {
	switch msg := message.(type) {
	case string:
		return msg
	case error:
		return msg.Error()
	default:
		return fmt.Sprintf("%v", msg)
	}
}

// log outputs a message if it should be logged based on level
// message can be a string or an error
func (l *Logger) log(level LogLevel, message any) {
	if l.shouldLog(level) {
		msgStr := messageToString(message)
		formatted := l.formatLogMessage(level, msgStr)
		fmt.Fprintln(l.Output, formatted)
	}
}

// logf outputs a formatted message if it should be logged based on level
func (l *Logger) logf(level LogLevel, format string, args ...any) {
	if l.shouldLog(level) {
		l.log(level, fmt.Sprintf(format, args...))
	}
}

// Debug logs a message at DEBUG level
// message can be a string or an error
func (l Logger) Debug(message any) {
	l.log(LogLevelDebug, message)
}

// Debugf logs a formatted message at DEBUG level
func (l Logger) Debugf(format string, args ...any) {
	l.logf(LogLevelDebug, format, args...)
}

// Info logs a message at INFO level
// message can be a string or an error
func (l Logger) Info(message any) {
	l.log(LogLevelInfo, message)
}

// Infof logs a formatted message at INFO level
func (l Logger) Infof(format string, args ...any) {
	l.logf(LogLevelInfo, format, args...)
}

// Warn logs a message at WARN level
// message can be a string or an error
func (l Logger) Warn(message any) {
	l.log(LogLevelWarn, message)
}

// Warnf logs a formatted message at WARN level
func (l Logger) Warnf(format string, args ...any) {
	l.logf(LogLevelWarn, format, args...)
}

// Error logs a message at ERROR level
// message can be a string or an error
func (l Logger) Error(message any) {
	l.log(LogLevelError, message)
}

// Errorf logs a formatted message at ERROR level
func (l Logger) Errorf(format string, args ...any) {
	l.logf(LogLevelError, format, args...)
}

// Fatal logs a message at FATAL level and exits the program
// message can be a string or an error
func (l Logger) Fatal(message any) {
	l.log(LogLevelFatal, message)
	os.Exit(1)
}

// Fatalf logs a formatted message at FATAL level and exits the program
func (l Logger) Fatalf(format string, args ...any) {
	l.logf(LogLevelFatal, format, args...)
	os.Exit(1)
}

// formatLabel returns a formatted label string for the given level
func formatLabel(level LogLevel, noColor bool) string {
	labelText := labels[level]

	if noColor {
		return fmt.Sprintf("[%s]", labelText)
	}

	return labelColors[level].Sprintf("[%s]", labelText)
}

// Global convenience functions that use the default logger

// Debug logs a message at DEBUG level using the global logger
// message can be a string or an error
func Debug(message any) { G.Debug(message) }

// Debugf logs a formatted message at DEBUG level using the global logger
func Debugf(format string, vals ...any) { G.Debugf(format, vals...) }

// Info logs a message at INFO level using the global logger
// message can be a string or an error
func Info(message any) { G.Info(message) }

// Infof logs a formatted message at INFO level using the global logger
func Infof(format string, vals ...any) { G.Infof(format, vals...) }

// Warn logs a message at WARN level using the global logger
// message can be a string or an error
func Warn(message any) { G.Warn(message) }

// Warnf logs a formatted message at WARN level using the global logger
func Warnf(format string, vals ...any) { G.Warnf(format, vals...) }

// Error logs a message at ERROR level using the global logger
// message can be a string or an error
func Error(message any) { G.Error(message) }

// Errorf logs a formatted message at ERROR level using the global logger
func Errorf(format string, vals ...any) { G.Errorf(format, vals...) }

// Fatal logs a message at FATAL level and exits the program using the global logger
// message can be a string or an error
func Fatal(message any) { G.Fatal(message) }

// Fatalf logs a formatted message at FATAL level and exits the program using the global logger
func Fatalf(format string, vals ...any) { G.Fatalf(format, vals...) }

// Context returns the current context string from the global logger
func Context() string { return G.Context() }

// SetContext sets a single context value for the global logger
func SetContext(context string) { G = G.SetContext(context) }

// ClearContext removes all context from the global logger
func ClearContext() { G.ClearContext() }

// Add returns a copy of the global logger with additional context
func Add(context string) Logger { return G.Add(context) }

// Wrap wraps an error with the current context from the global logger
// If a string is provided, it will be converted to an error
func Wrap(val any) error { return G.Wrap(val) }

// EnableColors enables colored output for the global logger
func EnableColors() { G.EnableColors() }
