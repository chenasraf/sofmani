package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

type Logger struct {
	fileLogger *log.Logger
	consoleOut *log.Logger
	logFile    *os.File
	debug      bool
}

var logger *Logger

func GetLogDir() string {
	var logDir string
	switch runtime.GOOS {
	case "linux":
		logDir = filepath.Join("var", "log", "sofmani")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Could not get user home directory: %v\n", err)
			panic(err)
		}
		logDir = filepath.Join(home, "Library", "Logs", "sofmani")
	case "windows":
		appData := os.Getenv("APPDATA")
		logDir = filepath.Join(appData, "sofmani", "Logs")
	}
	return logDir
}

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

func (l *Logger) log(level string, colorizer *color.Color, format string, args ...interface{}) {
	// Create timestamped message
	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, args...))

	// Write to file
	l.fileLogger.Println(message)

	// Write to console with color
	if colorizer != nil {
		l.consoleOut.Println(colorizer.Sprint(message))
	} else {
		l.consoleOut.Println(message)
	}
}

func Info(format string, args ...interface{}) {
	colorBlue := color.New(color.FgBlue).Add(color.Bold)
	logger.log(" INFO", colorBlue, format, args...)
}

func Warn(format string, args ...interface{}) {
	colorYellow := color.New(color.FgYellow).Add(color.Bold)
	logger.log(" WARN", colorYellow, format, args...)
}

func Error(format string, args ...interface{}) {
	colorRed := color.New(color.FgRed).Add(color.Bold)
	logger.log("ERROR", colorRed, format, args...)
}

func Debug(format string, args ...interface{}) {
	if !logger.debug {
		return
	}
	colorGreen := color.New(color.FgGreen).Add(color.Bold)
	logger.log("DEBUG", colorGreen, format, args...)
}

func Spew(v interface{}) {
	// Print/debug the passed value (works like spew.Dump)
	spewDump := spew.Sdump(v)
	Debug("%s", spewDump)
}

func CloseLogger() {
	if logger != nil && logger.logFile != nil {
		logger.logFile.Close()
	}
}
