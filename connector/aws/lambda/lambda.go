package lambda

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/happay/cms-utils-go/v2/connector/aws/cred"
	"github.com/happay/cms-utils-go/v2/util"
)

// LambdaClient ...
type LambdaClient struct {
	cred.Cred
	lambda *lambda.Lambda
}

// InvokeAWSLambdaFunc ...
func (lambdaClient *LambdaClient) InvokeAWSLambdaFunc(functionName string, requestData util.PropertyMap) (result *lambda.InvokeOutput, err error) {
	payload, err := json.Marshal(requestData)
	if err != nil {
		err = fmt.Errorf("error while marshalling request data: %s", err)
		return
	}

	if lambdaClient.lambda == nil {
		if err = lambdaClient.New(); err != nil {
			return
		}
	}

	result, err = lambdaClient.lambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		err = fmt.Errorf("error while invoking lambda func: %s", err)
		return
	}

	return
}

// New creates new client for Lambda.
func (lambdaClient *LambdaClient) New() (err error) {
	var config aws.Config
	config.Region = aws.String(lambdaClient.Region)
	if lambdaClient.Key != "" && lambdaClient.Secret != "" {
		config.Credentials = credentials.NewStaticCredentials(lambdaClient.Key, lambdaClient.Secret, "")
	}

	sess, err := session.NewSession(&config)
	if err != nil {
		reason := fmt.Sprintf("error while creating the lambda session : %s", err)
		err = errors.New(reason)
		return
	}
	lambdaClient.lambda = lambda.New(sess)
	return
}
