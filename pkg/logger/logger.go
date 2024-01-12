package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type ConsoleLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	file  *log.Logger
}

type FileLogger struct {
	file *log.Logger
}

func NewConsoleLogger() *ConsoleLogger {
	fileLog, err := os.Create("../general-log.log")
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	return &ConsoleLogger{
		debug: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stdout, "ERROR: ", log.LstdFlags),
		file:  log.New(fileLog, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *ConsoleLogger) Debug(format string, v ...interface{}) {
	l.debug.Printf(format, fmt.Sprint(v...))
	l.file.Printf(format, fmt.Sprint(v...))
}

func (l *ConsoleLogger) Info(format string, v ...interface{}) {
	l.info.Printf(format, fmt.Sprint(v...))
	l.file.Printf(format, fmt.Sprint(v...))
}

func (l *ConsoleLogger) Warn(format string, v ...interface{}) {
	l.warn.Println(format, fmt.Sprint(v...))
	l.file.Printf(format, fmt.Sprint(v...))
}

func (l *ConsoleLogger) Error(format string, v ...interface{}) {
	l.error.Println(format, fmt.Sprint(v...))
	l.file.Printf(format, fmt.Sprint(v...))
}

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}
