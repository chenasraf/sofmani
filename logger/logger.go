package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/platform"
	"github.com/davecgh/go-spew/spew"
)

// Highlight markers (using unlikely byte sequences)
const (
	highlightStart = "\x00HS\x00"
	highlightEnd   = "\x00HE\x00"
)

// ANSI color codes
const (
	ansiHighlight  = "\033[1;96m" // Bright cyan for highlights
	ansiBlueBold   = "\033[1;34m"
	ansiYellowBold = "\033[1;33m"
	ansiRedBold    = "\033[1;31m"
	ansiGreenBold  = "\033[1;32m"
	ansiReset      = "\033[0m"
)

// Highlight marks text to be displayed in white/bold in console output.
// Use this for installer names, types, and other important identifiers.
func Highlight(text string) string {
	return highlightStart + text + highlightEnd
}

// H is a shorthand alias for Highlight.
func H(text string) string {
	return Highlight(text)
}

// Logger provides logging functionality with support for file and console output.
type Logger struct {
	fileLogger  *log.Logger // fileLogger is the logger for writing to the log file.
	consoleOut  *log.Logger // consoleOut is the logger for writing to the console.
	logFile     *os.File    // logFile is the opened log file.
	logFilePath string      // logFilePath is the path to the log file.
	debug       bool        // debug indicates whether debug logging is enabled.
}

var logger *Logger        // logger is the global logger instance.
var customLogFile *string // customLogFile holds the custom log file path if set.

// GetLogDir returns the appropriate log directory based on the operating system.
func GetLogDir() string {
	var logDir string
	switch platform.GetPlatform() {
	case platform.PlatformLinux:
		stateDir := os.Getenv("XDG_STATE_HOME")
		if stateDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Could not get user home directory: %v\n", err)
				panic(err)
			}
			stateDir = filepath.Join(home, ".local", "state")
		}
		logDir = filepath.Join(stateDir, "sofmani")
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

// GetDefaultLogFile returns the default log file path.
func GetDefaultLogFile() string {
	return filepath.Join(GetLogDir(), "sofmani.log")
}

// GetLogFile returns the current log file path (custom or default).
func GetLogFile() string {
	if customLogFile != nil {
		return *customLogFile
	}
	return GetDefaultLogFile()
}

// SetLogFile sets a custom log file path.
func SetLogFile(path string) {
	customLogFile = &path
}

const maxLogSize = 10 * 1024 * 1024 // 10MB

// InitLogger initializes the global logger with the specified debug mode.
// It creates the log directory and file if they don't exist.
// If the log file exceeds maxLogSize, it will be truncated.
func InitLogger(debug bool) *Logger {
	filePath := GetLogFile()
	logDir := filepath.Dir(filePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			fmt.Printf("Could not create log directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Truncate log file if it exceeds maxLogSize
	if info, err := os.Stat(filePath); err == nil && info.Size() > maxLogSize {
		if err := os.Truncate(filePath, 0); err != nil {
			fmt.Printf("Could not truncate log file: %v\n", err)
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
		fileLogger:  fileLogger,
		consoleOut:  consoleOut,
		logFile:     logFile,
		logFilePath: filePath,
		debug:       debug,
	}

	return logger
}

// stripHighlightMarkers removes highlight markers from text (for file output).
func stripHighlightMarkers(text string) string {
	text = strings.ReplaceAll(text, highlightStart, "")
	text = strings.ReplaceAll(text, highlightEnd, "")
	return text
}

// processHighlights converts highlight markers to ANSI codes for console output.
func processHighlights(text string, baseColorSeq string) string {
	// Replace start marker with highlight color
	text = strings.ReplaceAll(text, highlightStart, ansiHighlight)
	// Replace end marker with reset + base color (to restore the log level color)
	text = strings.ReplaceAll(text, highlightEnd, ansiReset+baseColorSeq)
	return text
}

// log is an internal helper function for logging messages with a specific level and color.
func (l *Logger) log(level string, colorSeq string, format string, args ...any) {
	message := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, args...))

	// Write to file (strip all highlight markers - file should have no colors)
	fileMessage := stripHighlightMarkers(message)
	l.fileLogger.Println(fileMessage)

	if level == "DEBUG" && !l.debug {
		return
	}

	// Write to console with color
	if colorSeq != "" {
		consoleMessage := processHighlights(message, colorSeq)
		// Wrap entire message in base color and reset at end
		l.consoleOut.Println(colorSeq + consoleMessage + ansiReset)
	} else {
		// No base color - just convert highlights to white
		consoleMessage := processHighlights(message, "")
		l.consoleOut.Println(consoleMessage)
	}
}

// Info logs an informational message.
func Info(format string, args ...any) {
	logger.log(" INFO", ansiBlueBold, format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	logger.log(" WARN", ansiYellowBold, format, args...)
}

// Error logs an error message.
func Error(format string, args ...any) {
	logger.log("ERROR", ansiRedBold, format, args...)
}

// Debug logs a debug message. Only printed if debug mode is enabled.
func Debug(format string, args ...any) {
	logger.log("DEBUG", ansiGreenBold, format, args...)
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
