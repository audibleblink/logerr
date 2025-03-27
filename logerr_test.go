package logerr

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// Test that log level constants are defined correctly
	levels := []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
	}

	for i, level := range levels {
		if int(level) != i {
			t.Errorf("Expected LogLevel %d to be %d, got %d", i, i, level)
		}
	}

	// Test that log level labels are defined correctly
	expectedLabels := map[LogLevel]string{
		LogLevelDebug: "DEBUG",
		LogLevelInfo:  "INFO",
		LogLevelWarn:  "WARN",
		LogLevelError: "ERROR",
		LogLevelFatal: "FATAL",
	}

	for level, expectedLabel := range expectedLabels {
		if labels[level] != expectedLabel {
			t.Errorf(
				"Expected label for level %d to be %s, got %s",
				level,
				expectedLabel,
				labels[level],
			)
		}
	}
}

func TestLoggerContext(t *testing.T) {
	logger := DefaultLogger()

	// Test SetContext
	logger = logger.SetContext("test-context")
	if logger.Context() != "test-context" {
		t.Errorf("Expected context to be 'test-context', got '%s'", logger.Context())
	}

	// Test Add
	newLogger := logger.Add("additional-context")
	expected := "test-context | additional-context"
	if newLogger.Context() != expected {
		t.Errorf("Expected context to be '%s', got '%s'", expected, newLogger.Context())
	}

	// Test ClearContext
	logger.ClearContext()
	if logger.Context() != "" {
		t.Errorf("Expected context to be empty after clear, got '%s'", logger.Context())
	}
}

func TestLogMessages(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	origOutput := os.Stderr

	// Create a logger with the buffer as output
	logger := DefaultLogger()
	logger.Output = &buf
	logger.Level = LogLevelDebug // Make sure all messages are logged
	logger.NoColor = true        // Disable colors for testing

	// Test basic logging methods
	tests := []struct {
		logFunc   func(string)
		message   string
		levelText string
	}{
		{logger.Debug, "debug test", "[DEBUG]"},
		{logger.Info, "info test", "[INFO]"},
		{logger.Warn, "warn test", "[WARN]"},
		{logger.Error, "error test", "[ERROR]"},
	}

	for _, test := range tests {
		buf.Reset()
		test.logFunc(test.message)
		output := buf.String()
		if !strings.Contains(output, test.levelText) || !strings.Contains(output, test.message) {
			t.Errorf("Expected message to contain '%s' and '%s', got: %s",
				test.levelText, test.message, output)
		}
	}

	// Test formatted logging methods
	formatTests := []struct {
		logFunc   func(string, ...any)
		format    string
		args      []any
		levelText string
		expected  string
	}{
		{logger.Debugf, "debug %s", []any{"formatted"}, "[DEBUG]", "debug formatted"},
		{logger.Infof, "info %s", []any{"formatted"}, "[INFO]", "info formatted"},
		{logger.Warnf, "warn %s", []any{"formatted"}, "[WARN]", "warn formatted"},
		{logger.Errorf, "error %s", []any{"formatted"}, "[ERROR]", "error formatted"},
	}

	for _, test := range formatTests {
		buf.Reset()
		test.logFunc(test.format, test.args...)
		output := buf.String()
		if !strings.Contains(output, test.levelText) || !strings.Contains(output, test.expected) {
			t.Errorf("Expected formatted message to contain '%s' and '%s', got: %s",
				test.levelText, test.expected, output)
		}
	}

	// Restore original output
	logger.Output = origOutput
}

