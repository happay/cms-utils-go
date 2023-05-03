package logger

import (
	cmsS3 "cms-utils-go/connector/aws/s3"
	"cms-utils-go/util"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

// ============ Internal(private) Methods - can only be called from inside this package ==============

var loggerV2 *logrus.Logger
var logWriterV2 io.Writer
var logInitV2 sync.Once

type LogInitializerObject struct {
	LogPrefix     string
	LogPath       string
	LogS3FilePath string
	BankServer    string
	BucketName    string
	Acl           string
	AppName       string
	S3Config      *aws.Config
}

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

func initializeLoggerWithLogRotation(logInitializerObject LogInitializerObject) {

	logPath := logInitializerObject.LogPath
	logS3FilePath := logInitializerObject.LogS3FilePath
	bankServer := logInitializerObject.BankServer
	bucketName := logInitializerObject.BucketName
	acl := logInitializerObject.Acl
	appName := logInitializerObject.AppName
	s3Config := logInitializerObject.S3Config

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
			panic(err)
		}
	}
	txnLogFile := logDir + "/" + logFileName + ".log"
	if _, err := os.OpenFile(txnLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,
		os.FileMode(FileWritePermissionMode)); err != nil {
		panic(err)
	} else {

		rl, _ := rotatelogs.New(txnLogFile+"."+"%Y-%m-%d-%H-%M-%S",
			rotatelogs.WithClock(rotatelogs.UTC),
			rotatelogs.WithRotationTime(4*time.Hour),
			rotatelogs.WithMaxAge(3*24*time.Hour),
			rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
				if e.Type() == rotatelogs.FileRotatedEventType {
					fre := e.(*rotatelogs.FileRotatedEvent)
					fmt.Println(fmt.Sprintf("Prev %s \nCurrent %s", fre.PreviousFile(), fre.CurrentFile()))
					UploadLogFile(fre.PreviousFile(), logFileName, logS3FilePath, bankServer, bucketName, acl, s3Config)

				}
			})),
		)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			os.Interrupt,
			os.Kill,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
			syscall.SIGQUIT)

		go func() {
			sig := <-ch
			fmt.Println("Force Rotating Log, Signal ", sig.String())
			rl.Rotate()
		}()

		var gw *gelf.Writer
		if gw, err = GrayLogWriter(); err != nil {
			panic(err)
		}

		//Grey Log hook
		grayLogHook := GrayLogHook(appName)
		defer grayLogHook.Flush()

		//File Writer hook
		localFileHook := lfshook.NewHook(lfshook.WriterMap{
			logrus.InfoLevel:  rl,
			logrus.ErrorLevel: rl,
		}, &logrus.JSONFormatter{})

		//create a log writer
		logWriterV2 = io.MultiWriter(rl, gw)

		loggerV2 = logrus.New()
		loggerV2.SetOutput(os.Stdout)

		//set the formatter of logs
		loggerV2.SetFormatter(&logrus.JSONFormatter{})

		//set file name and line number
		loggerV2.SetReportCaller(true)

		// Set log level
		loggerV2.SetLevel(logrus.DebugLevel)

		//Adding the hooks
		loggerV2.AddHook(grayLogHook)
		loggerV2.AddHook(localFileHook)

		// This will disable logging to stdout
		loggerV2.Out = io.Discard

	}
}

func UploadLogFile(currentFile, logFileName, logS3FilePath, bankServer, bucketName, acl string, s3Config *aws.Config) {

	fileNameSplit := strings.Split(currentFile, logFileName+".")
	var fileNameSplitTime string
	if len(fileNameSplit) > 1 {
		fileNameSplitTime = fileNameSplit[1]
	} else {
		return
	}

	fileNameTimeSplit := strings.Split(fileNameSplitTime, "-")
	year := fileNameTimeSplit[0]
	month := fileNameTimeSplit[1]
	day := fileNameTimeSplit[2]

	numStr, _ := util.GenerateRandomNumberOnlyString(8)
	fileName := filepath.Base(currentFile) + "-" + numStr

	fileNameS3Location := logS3FilePath + bankServer + "/" + year + "/" + month + "/" + day + "/"
	fileNameKey := fileNameS3Location + fileName + ".log"

	sess, err := session.NewSession(s3Config)
	if err != nil {
		reason := fmt.Sprintf("error while creating the S3 session : %s", err)
		err = errors.New(reason)
		logger.Print(err)
		return
	}

	s3Client := &cmsS3.S3Client{
		BucketName: bucketName,
	}
	s3Client.SetSession(sess)

	exists := IsFileExists(fileNameKey, bucketName, s3Config)
	if !exists {
		_, err = s3Client.UploadFile(currentFile, fileNameS3Location, fileName+".log", acl)
		if err != nil {
			log.Printf("Upload LogFile failed: %s %s", currentFile, err)
			return
		}
		log.Printf("Upload LogFile successful: %s", currentFile)
	}
	return
}

func IsFileExists(key, bucketName string, s3Config *aws.Config) (exists bool) {
	sess, err := session.NewSession(s3Config)
	if err != nil {
		reason := fmt.Sprintf("error while creating the S3 session : %s", err)
		err = errors.New(reason)
		logger.Print(err)
		return
	}
	s3svc := s3.New(sess)
	_, err = s3svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				exists = false
				return
			default:
				return
			}
		}
	}
	exists = true
	return
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

func GetLoggerWithLogRotation(logInitializerObject LogInitializerObject) *logrus.Logger {
	logInit.Do(func() {
		initializeLoggerWithLogRotation(logInitializerObject)
	})
	return loggerV2
}

func GetLogWriterWithLogRotation(logInitializerObject LogInitializerObject) *io.Writer {
	logInit.Do(func() {
		initializeLoggerWithLogRotation(logInitializerObject)
	})
	return &logWriterV2
}

func GrayLogHook(appName string) *graylog.GraylogHook {
	graylogAddr := os.Getenv("GRAYLOG_URL")

	hook := graylog.NewAsyncGraylogHook(graylogAddr, map[string]interface{}{"app": appName})
	return hook
}

func GrayLogWriter() (gw *gelf.Writer, err error) {
	graylogAddr := os.Getenv("GRAYLOG_URL")
	gw, err = gelf.NewWriter(graylogAddr)
	if err != nil {
		log.Printf("Failed to connect graylog server: error = %s", err)
		return
	}
	return
}
