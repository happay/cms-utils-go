package session

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var once sync.Once

func GetSession(config *aws.Config) (sharedSession *session.Session, err error) {
	once.Do(func() {
		sharedSession, err = session.NewSession(config)
	})
	return
}