func TestLogLevelFiltering(t *testing.T) {
	// Test with actual log output
	var buf bytes.Buffer
	origOutput := os.Stderr

	// Test non-exclusive mode (default)
	logger := DefaultLogger()
	logger.Output = &buf
	logger.Level = LogLevelWarn
	logger.Exclusive = false
	logger.NoColor = true

	// Test each log level
	logger.Debug("debug test") // Should not log
	logger.Info("info test")   // Should not log
	logger.Warn("warn test")   // Should log
	logger.Error("error test") // Should log

	output := buf.String()
	if strings.Contains(output, "debug test") {
		t.Errorf("DEBUG level message should be filtered when Level is WARN, got: %s", output)
	}
	if strings.Contains(output, "info test") {
		t.Errorf("INFO level message should be filtered when Level is WARN, got: %s", output)
	}
	if !strings.Contains(output, "warn test") {
		t.Errorf("WARN level message should be logged when Level is WARN, got: %s", output)
	}
	if !strings.Contains(output, "error test") {
		t.Errorf("ERROR level message should be logged when Level is WARN, got: %s", output)
	}
	buf.Reset()

	// Test exclusive mode
	logger.Exclusive = true

	logger.Debug("debug exclusive") // Should not log
	logger.Info("info exclusive")   // Should not log
	logger.Warn("warn exclusive")   // Should log
	logger.Error("error exclusive") // Should not log

	output = buf.String()
	if strings.Contains(output, "debug exclusive") {
		t.Errorf("DEBUG level message should be filtered in exclusive mode, got: %s", output)
	}
	if strings.Contains(output, "info exclusive") {
		t.Errorf("INFO level message should be filtered in exclusive mode, got: %s", output)
	}
	if !strings.Contains(output, "warn exclusive") {
		t.Errorf("WARN level message should be logged in exclusive mode, got: %s", output)
	}
	if strings.Contains(output, "error exclusive") {
		t.Errorf("ERROR level message should be filtered in exclusive mode, got: %s", output)
	}
	buf.Reset()

	// Test shouldLog method directly
	tests := []struct {
		level     LogLevel
		logLevel  LogLevel
		exclusive bool
		expected  bool
	}{
		{LogLevelDebug, LogLevelWarn, false, false}, // Debug < Warn (non-exclusive)
		{LogLevelInfo, LogLevelWarn, false, false},  // Info < Warn (non-exclusive)
		{LogLevelWarn, LogLevelWarn, false, true},   // Warn == Warn (non-exclusive)
		{LogLevelError, LogLevelWarn, false, true},  // Error > Warn (non-exclusive)
		{LogLevelFatal, LogLevelWarn, false, true},  // Fatal > Warn (non-exclusive)

		{LogLevelDebug, LogLevelWarn, true, false}, // Debug != Warn (exclusive)
		{LogLevelInfo, LogLevelWarn, true, false},  // Info != Warn (exclusive)
		{LogLevelWarn, LogLevelWarn, true, true},   // Warn == Warn (exclusive)
		{LogLevelError, LogLevelWarn, true, false}, // Error != Warn (exclusive)
		{LogLevelFatal, LogLevelWarn, true, false}, // Fatal != Warn (exclusive)
	}

	for _, test := range tests {
		logger.Level = test.logLevel
		logger.Exclusive = test.exclusive
		result := logger.shouldLog(test.level)
		if result != test.expected {
			t.Errorf("shouldLog(%v) with Level=%v, Exclusive=%v returned %v, expected %v",
				test.level, test.logLevel, test.exclusive, result, test.expected)
		}
	}

	// Restore original output
	logger.Output = origOutput
}

func TestErrorWrapping(t *testing.T) {
	// Create a logger with context
	logger := DefaultLogger().SetContext("test-context")
	logger.LogWrappedErrors = true

	// Test error wrapping
	originalErr := errors.New("original error")
	wrappedErr := logger.Wrap(originalErr)

	// Check that the wrapped error contains the context
	expected := fmt.Sprintf("test-context | %s", originalErr.Error())
	if wrappedErr.Error() != expected {
		t.Errorf("Expected wrapped error '%s', got '%s'", expected, wrappedErr.Error())
	}

	// Check that the original error is still accessible
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("Expected wrapped error to contain original error")
	}
}

