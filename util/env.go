package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
)

// GetConfigValue get the environment value using the key.
// if not found, then fetches it from AWS Parameter Store
func GetConfigValue(key string) string {
	//TODO: include AWS Parameter Store
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	// /hpy/uat/cms-matrix/source_ami_id
	prefix := os.Getenv("SSM_PS_NP")
	region := os.Getenv("SSM_PS_RG")
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials("AKIASXMJT3QEGMB5HPUG", "CgNpj+taJzefqsda7FqGkGxEMmdRYwQn6mnWYOx3", ""),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return ""
	}

	paramterKey := prefix + "/" + key
	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(region))
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