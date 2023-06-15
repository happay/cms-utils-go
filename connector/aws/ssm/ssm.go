package ssm

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
)

var AWS_SSM_REGION = os.Getenv("AWS_REGION")
var AWS_ACCESS_KEY_ID = os.Getenv("AWS_ACCESS_KEY_ID")
var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
var SSM_PREFIX = os.Getenv("PARAMETER_STORE_PREFIX")

const EmptyString = ""

var ssmConfig = &aws.Config{
	Region:      aws.String(AWS_SSM_REGION),
	Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, "")}

func GetParametersByPath(path string) (params []*ssm.Parameter, err error) {
	// Error handling
	if path == EmptyString {
		err = fmt.Errorf("empty prefix path provided for getting parameters : %s", err)
		return
	}
	// Load AWS credentials and region from environment variables, shared config, or EC2 instance metadata
	ssmSession, err := session.NewSession(ssmConfig)
	if err != nil {
		err = fmt.Errorf("error while creating session for SSM: %s", err)
		return
	}
	ssmClient := ssm.New(ssmSession)

	// Get the parameters by path from Parameter Store
	err = ssmClient.GetParametersByPathPages(&ssm.GetParametersByPathInput{
		Path:           &path,
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}, func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
		params = append(params, page.Parameters...)
		return !lastPage
	})
	if err != nil {
		err = fmt.Errorf("%s | Error in retrieving parameter from parameter store", err)
		return
	}
	return
}
func SetEnvironmentVariable() (err error) {
	params, err := GetParametersByPath(SSM_PREFIX)
	if err != nil {
		err = fmt.Errorf("failed to get parameter from parameter store. Unable to set environment variables: %s", err)
		return
	}
	for _, param := range params {
		parameterName := aws.StringValue(param.Name)[len(SSM_PREFIX)+1:]
		err = os.Setenv(parameterName, *param.Value)
		if err != nil {
			err = fmt.Errorf("error while setting OS ENV variable: %s", err)
			return
		}
	}
	return
}
