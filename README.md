# logerr

A simple Go package for error logging with some added comfort features.

## Features

- **Contextual Error Wrapping**: Automatically wrap errors with context information using the `Wrap()` method
- **Exclusive Log Levels**: Option to show only a specific log level with the `Exclusive` flag
- **Colored Output**: Configurable colored log level indicators
- **Logger Chaining**: Create context-specific loggers with the `Add()` method
- **Error Auto-Logging**: Configurable automatic logging of wrapped errors with `LogWrappedErrors`
- **Global and Instance Loggers**: Use the global logger or create custom instances
- **Context Management**: Add, set, and clear context information that gets included in log messages

