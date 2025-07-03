package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/chenasraf/sofmani/platform"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

// Logger provides logging functionality with support for file and console output.
type Logger struct {
	fileLogger *log.Logger // fileLogger is the logger for writing to the log file.
	consoleOut *log.Logger // consoleOut is the logger for writing to the console.
	logFile    *os.File    // logFile is the opened log file.
	debug      bool        // debug indicates whether debug logging is enabled.
}

var logger *Logger // logger is the global logger instance.

// GetLogDir returns the appropriate log directory based on the operating system.
func GetLogDir() string {
	var logDir string
	switch platform.GetPlatform() {
	case platform.PlatformLinux:
		logDir = filepath.Join("var", "log", "sofmani")
	case platform.PlatformMacos:
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Could not get user home directory: %v\n", err)
			panic(err)
		}
		logDir = filepath.Join(home, "Library", "Logs", "sofmani")
	case platform.PlatformWindows:
		appData := os.Getenv("APPDATA")
		logDir = filepath.Join(appData, "sofmani", "Logs")
	}
	return logDir
}

// InitLogger initializes the global logger with the specified debug mode.
// It creates the log directory and file if they don't exist.
func InitLogger(debug bool) *Logger {
	logDir := GetLogDir()
	filePath := filepath.Join(logDir, "sofmani.log")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			fmt.Printf("Could not create log directory: %v\n", err)
			os.Exit(1)
		}
	}
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not create log file: %v\n", err)
		os.Exit(1)
	}

	// Create file and console loggers
	fileLogger := log.New(logFile, "", log.LstdFlags)
	consoleOut := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize the logger
	logger = &Logger{
		fileLogger: fileLogger,
		consoleOut: consoleOut,
		logFile:    logFile,
		debug:      debug,
	}

	return logger
}

// log is an internal helper function for logging messages with a specific level and color.
func (l *Logger) log(level string, colorizer *color.Color, format string, args ...any) {
	// Create timestamped message
	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, args...))

	// Write to file
	l.fileLogger.Println(message)

	if level == "DEBUG" && !l.debug {
		return
	}

	// Write to console with color
	if colorizer != nil {
		l.consoleOut.Println(colorizer.Sprint(message))
	} else {
		l.consoleOut.Println(message)
	}
}

// Info logs an informational message.
func Info(format string, args ...any) {
	colorBlue := color.New(color.FgBlue).Add(color.Bold)
	logger.log(" INFO", colorBlue, format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	colorYellow := color.New(color.FgYellow).Add(color.Bold)
	logger.log(" WARN", colorYellow, format, args...)
}

// Error logs an error message.
func Error(format string, args ...any) {
	colorRed := color.New(color.FgRed).Add(color.Bold)
	logger.log("ERROR", colorRed, format, args...)
}

// Debug logs a debug message. Only printed if debug mode is enabled.
func Debug(format string, args ...any) {
	colorGreen := color.New(color.FgGreen).Add(color.Bold)
	logger.log("DEBUG", colorGreen, format, args...)
}

// Spew logs a detailed representation of a value using spew.Dump.
// This is typically used for debugging complex data structures.
func Spew(v any) {
	// Print/debug the passed value (works like spew.Dump)
	spewDump := spew.Sdump(v)
	Debug("%s", spewDump)
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logger != nil && logger.logFile != nil {
		err := logger.logFile.Close()
		if err != nil {
			fmt.Printf("Could not close log file: %v\n", err)
		}
	}
}
