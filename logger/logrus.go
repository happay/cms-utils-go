package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/sirupsen/logrus"
)

// ============ Internal(private) Methods - can only be called from inside this package ==============

var loggerV2 *logrus.Logger

func initializeLoggerV2(logPrefix, logPath, appName string) {
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
	txnLogFile := logDir + "/" + logFileName + ".log"
	if file, err := os.OpenFile(txnLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,
		os.FileMode(FileWritePermissionMode)); err != nil {
		panic(err)
	} else {
		logWriter = io.MultiWriter(file)
		hook := GrayLogHook(appName)
		defer hook.Flush()
		loggerV2 = logrus.New()

		loggerV2.SetFormatter(&logrus.JSONFormatter{})
		loggerV2.SetOutput(logWriter)
		//set file name and line number
		loggerV2.SetReportCaller(true)
		loggerV2.AddHook(hook)

		// Set log level
		loggerV2.SetLevel(logrus.DebugLevel)

		// This will disable logging to stdout
		loggerV2.Out = io.Discard

	}
}

// =========== Exposed (public) Methods - can be called from external packages ============

// GetLogger returns the logrus logger object. It takes three input parameters.
// - logPrefix - it is a string used as Prefix on each log line
// - logPath - absolute path of the log file where the logs will be written
// - appName - It is app Name, from which service this function is being called to route the log to a specific Graylog stream.
func GetLoggerV2(logPrefix, logPath, appName string) *logrus.Logger {
	logInit.Do(func() {
		initializeLoggerV2(logPrefix, logPath, appName)
	})
	return loggerV2
}

func GrayLogHook(appName string) *graylog.GraylogHook {
	graylogAddr := os.Getenv("GRAYLOG_URL")

	hook := graylog.NewAsyncGraylogHook(graylogAddr, map[string]interface{}{"app": appName})
	return hook
}