func TestColorControl(t *testing.T) {
	logger := DefaultLogger()

	// Test EnableColors and DisableColors
	logger = logger.EnableColors()
	if logger.NoColor {
		t.Errorf("Expected NoColor to be false after EnableColors")
	}

	logger = logger.DisableColors()
	if !logger.NoColor {
		t.Errorf("Expected NoColor to be true after DisableColors")
	}
}

func TestFormatLabel(t *testing.T) {
	// Test with colors disabled
	labelStr := formatLabel(LogLevelDebug, true)
	expected := "[DEBUG]"
	if labelStr != expected {
		t.Errorf("Expected formatLabel with noColor=true to return %q, got %q", expected, labelStr)
	}

	// We can't easily test the colored version without mocking the color package
}

func TestGlobalFunctions(t *testing.T) {
	// Save original global logger and restore after test
	originalG := G
	defer func() {
		G = originalG
	}()

	// Create a test logger
	var buf bytes.Buffer
	testLogger := DefaultLogger()
	testLogger.Output = &buf
	testLogger.Level = LogLevelDebug
	testLogger.NoColor = true

	// Set as global logger
	testLogger.SetAsGlobal()

	// Test global functions
	Debug("global debug")
	if !strings.Contains(buf.String(), "global debug") {
		t.Errorf("Global Debug function failed to log message")
	}
	buf.Reset()

	Debugf("global %s", "debugf")
	if !strings.Contains(buf.String(), "global debugf") {
		t.Errorf("Global Debugf function failed to log message")
	}
	buf.Reset()

	Info("global info")
	if !strings.Contains(buf.String(), "global info") {
		t.Errorf("Global Info function failed to log message")
	}
	buf.Reset()

	Infof("global %s", "infof")
	if !strings.Contains(buf.String(), "global infof") {
		t.Errorf("Global Infof function failed to log message")
	}
	buf.Reset()

	Warn("global warn")
	if !strings.Contains(buf.String(), "global warn") {
		t.Errorf("Global Warn function failed to log message")
	}
	buf.Reset()

	Warnf("global %s", "warnf")
	if !strings.Contains(buf.String(), "global warnf") {
		t.Errorf("Global Warnf function failed to log message")
	}
	buf.Reset()

	Error("global error")
	if !strings.Contains(buf.String(), "global error") {
		t.Errorf("Global Error function failed to log message")
	}
	buf.Reset()

	Errorf("global %s", "errorf")
	if !strings.Contains(buf.String(), "global errorf") {
		t.Errorf("Global Errorf function failed to log message")
	}
	buf.Reset()

	// We can't test Fatal and Fatalf since they call os.Exit

	// Test global context functions
	SetContext("global context")

	// Create a new logger with additional context
	logger := Add("additional context")
	if !strings.Contains(logger.Context(), "additional context") {
		t.Errorf("Global Add failed")
	}

	ClearContext()

	// Test global error wrapping
	err := errors.New("test error")
	wrappedErr := Wrap(err)
	if !errors.Is(wrappedErr, err) {
		t.Errorf("Global Wrap failed to preserve original error")
	}

	// Test global color control
	EnableColors()
}

func TestFormatLogMessage(t *testing.T) {
	logger := DefaultLogger()
	logger.NoColor = true

	// Test with empty context
	logger.ClearContext()
	message := logger.formatLogMessage(LogLevelInfo, "test message")
	expected := "[INFO] | test message"
	if message != expected {
		t.Errorf(
			"Expected formatLogMessage with empty context to return %q, got %q",
			expected,
			message,
		)
	}

	// Test with context
	logger = logger.SetContext("test context")
	message = logger.formatLogMessage(LogLevelInfo, "test message")
	expected = "[INFO] test context | test message"
	if message != expected {
		t.Errorf("Expected formatLogMessage with context to return %q, got %q", expected, message)
	}
}

