package logger

import (
	"fmt"
	"os"
)

// Debug prints debug messages only if LOG environment variable is set to DEBUG
func Debug(format string, args ...interface{}) {
	logLevel := os.Getenv("LOG")
	if logLevel == "DEBUG" {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format, args...)
	}
}

// Info prints informational messages
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "INFO: "+format, args...)
}

// Error prints error messages
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format, args...)
}
