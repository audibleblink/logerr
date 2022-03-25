package logerr

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var labels map[LogLevel]string = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
}

var labelColors map[LogLevel]*color.Color = map[LogLevel]*color.Color{
	LogLevelDebug: color.New(color.FgCyan),
	LogLevelInfo:  color.New(color.FgGreen),
	LogLevelWarn:  color.New(color.FgYellow),
	LogLevelError: color.New(color.FgRed),
	LogLevelFatal: color.New(color.BgRed),
}

// Logger is logger
type Logger struct {

	// Level dictates the LogLevel
	Level LogLevel

	// Output destination
	Output io.Writer

	// Exclusive dictates whether _only_ the configured loglevel messages are shown
	// Defaults to printing everything below the configured LogLevel
	// Fatal, Error, Warn, Info, Debug
	Exclusive bool

	// LogWrappedErrors, when enabled, will print the error text,
	// according to level and context, before returning the error
	LogWrappedErrors bool

	// Additional prefix text to add context to log messages
	context []string

	// Enable colored status printing
	NoColor bool
}

var G = DefaultLogger()

func Debug(s string)                       { G.Debug(s) }
func Debugf(s string, vals ...interface{}) { G.Debugf(s, vals) }

func Info(s string)                       { G.Info(s) }
func Infof(s string, vals ...interface{}) { G.Infof(s, vals) }

func Warn(s string)                       { G.Warn(s) }
func Warnf(s string, vals ...interface{}) { G.Warnf(s, vals) }

func Error(s string)                       { G.Error(s) }
func Errorf(s string, vals ...interface{}) { G.Errorf(s, vals) }

func Fatal(s string)                       { G.Fatal(s) }
func Fatalf(s string, vals ...interface{}) { G.Fatalf(s, vals) }

func Context() string           { return G.Context() }
func SetContext(context string) { G.SetContext(context) }
func ClearContext()             { G.ClearContext() }
func Add(context string) Logger { return G.Add(context) }
func Wrap(err error) error      { return G.Wrap(err) }
func EnableColors()             { G.EnableColors() }

func DefaultLogger() *Logger {
	logger := &Logger{
		Level:   LogLevelError,
		Output:  os.Stderr,
		NoColor: true,
	}
	color.NoColor = logger.NoColor
	return logger.SetContext("")
}

func (l Logger) SetAsGlobal() {
	G = &l
}

func (l Logger) Context() string {
	ctx := strings.Join(l.context, " | ")
	return ctx
}

func (l *Logger) ClearContext() {
	l.context = make([]string, 0)
}

func (l Logger) SetContext(s string) *Logger {
	l.context = []string{s}
	return &l
}

// Add returns a copy of d with additional context. Useful for loggers
// that can die once the scope in which they're defined exits
func (l *Logger) Add(context string) Logger {
	dup := *l
	dup.context = append(dup.context, context)
	return dup
}

func (l *Logger) EnableColors() *Logger {
	l.NoColor = false
	return l
}

func (l *Logger) DisableColors() *Logger {
	l.NoColor = true
	return l
}

func (l Logger) Debug(s string) {
	loggerGen(LogLevelDebug, &l)(s)
}

func (l Logger) Debugf(s string, vals ...interface{}) {
	loggerGenF(LogLevelDebug, &l)(s, vals...)
}

func (l Logger) Info(s string) {
	loggerGen(LogLevelInfo, &l)(s)
}

func (l Logger) Infof(s string, vals ...interface{}) {
	loggerGenF(LogLevelInfo, &l)(s, vals...)
}

func (l Logger) Warn(s string) {
	loggerGen(LogLevelWarn, &l)(s)
}

func (l Logger) Warnf(s string, vals ...interface{}) {
	loggerGenF(LogLevelWarn, &l)(s, vals...)
}

func (l Logger) Error(s string) {
	loggerGen(LogLevelError, &l)(s)
}

func (l Logger) Errorf(s string, vals ...interface{}) {
	loggerGenF(LogLevelError, &l)(s, vals...)
}

func (l Logger) Fatal(s string) {
	loggerGen(LogLevelFatal, &l)(s)
	os.Exit(1)
}

func (l Logger) Fatalf(s string, vals ...interface{}) {
	loggerGenF(LogLevelFatal, &l)(s, vals...)
	os.Exit(1)
}

func (l Logger) Wrap(err error) error {
	if l.LogWrappedErrors {
		l.Error(err.Error())
	}
	return fmt.Errorf("%s | %w", l.Context(), err)
}

// generator that return a configured logger function
func loggerGen(level LogLevel, l *Logger) func(string) {
	color.NoColor = l.NoColor
	newTmpl := fmt.Sprint(label(level), l.Context())
	return func(s string) {
		if (l.Level == level && l.Exclusive) || (l.Level <= level && !l.Exclusive) {
			out := fmt.Sprintf("%s | %s", newTmpl, s)
			log.Print(out)
		}
	}
}

// generator that returns a configured formatting logger function
func loggerGenF(level LogLevel, l *Logger) func(string, ...interface{}) {
	color.NoColor = l.NoColor
	newTmpl := fmt.Sprint(label(level), l.Context())
	return func(fmtString string, vals ...interface{}) {
		msg := fmt.Sprintf(fmtString, vals...)
		if (l.Level == level && l.Exclusive) || (l.Level <= level && !l.Exclusive) {
			fmtMsg := fmt.Sprint(newTmpl, " | ", msg)
			log.Print(fmtMsg)
		}
	}
}

func label(lvl LogLevel) string {
	if color.NoColor {
		return fmt.Sprintf("[%s] ", labels[lvl])
	}

	labelColor := labelColors[lvl]
	return labelColor.Sprintf("[%s] ", labels[lvl])
}
