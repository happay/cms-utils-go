package sqs

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/happay/cms-utils-go/v2/connector/aws/cred"
	utilSession "github.com/happay/cms-utils-go/v2/connector/aws/session"
	"github.com/happay/cms-utils-go/v2/logger"
)

var Region = os.Getenv("SSM_PS_RG")

type QueueMessage struct {
	// Message that need to be added to the queue
	Message string

	// Attributes of the enqueueing packet
	Attributes map[string]string

	// Delay Number of seconds packet needs to be delayed
	Delay int64

	// MessageId ack id of the packet once enqueued
	MessageId string

	//
	ReceiptHandle string
}

type QueueClient struct {
	cred.Cred
	Url     string
	session *session.Session
	sqs     *sqs.SQS
}

// Enqueue ...
func (qClient *QueueClient) Enqueue(q QueueMessage) (err error) {
	attributes := make(map[string]*sqs.MessageAttributeValue, 0)
	for key, value := range q.Attributes {
		attributes[key] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	queueMessage := &sqs.SendMessageInput{
		DelaySeconds:      aws.Int64(q.Delay),
		MessageBody:       aws.String(q.Message),
		MessageAttributes: attributes,
		QueueUrl:          &qClient.Url,
	}
	sqsResponse, err := qClient.sqs.SendMessage(queueMessage)
	if err != nil {
		return err
	}
	q.MessageId = *sqsResponse.MessageId
	return
}

// Dequeue ...
func (qClient *QueueClient) Dequeue(numOfPackets ...int64) (queueMessageList []QueueMessage, err error) {
	var result *sqs.ReceiveMessageOutput
	var size int64
	if len(numOfPackets) == 0 {
		size = 1
	} else {
		size = numOfPackets[0]
	}

	receiveMessage := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &qClient.Url,
		MaxNumberOfMessages: aws.Int64(size), // Read 1- messages at a time
		VisibilityTimeout:   aws.Int64(300),  // 5 Mins
		WaitTimeSeconds:     aws.Int64(3),    // wait for 3 seconds
	}

	result, err = qClient.sqs.ReceiveMessage(receiveMessage)
	if err != nil {
		return
	}

	if len(result.Messages) == 0 {
		//common.GetLogger().Print("no messages found")
		//time.Sleep(500 * time.Millisecond)
		return
	}

	queueMessageList = make([]QueueMessage, 0)
	for _, message := range result.Messages {
		tempQueueObj := QueueMessage{}
		tempQueueObj.Message = *message.Body

		attributes := make(map[string]string, 0)
		for attrKey, attrValue := range message.MessageAttributes {
			attributes[attrKey] = *attrValue.StringValue
		}
		tempQueueObj.Attributes = attributes
		tempQueueObj.ReceiptHandle = *message.ReceiptHandle
		queueMessageList = append(queueMessageList, tempQueueObj)
	}
	return
}

// Delete removes the message from the queue after the processing of the event
// Event is removed from the queue even on failure but enqueued again with the delay
func (qClient *QueueClient) Delete(msg QueueMessage) (err error) {
	queue := sqs.New(qClient.session)
	_, err = queue.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &qClient.Url,
		ReceiptHandle: &msg.ReceiptHandle,
	})
	return
}

// InitQueue creates QueueClinet optArgs [AwsKey, AwsSecret]
func InitQueue(url string, optArgs ...string) (queueClient *QueueClient, err error) {
	awsConfig := &aws.Config{
		Region: aws.String(Region),
	}
	if len(optArgs) == 2 {
		awsKey := optArgs[0]
		awsSecret := optArgs[1]
		awsConfig.Credentials = credentials.NewStaticCredentials(awsKey, awsSecret, "")
	}
	sess, err := utilSession.GetSession(awsConfig)
	if err != nil {
		logger.GetLoggerV3().Error("Error creating session for SQS" + err.Error())
	}
	sqs := sqs.New(sess)
	queueClient = &QueueClient{
		Url:     url,
		session: sess,
		sqs:     sqs,
	}
	return
}

func (qClient *QueueClient) New() (err error) {
	awsConfig := &aws.Config{
		Region:     aws.String(qClient.Region),
		MaxRetries: aws.Int(10),
	}
	if qClient.Key != "" && qClient.Secret != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(qClient.Key, qClient.Secret, "")
	}
	qClient.session, err = utilSession.GetSession(awsConfig)
	if err != nil {
		logger.GetLoggerV3().Error("Error creating session for SQS" + err.Error())
	}
	qClient.sqs = sqs.New(qClient.session)
	return
}

func (qClient *QueueClient) EnqueueInFifo(q QueueMessage) (err error) {
	attributes := make(map[string]*sqs.MessageAttributeValue, 0)
	for key, value := range q.Attributes {
		attributes[key] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	queueMessage := &sqs.SendMessageInput{
		DelaySeconds:   aws.Int64(q.Delay),
		MessageBody:    aws.String(q.Message),
		QueueUrl:       &qClient.Url,
		MessageGroupId: &q.MessageId,
	}
	sqsResponse, err := qClient.sqs.SendMessage(queueMessage)
	if err != nil {
		return err
	}
	q.MessageId = *sqsResponse.MessageId
	return
}

func (qClient *QueueClient) NextMessages(MaxNumberOfMessages int64, WaitTimeSeconds int64) ([]*sqs.Message, error) {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(qClient.Url),
		MaxNumberOfMessages: aws.Int64(MaxNumberOfMessages),
		WaitTimeSeconds:     aws.Int64(WaitTimeSeconds),
	}

	resp, err := qClient.sqs.ReceiveMessage(params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

// =========================== Private Functions =================================
