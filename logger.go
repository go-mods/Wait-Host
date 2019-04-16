package waithost

import (
	"log"
	"os"
)

var (
	defaultLogger = Logger{log.New(os.Stdout, "\n", 0)}
)

type logger interface {
	Print(v ...interface{})
}

// LogWriter log writer interface
type LogWriter interface {
	Println(v ...interface{})
}

// Logger default logger
type Logger struct {
	LogWriter
}

// Print format & print log
func (logger Logger) Print(values ...interface{}) {
	logger.Println(values...)
}
