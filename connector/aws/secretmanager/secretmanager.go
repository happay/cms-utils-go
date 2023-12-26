package secretmanager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	utilSession "github.com/happay/cms-utils-go/v2/connector/aws/session"
	"github.com/happay/cms-utils-go/v2/util"
)

var Region = util.GetConfigValue("SSM_PS_RG")

// GetValueFromSecretManager gets the value from the AWS secret manager using the input key
func GetValueFromSecretManager(key string) (result *secretsmanager.GetSecretValueOutput, err error) {
	sess, err := utilSession.GetSession(secretManagerConfig)
	if err != nil {
		return
	}
	svc := secretsmanager.New(sess)
	result, err = svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(key),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		return
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============
var secretManagerConfig = &aws.Config{
	Region: aws.String(Region),
}
