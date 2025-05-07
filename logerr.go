package logerr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	LogLevelDebug: "DBG",
	LogLevelInfo:  "INF",
	LogLevelWarn:  "WRN",
	LogLevelError: "ERR",
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

	// ShowTimestamps adds timestamps to log messages when true
	ShowTimestamps bool

	// ContextSeparator is used to join context elements
	// Defaults to " | "
	ContextSeparator string
}

// DefaultLogger creates a new logger with default settings
func DefaultLogger() *Logger {
	logger := &Logger{
		Level:            LogLevelError,
		Output:           os.Stderr,
		NoColor:          true,
		ContextSeparator: " | ",
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
	return strings.Join(l.context, l.ContextSeparator)
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

// SetContextSeparator sets the separator used for joining context elements
func (l *Logger) SetContextSeparator(separator string) *Logger {
	l.ContextSeparator = separator
	return l
}

// EnableTimestamps enables timestamps in log messages
func (l *Logger) EnableTimestamps() *Logger {
	l.ShowTimestamps = true
	return l
}

// DisableTimestamps disables timestamps in log messages
func (l *Logger) DisableTimestamps() *Logger {
	l.ShowTimestamps = false
	return l
}

// SetLogLevel sets the log level logger
func (l *Logger) SetLogLevel(lvl LogLevel) *Logger {
	l.Level = lvl
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

	return fmt.Errorf("%s%s%w", l.Context(), l.ContextSeparator, err)
}

// shouldLog determines if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	return (l.Level == level && l.Exclusive) || (l.Level <= level && !l.Exclusive)
}

// formatLogMessage creates a formatted log message with the level and context
func (l *Logger) formatLogMessage(level LogLevel, msg string) string {
	prefix := formatLabel(level, l.NoColor)
	ctx := l.Context()

	var timestamp string
	if l.ShowTimestamps {
		timestamp = time.Now().Format("2006-01-02 15:04:05.000 ")
	}

	if ctx == "" {
		return fmt.Sprintf("%s%s%s", timestamp, prefix, msg)
	}

	return fmt.Sprintf("%s%s%s%s%s", timestamp, prefix, ctx, l.ContextSeparator, msg)
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
// first argument can be a string or an error, any additional arguments are appended
func (l *Logger) log(level LogLevel, args ...any) {
	if l.shouldLog(level) {
		if len(args) == 0 {
			// No arguments provided
			return
		}

		// Format the message based on the number of arguments
		var msgStr string
		if len(args) == 1 {
			// Single argument case (backwards compatibility)
			msgStr = messageToString(args[0])
		} else {
			// Multiple arguments case
			msgParts := make([]string, len(args))
			for i, arg := range args {
				msgParts[i] = messageToString(arg)
			}
			msgStr = strings.Join(msgParts, " ")
		}

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
// First argument can be a string or an error, any additional arguments are appended
func (l Logger) Debug(args ...any) {
	l.log(LogLevelDebug, args...)
}

// Debugf logs a formatted message at DEBUG level
func (l Logger) Debugf(format string, args ...any) {
	l.logf(LogLevelDebug, format, args...)
}

// Info logs a message at INFO level
// First argument can be a string or an error, any additional arguments are appended
func (l Logger) Info(args ...any) {
	l.log(LogLevelInfo, args...)
}

// Infof logs a formatted message at INFO level
func (l Logger) Infof(format string, args ...any) {
	l.logf(LogLevelInfo, format, args...)
}

// Warn logs a message at WARN level
// First argument can be a string or an error, any additional arguments are appended
func (l Logger) Warn(args ...any) {
	l.log(LogLevelWarn, args...)
}

// Warnf logs a formatted message at WARN level
func (l Logger) Warnf(format string, args ...any) {
	l.logf(LogLevelWarn, format, args...)
}

// Error logs a message at ERROR level
// First argument can be a string or an error, any additional arguments are appended
func (l Logger) Error(args ...any) {
	l.log(LogLevelError, args...)
}

// Errorf logs a formatted message at ERROR level
func (l Logger) Errorf(format string, args ...any) {
	l.logf(LogLevelError, format, args...)
}

// Fatal logs a message at FATAL level and exits the program
// First argument can be a string or an error, any additional arguments are appended
func (l Logger) Fatal(args ...any) {
	l.log(LogLevelFatal, args...)
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
		return fmt.Sprintf("[%s] ", labelText)
	}

	return labelColors[level].Sprintf("[%s] ", labelText)
}

// Global convenience functions that use the default logger

// Debug logs a message at DEBUG level using the global logger
// First argument can be a string or an error, any additional arguments are appended
func Debug(args ...any) { G.Debug(args...) }

// Debugf logs a formatted message at DEBUG level using the global logger
func Debugf(format string, vals ...any) { G.Debugf(format, vals...) }

// Info logs a message at INFO level using the global logger
// First argument can be a string or an error, any additional arguments are appended
func Info(args ...any) { G.Info(args...) }

// Infof logs a formatted message at INFO level using the global logger
func Infof(format string, vals ...any) { G.Infof(format, vals...) }

// Warn logs a message at WARN level using the global logger
// First argument can be a string or an error, any additional arguments are appended
func Warn(args ...any) { G.Warn(args...) }

// Warnf logs a formatted message at WARN level using the global logger
func Warnf(format string, vals ...any) { G.Warnf(format, vals...) }

// Error logs a message at ERROR level using the global logger
// First argument can be a string or an error, any additional arguments are appended
func Error(args ...any) { G.Error(args...) }

// Errorf logs a formatted message at ERROR level using the global logger
func Errorf(format string, vals ...any) { G.Errorf(format, vals...) }

// Fatal logs a message at FATAL level and exits the program using the global logger
// First argument can be a string or an error, any additional arguments are appended
func Fatal(args ...any) { G.Fatal(args...) }

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

// SetContextSeparator sets the separator used for joining context elements for the global logger
func SetContextSeparator(separator string) { G = G.SetContextSeparator(separator) }

// EnableTimestamps enables timestamps in log messages for the global logger
func EnableTimestamps() { G = G.EnableTimestamps() }

// DisableTimestamps disables timestamps in log messages for the global logger
func DisableTimestamps() { G = G.DisableTimestamps() }

// SetLogLevel sets the log level for the global logger
func SetLogLevel(lvl LogLevel) { G = G.SetLogLevel(lvl) }
