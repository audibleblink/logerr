package logerr

import (
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
func (l Logger) Wrap(err error) error {
	if l.LogWrappedErrors {
		l.Error(err.Error())
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

// log outputs a message if it should be logged based on level
func (l *Logger) log(level LogLevel, msg string) {
	if l.shouldLog(level) {
		formatted := l.formatLogMessage(level, msg)
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
func (l Logger) Debug(msg string) {
	l.log(LogLevelDebug, msg)
}

// Debugf logs a formatted message at DEBUG level
func (l Logger) Debugf(format string, args ...any) {
	l.logf(LogLevelDebug, format, args...)
}

// Info logs a message at INFO level
func (l Logger) Info(msg string) {
	l.log(LogLevelInfo, msg)
}

// Infof logs a formatted message at INFO level
func (l Logger) Infof(format string, args ...any) {
	l.logf(LogLevelInfo, format, args...)
}

// Warn logs a message at WARN level
func (l Logger) Warn(msg string) {
	l.log(LogLevelWarn, msg)
}

// Warnf logs a formatted message at WARN level
func (l Logger) Warnf(format string, args ...any) {
	l.logf(LogLevelWarn, format, args...)
}

// Error logs a message at ERROR level
func (l Logger) Error(msg string) {
	l.log(LogLevelError, msg)
}

// Errorf logs a formatted message at ERROR level
func (l Logger) Errorf(format string, args ...any) {
	l.logf(LogLevelError, format, args...)
}

// Fatal logs a message at FATAL level and exits the program
func (l Logger) Fatal(msg string) {
	l.log(LogLevelFatal, msg)
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
func Debug(s string) { G.Debug(s) }

// Debugf logs a formatted message at DEBUG level using the global logger
func Debugf(s string, vals ...any) { G.Debugf(s, vals...) }

// Info logs a message at INFO level using the global logger
func Info(s string) { G.Info(s) }

// Infof logs a formatted message at INFO level using the global logger
func Infof(s string, vals ...any) { G.Infof(s, vals...) }

// Warn logs a message at WARN level using the global logger
func Warn(s string) { G.Warn(s) }

// Warnf logs a formatted message at WARN level using the global logger
func Warnf(s string, vals ...any) { G.Warnf(s, vals...) }

// Error logs a message at ERROR level using the global logger
func Error(s string) { G.Error(s) }

// Errorf logs a formatted message at ERROR level using the global logger
func Errorf(s string, vals ...any) { G.Errorf(s, vals...) }

// Fatal logs a message at FATAL level and exits the program using the global logger
func Fatal(s string) { G.Fatal(s) }

// Fatalf logs a formatted message at FATAL level and exits the program using the global logger
func Fatalf(s string, vals ...any) { G.Fatalf(s, vals...) }

// Context returns the current context string from the global logger
func Context() string { return G.Context() }

// SetContext sets a single context value for the global logger
func SetContext(context string) { G.SetContext(context) }

// ClearContext removes all context from the global logger
func ClearContext() { G.ClearContext() }

// Add returns a copy of the global logger with additional context
func Add(context string) Logger { return G.Add(context) }

// Wrap wraps an error with the current context from the global logger
func Wrap(err error) error { return G.Wrap(err) }

// EnableColors enables colored output for the global logger
func EnableColors() { G.EnableColors() }
