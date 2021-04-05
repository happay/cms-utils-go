package connector

import (
	"fmt"

	"github.com/happay/cms-utils-go/connector/aws/cred"
	"github.com/happay/cms-utils-go/connector/aws/lambda"
	"github.com/happay/cms-utils-go/util"
)

// SlackMessage is used to send the message on the specified slack channel.
// It internally invokes a lambda function "hcm-slack-notify" to send the message.
// Since this lambda function is hosted in ap-south-1 region, there is no need to provide the region.
type SlackMessage struct {
	cred.Cred
	MessageText  string
	Channel      string
	ServiceName  string
	lambdaClient lambda.LambdaClient
}

// SlackMessageLambdaFuncName ...
const SlackMessageLambdaFuncName = "hcm-slack-notify"

// HCMSlackNotifyRegion ...
const HCMSlackNotifyRegion = "ap-south-1"

func (sm SlackMessage) Send() (err error) {
	if err = sm.new(); err != nil {
		return
	}

	requestData := util.PropertyMap{}
	requestData["text"] = sm.MessageText
	requestData["channel"] = sm.Channel
	requestData["service_name"] = sm.ServiceName

	result, err := sm.lambdaClient.InvokeAWSLambdaFunc(SlackMessageLambdaFuncName, requestData)
	if err != nil {
		err = fmt.Errorf("error while invoking lambda function: %s, result: %s", err, result.GoString())
		return
	}

	return
}

func (sm SlackMessage) new() (err error) {
	if sm.Cred.Key == "" || sm.Cred.Secret == "" {
		err = fmt.Errorf("Empty aws key or secret")
		return
	}

	sm.Cred.Region = HCMSlackNotifyRegion

	if err = sm.lambdaClient.New(); err != nil {
		err = fmt.Errorf("error while creating lambda client: %s", err)
		return
	}

	return
}
