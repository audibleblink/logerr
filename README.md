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

## Examples

### Contextual Error Wrapping

```go
package main

import (
    "errors"
    "fmt"
    
    "github.com/audibleblink/logerr"
)

func fetchData() error {
    return errors.New("database connection failed")
}

func main() {
    // Set context for the global logger
    logerr.SetContext("UserService")
    
    err := fetchData()
    if err != nil {
        // Wrap the error with context
        wrappedErr := logerr.Wrap(err)
        fmt.Println(wrappedErr) 
        // Outputs: "UserService | database connection failed"
    }
}
```

### Logger Chaining with Context

```go
package main

import (
    "github.com/audibleblink/logerr"
)

func main() {
    // Create a base logger
    baseLogger := logerr.DefaultLogger().EnableColors()
    baseLogger.SetContext("API")
    
    // Create context-specific loggers
    authLogger := baseLogger.Add("Auth")
    dbLogger := baseLogger.Add("Database")
    
    // Use the loggers
    authLogger.Info("User authentication successful")
    dbLogger.Error("Connection pool exhausted")
    
    // Output:
    // [INFO] API | Auth | User authentication successful
    // [ERROR] API | Database | Connection pool exhausted
}
```

### Exclusive Log Levels and Auto-Logging Errors

```go
package main

import (
    "errors"
    "fmt"
    
    "github.com/audibleblink/logerr"
)

func ThingsAndStuff() error {
    // Create a custom logger
    mylog := logerr.Add("ThingsAndStuff")
    mylog.Level = logerr.LogLevelError
    mylog.Exclusive = true // Only show ERROR level messages
    mylog.LogWrappedErrors = true // Auto-log errors when wrapped
    
    mylog.Info("Starting process") // This won't be displayed
    mylog.Error("Process failed")
    // [ERROR] ThingsAndStuff | Process failed
    
    err := errors.New("validation failed")
    // This will auto-log the error before returning it
    return mylog.Wrap(err)
    // [ERROR]  | ThingsAndStuff | validation failed
}

func main() { 
    err := ThingsAndStuff()
    fmt.Println(err)
    // ThingsAndStuff | Process failed
}
```

