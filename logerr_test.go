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
		LogLevelDebug: "DBG",
		LogLevelInfo:  "INF",
		LogLevelWarn:  "WRN",
		LogLevelError: "ERR",
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

	// Test custom separator
	logger = logger.SetContext("test-context")
	logger = logger.SetContextSeparator(": ")
	newLogger = logger.Add("additional-context")
	customExpected := "test-context: additional-context"
	if newLogger.Context() != customExpected {
		t.Errorf("Expected context with custom separator to be '%s', got '%s'", customExpected, newLogger.Context())
	}

	// Test with empty separator
	logger = logger.SetContextSeparator("")
	newLogger = logger.Add("additional-context")
	if newLogger.Context() != "test-contextadditional-context" {
		t.Errorf("Expected context with empty separator to be '%s', got '%s'", "test-contextadditional-context", newLogger.Context())
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

	// Test basic logging methods with string (single argument - backward compatibility)
	tests := []struct {
		logFunc   func(...any)
		message   any
		levelText string
	}{
		{logger.Debug, "debug test", "[DBG]"},
		{logger.Info, "info test", "[INF]"},
		{logger.Warn, "warn test", "[WRN]"},
		{logger.Error, "error test", "[ERR]"},
	}

	for _, test := range tests {
		buf.Reset()
		test.logFunc(test.message)
		output := buf.String()
		messageStr, ok := test.message.(string)
		if !ok {
			messageStr = fmt.Sprintf("%v", test.message)
		}
		if !strings.Contains(output, test.levelText) || !strings.Contains(output, messageStr) {
			t.Errorf("Expected message to contain '%s' and '%s', got: %s",
				test.levelText, messageStr, output)
		}
	}

	// Test basic logging methods with error (single argument - backward compatibility)
	errorTests := []struct {
		logFunc   func(...any)
		message   any
		levelText string
	}{
		{logger.Debug, errors.New("debug error"), "[DBG]"},
		{logger.Info, errors.New("info error"), "[INF]"},
		{logger.Warn, errors.New("warn error"), "[WRN]"},
		{logger.Error, errors.New("error error"), "[ERR]"},
	}

	for _, test := range errorTests {
		buf.Reset()
		test.logFunc(test.message)
		output := buf.String()
		err, ok := test.message.(error)
		if !ok {
			t.Fatalf("Test message is not an error: %v", test.message)
		}
		if !strings.Contains(output, test.levelText) || !strings.Contains(output, err.Error()) {
			t.Errorf("Expected message to contain '%s' and '%s', got: %s",
				test.levelText, err.Error(), output)
		}
	}

	// Test the new variadic functionality
	variadicTests := []struct {
		logFunc   func(...any)
		args      []any
		levelText string
		expected  []string
	}{
		{logger.Debug, []any{"User logged in", errors.New("with admin rights")}, "[DBG]", []string{"User logged in", "with admin rights"}},
		{logger.Info, []any{"Operation completed", 42, "items processed"}, "[INF]", []string{"Operation completed", "42", "items processed"}},
		{logger.Warn, []any{"Missing configuration", errors.New("defaulting to standard settings")}, "[WRN]", []string{"Missing configuration", "defaulting to standard settings"}},
		{logger.Error, []any{"Failed to connect", errors.New("connection timeout")}, "[ERR]", []string{"Failed to connect", "connection timeout"}},
	}

	for _, test := range variadicTests {
		buf.Reset()
		test.logFunc(test.args...)
		output := buf.String()

		if !strings.Contains(output, test.levelText) {
			t.Errorf("Expected variadic message to contain level '%s', got: %s",
				test.levelText, output)
		}

		for _, expectedStr := range test.expected {
			if !strings.Contains(output, expectedStr) {
				t.Errorf("Expected variadic message to contain '%s', got: %s",
					expectedStr, output)
			}
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
		{logger.Debugf, "debug %s", []any{"formatted"}, "[DBG]", "debug formatted"},
		{logger.Infof, "info %s", []any{"formatted"}, "[INF]", "info formatted"},
		{logger.Warnf, "warn %s", []any{"formatted"}, "[WRN]", "warn formatted"},
		{logger.Errorf, "error %s", []any{"formatted"}, "[ERR]", "error formatted"},
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

	// Test formatted logging with error arguments
	err := errors.New("formatted error")

	// Test each formatted log function with error argument
	buf.Reset()
	logger.Debugf("Got error: %v", err)
	if !strings.Contains(buf.String(), "Got error: formatted error") {
		t.Errorf("Expected Debugf with error to contain error message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Infof("Error occurred: %s", err)
	if !strings.Contains(buf.String(), "Error occurred: formatted error") {
		t.Errorf("Expected Infof with error to contain error message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Warnf("Warning with error: %s", err)
	if !strings.Contains(buf.String(), "Warning with error: formatted error") {
		t.Errorf("Expected Warnf with error to contain error message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Errorf("Error details: %s", err)
	if !strings.Contains(buf.String(), "Error details: formatted error") {
		t.Errorf("Expected Errorf with error to contain error message, got: %s", buf.String())
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
		t.Errorf("DBG level message should be filtered when Level is WRN, got: %s", output)
	}
	if strings.Contains(output, "info test") {
		t.Errorf("INF level message should be filtered when Level is WRN, got: %s", output)
	}
	if !strings.Contains(output, "warn test") {
		t.Errorf("WRN level message should be logged when Level is WRN, got: %s", output)
	}
	if !strings.Contains(output, "error test") {
		t.Errorf("ERR level message should be logged when Level is WRN, got: %s", output)
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
		t.Errorf("DBG level message should be filtered in exclusive mode, got: %s", output)
	}
	if strings.Contains(output, "info exclusive") {
		t.Errorf("INF level message should be filtered in exclusive mode, got: %s", output)
	}
	if !strings.Contains(output, "warn exclusive") {
		t.Errorf("WRN level message should be logged in exclusive mode, got: %s", output)
	}
	if strings.Contains(output, "error exclusive") {
		t.Errorf("ERR level message should be filtered in exclusive mode, got: %s", output)
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

	// Test error wrapping with error type
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

	// Test wrapping with string type
	errString := "string error"
	wrappedStringErr := logger.Wrap(errString)
	expectedString := "test-context | " + errString
	if wrappedStringErr.Error() != expectedString {
		t.Errorf("Expected wrapped string error '%s', got '%s'", expectedString, wrappedStringErr.Error())
	}

	// Test wrapping with other type (int)
	intVal := 42
	wrappedIntErr := logger.Wrap(intVal)
	expectedIntString := "test-context | 42"
	if wrappedIntErr.Error() != expectedIntString {
		t.Errorf("Expected wrapped int error '%s', got '%s'", expectedIntString, wrappedIntErr.Error())
	}

	// Test with LogWrappedErrors disabled
	logger.LogWrappedErrors = false
	logger.Wrap(originalErr) // Should not log the error

	// Test with custom separator
	logger.SetContextSeparator(": ")
	customWrappedErr := logger.Wrap(originalErr)
	customExpected := fmt.Sprintf("test-context: %s", originalErr.Error())
	if customWrappedErr.Error() != customExpected {
		t.Errorf("Expected wrapped error with custom separator '%s', got '%s'", customExpected, customWrappedErr.Error())
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

	// Test global functions with string messages
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

	// Test global functions with error messages
	debugErr := errors.New("debug error")
	Debug(debugErr)
	if !strings.Contains(buf.String(), debugErr.Error()) {
		t.Errorf("Global Debug function failed to log error message")
	}
	buf.Reset()

	infoErr := errors.New("info error")
	Info(infoErr)
	if !strings.Contains(buf.String(), infoErr.Error()) {
		t.Errorf("Global Info function failed to log error message")
	}
	buf.Reset()

	warnErr := errors.New("warn error")
	Warn(warnErr)
	if !strings.Contains(buf.String(), warnErr.Error()) {
		t.Errorf("Global Warn function failed to log error message")
	}
	buf.Reset()

	errorErr := errors.New("error error")
	Error(errorErr)
	if !strings.Contains(buf.String(), errorErr.Error()) {
		t.Errorf("Global Error function failed to log error message")
	}
	buf.Reset()

	// Test the variadic global functions
	Debug("Global debug message", errors.New("with debug error"))
	if !strings.Contains(buf.String(), "Global debug message") || !strings.Contains(buf.String(), "with debug error") {
		t.Errorf("Global Debug variadic function failed to log multiple arguments")
	}
	buf.Reset()

	Info("Global info message", 42, "items")
	if !strings.Contains(buf.String(), "Global info message") || !strings.Contains(buf.String(), "42") || !strings.Contains(buf.String(), "items") {
		t.Errorf("Global Info variadic function failed to log multiple arguments")
	}
	buf.Reset()

	Warn("Global warn message", errors.New("with warning details"))
	if !strings.Contains(buf.String(), "Global warn message") || !strings.Contains(buf.String(), "with warning details") {
		t.Errorf("Global Warn variadic function failed to log multiple arguments")
	}
	buf.Reset()

	Error("Global error message", errors.New("with error details"))
	if !strings.Contains(buf.String(), "Global error message") || !strings.Contains(buf.String(), "with error details") {
		t.Errorf("Global Error variadic function failed to log multiple arguments")
	}
	buf.Reset()

	// We can't test Fatal and Fatalf since they call os.Exit

	// Test global context functions
	// Test SetContext global function
	SetContext("global context")
	// The SetContext function updates the global logger internally

	// Test global Context() function
	contextStr := Context()
	if contextStr != "global context" {
		t.Errorf("Global Context() failed. Expected 'global context', got '%s'", contextStr)
	}

	// Create a new logger with additional context
	newLogger := Add("additional context")
	if !strings.Contains(newLogger.Context(), "additional context") {
		t.Errorf("Global Add failed")
	}

	// Test global ClearContext function
	ClearContext()

	// Verify context is cleared
	contextStr = Context()
	if contextStr != "" {
		t.Errorf("Global Context() after clear failed. Expected empty string, got '%s'", contextStr)
	}

	// Test global error wrapping with error type
	err := errors.New("test error")
	wrappedErr := Wrap(err)
	if !errors.Is(wrappedErr, err) {
		t.Errorf("Global Wrap failed to preserve original error")
	}

	// Test global error wrapping with string
	strWrappedErr := Wrap("string error")
	if !strings.Contains(strWrappedErr.Error(), "string error") {
		t.Errorf("Global Wrap failed to handle string input")
	}

	// Test global color control
	EnableColors()

	// Test global context separator
	SetContextSeparator(": ")
	SetContext("test-context")
	addLogger := Add("additional-context")
	// We need to update the global logger so our context changes are reflected
	G = &addLogger
	customSepContext := Context()
	expectedCustomSep := "test-context: additional-context"
	if customSepContext != expectedCustomSep {
		t.Errorf("Global context with custom separator: expected '%s', got '%s'", expectedCustomSep, customSepContext)
	}
}

func TestFormatLogMessage(t *testing.T) {
	logger := DefaultLogger()
	logger.NoColor = true

	// Test with empty context
	logger.ClearContext()
	message := logger.formatLogMessage(LogLevelInfo, "test message")
	expected := "[INF] test message"
	if message != expected {
		t.Errorf(
			"Expected formatLogMessage with empty context to return %q, got %q",
			expected,
			message,
		)
	}

	// Reset the separator to the default
	logger = logger.SetContextSeparator(" | ")

	// Test with context
	logger = logger.SetContext("test context")
	message = logger.formatLogMessage(LogLevelInfo, "test message")
	expected = "[INF] test context | test message"
	if message != expected {
		t.Errorf("Expected formatLogMessage with context to return %q, got %q", expected, message)
	}

	// Test with custom separator
	logger = logger.SetContextSeparator(": ")
	message = logger.formatLogMessage(LogLevelInfo, "test message")
	expected = "[INF] test context: test message"
	if message != expected {
		t.Errorf("Expected formatLogMessage with custom separator to return %q, got %q", expected, message)
	}

	// Test with timestamps enabled
	logger = logger.EnableTimestamps()
	message = logger.formatLogMessage(LogLevelInfo, "test message")
	// Check that the message contains a timestamp in the expected format (YYYY-MM-DD HH:MM:SS.mmm)
	if !strings.HasPrefix(message, "20") || !strings.Contains(message, ":") || !strings.Contains(message, "[INF]") {
		t.Errorf("Expected message with timestamps to have date/time prefix, got: %q", message)
	}

	// Test with timestamps disabled
	logger = logger.DisableTimestamps()
	message = logger.formatLogMessage(LogLevelInfo, "test message")
	expected = "[INF] test context: test message"
	if message != expected {
		t.Errorf("Expected message without timestamps to equal %q, got: %q", expected, message)
	}
}

func TestMessageToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string input",
			input:    "test string",
			expected: "test string",
		},
		{
			name:     "error input",
			input:    errors.New("test error"),
			expected: "test error",
		},
		{
			name:     "integer input",
			input:    123,
			expected: "123",
		},
		{
			name:     "boolean input",
			input:    true,
			expected: "true",
		},
		{
			name:     "struct input",
			input:    struct{ Name string }{"test"},
			expected: "{test}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := messageToString(test.input)
			if result != test.expected {
				t.Errorf("messageToString(%v) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestTimestampFunctionality(t *testing.T) {
	// Save original global logger and restore after test
	originalG := G
	defer func() {
		G = originalG
	}()

	// Create a test logger
	var buf bytes.Buffer
	logger := DefaultLogger()
	logger.Output = &buf
	logger.Level = LogLevelDebug
	logger.NoColor = true

	// Test EnableTimestamps and DisableTimestamps methods
	logger = logger.EnableTimestamps()
	if !logger.ShowTimestamps {
		t.Errorf("EnableTimestamps() should set ShowTimestamps to true")
	}

	logger = logger.DisableTimestamps()
	if logger.ShowTimestamps {
		t.Errorf("DisableTimestamps() should set ShowTimestamps to false")
	}

	// Test logging with timestamps enabled
	logger = logger.EnableTimestamps()
	buf.Reset()
	logger.Info("test message with timestamp")
	output := buf.String()

	// Output should contain timestamp in format: 2006-01-02 15:04:05.000
	if !strings.Contains(output, "-") || !strings.Contains(output, ":") {
		t.Errorf("Log output with enabled timestamps should contain date/time format, got: %q", output)
	}

	// Test logging with timestamps disabled
	logger = logger.DisableTimestamps()
	buf.Reset()
	logger.Info("test message without timestamp")
	output = buf.String()

	// Output should start with the log level, not a digit (timestamp)
	if strings.HasPrefix(output, "2") {
		t.Errorf("Log output with disabled timestamps should not start with a timestamp, got: %q", output)
	}

	// Test global timestamp functions
	testLogger := DefaultLogger()
	testLogger.SetAsGlobal()

	DisableTimestamps() // Make sure timestamps are disabled initially
	if G.ShowTimestamps {
		t.Errorf("Global logger should have timestamps disabled after DisableTimestamps()")
	}

	EnableTimestamps()
	if !G.ShowTimestamps {
		t.Errorf("Global logger should have timestamps enabled after EnableTimestamps()")
	}

	DisableTimestamps()
	if G.ShowTimestamps {
		t.Errorf("Global logger should have timestamps disabled after DisableTimestamps()")
	}
}

