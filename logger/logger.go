package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// ============ Constants =============

const GenerateDirectoryPermissionMode = 0750
const FileWritePermissionMode = 0644

// ============ Internal(private) Methods - can only be called from inside this package ==============

// for requests logging
var logWriter io.Writer
var logger *log.Logger
var logInit sync.Once

type Logger interface {
	Infof(msg string, args ...any)
	Errorf(msg string, args ...any)
	Print(v ...interface{})
}

func initializeLogger(logPrefix, logPath string) {
	logDirSplit := strings.Split(logPath, "/")
	var logDirSlice []string
	for i := 0; i < len(logDirSplit)-1; i++ {
		logDirSlice = append(logDirSlice, logDirSplit[i])
	}
	logDir := strings.Join(logDirSlice, "/")
	logFileName := logDirSplit[len(logDirSplit)-1]

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, os.FileMode(GenerateDirectoryPermissionMode))
		if err != nil {
			reason := fmt.Sprintf("create directory %s failed: %s", logDir, err)
			err = errors.New(reason)
			fmt.Println(err)
			return
		}
	}
	txnLogFile := logDir + "/" + logFileName
	if file, err := os.OpenFile(txnLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,
		os.FileMode(FileWritePermissionMode)); err != nil {
		panic(err)
	} else {
		logWriter = io.MultiWriter(file)
		logger = log.New(logWriter, logPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Llongfile)
		logger.Printf("logger initialized at %s", txnLogFile)
	}
}

// =========== Exposed (public) Methods - can be called from external packages ============

// GetLogger returns the logger object. It takes two input parameters.
// - logPrefix - it is a string used as Prefix on each log line
// - logPath - absolute path of the log file where the logs will be written
func GetLogger(logPrefix, logPath string) *log.Logger {
	logInit.Do(func() {
		initializeLogger(logPrefix, logPath)
	})
	return logger
}

// GetLogWriter returns an writer interface. This can be used in the middleware.
func GetLogWriter(logPrefix, logPath string) *io.Writer {
	logInit.Do(func() {
		initializeLogger(logPrefix, logPath)
	})
	return &logWriter
}
