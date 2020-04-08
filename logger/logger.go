package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// ============ Constants =============
//	A space is required at the end of prefix to have clear
//	visibility in the log files.
const (
	LogPrefix = "TEST "
)

const GenerateDirectoryPermissionMode = 0750
const FileWritePermissionMode = 0644
const FileReadPermissionMode = 0644

const (
	LogDirPath  = "/rest/logs/"
	LogFileName = "debug.log"
)

// ============ Internal(private) Methods - can only be called from inside this package ==============

// for requests logging
var logWriter io.Writer
var logger *log.Logger
var logInit sync.Once

func initializeLogger() {
	logDir := LogDirPath
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, os.FileMode(GenerateDirectoryPermissionMode))
		if err != nil {
			reason := fmt.Sprintf("error while creating %s directory: %s", logDir, err)
			GetLogger().Print(reason)
			err = errors.New(reason)
			return
		}
	}
	txnLogFile := logDir + LogFileName
	if file, err := os.OpenFile(txnLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(FileWritePermissionMode)); err != nil {
		panic(err)
	} else {
		logWriter = io.MultiWriter(file)
		logger = log.New(logWriter, LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Llongfile)
	}
}

// =========== Exposed (public) Methods - can be called from external packages ============

// used for middleware, as they require an io.Writer for writing logs
func GetLogger() *log.Logger {
	logInit.Do(func() {
		initializeLogger()
	})
	return logger
}

// used for middleware, as they require an io.Writer for writing logs
func GetLogWriter() *io.Writer {
	logInit.Do(func() {
		initializeLogger()
	})
	return &logWriter
}
