package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
)

var Region = os.Getenv("SSM_PS_RG")

var sess, _ = session.NewSessionWithOptions(session.Options{
	Config: aws.Config{Region: aws.String(Region),
	},
	SharedConfigState: session.SharedConfigEnable,
})

var ssmsvc = ssm.New(sess, aws.NewConfig().WithRegion(Region))

// GetConfigValue get the environment value using the key.
// if not found, then fetches it from AWS Parameter Store
func GetConfigValue(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	prefix := os.Getenv("SSM_PS_NP")
	paramterKey := prefix + "/" + key

	withDecryption := false
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &paramterKey,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return ""
	}

	value = *param.Parameter.Value
	return value
}
