package main

import (
    "errors"
    "io"
    "log"
    "os"
)

// Logger struct defines the path to the log file
type Logger struct {
    fileHandle  *os.File
    infoLogger  *log.Logger
    errorLogger *log.Logger
}

// Info prints and logs a specified info message
func (l *Logger) Info(format string, v ...interface{}) {
    l.infoLogger.Printf(format, v...)
}

// Error prints and logs a specified error message
func (l *Logger) Error(format string, v ...interface{}) {
    l.errorLogger.Printf(format, v...)
}

// CloseFile closes the file handle
func (l *Logger) CloseFile() error {
    return l.fileHandle.Close()
}

// NewFileLogger opens the logging file specified and redirects log output to 
// both the file and stdout. It returns pointer to Logger object.
func NewFileLogger(path string) (*Logger, error) {
    // check if log file exists, if not then create it and continue
    _, err := os.Stat(path)
    if errors.Is(err, os.ErrNotExist) {
        _, err = os.Create(path)
        if err != nil {
            return nil, err
        }
    }

    // Open log file
    f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if e != nil {
        return nil, e
    }

    infoMultiWriter := io.MultiWriter(os.Stdout, f)
    errMultiWriter := io.MultiWriter(os.Stderr, f)

    return &Logger{
        fileHandle:  f,
        infoLogger:  log.New(
            infoMultiWriter, 
            "INFO: ", 
            log.Ldate|log.Ltime|log.Lmsgprefix,
        ),
        errorLogger: log.New(
            errMultiWriter, 
            "ERROR: ", 
            log.Ldate|log.Ltime|log.Lmsgprefix,
        ),
    }, nil
}