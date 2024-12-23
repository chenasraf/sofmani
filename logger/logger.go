package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/chenasraf/sofmani/appconfig"
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

func InitLogger(filePath string, config *appconfig.AppConfig) *Logger {
	// Open the log file
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
		debug:      config.Debug,
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

