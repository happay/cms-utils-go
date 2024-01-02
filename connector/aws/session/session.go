package session

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/happay/cms-utils-go/v2/logger"
)

var sharedSession *session.Session
var err error
var expiration time.Time

func GetSession(config *aws.Config) (*session.Session, error) {
	if sharedSession != nil && time.Now().Before(expiration) {
		return sharedSession, nil
	}
	// Recreate the session
	sharedSession, err = session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")), // Replace with your AWS region
	})
	if err != nil {
		logger.GetLoggerV3().Error("Error while creating an AWS session" + err.Error())
		return sharedSession, err
	}
	expiration = time.Now().Add(1 * time.Hour) // Set expiration time to 1 hour
	return sharedSession, err
}
