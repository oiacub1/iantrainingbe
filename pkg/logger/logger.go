package logger

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l *Logger) Error(msg string, err error) {
	if err != nil {
		l.errorLogger.Printf("%s: %v\n", msg, err)
	} else {
		l.errorLogger.Println(msg)
	}
}

func (l *Logger) Fatal(msg string, err error) {
	if err != nil {
		l.errorLogger.Fatalf("%s: %v\n", msg, err)
	} else {
		l.errorLogger.Fatal(msg)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}
