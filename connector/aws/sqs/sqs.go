package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
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
	Cred
	Url    string
	session *session.Session
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

	queue := sqs.New(qClient.session)
	sqsResponse, err := queue.SendMessage(queueMessage)
	if err != nil {
		return err
	}
	q.MessageId = *sqsResponse.MessageId
	return
}

// Dequeue ...
func (qClient *QueueClient) Dequeue(numOfPackets ...int64) (queueMessageList []QueueMessage, err error) {
	var result *sqs.ReceiveMessageOutput
	queue := sqs.New(qClient.session)

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

	result, err = queue.ReceiveMessage(receiveMessage)
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

// Delete Remove the message from the queue after the processing of the event
// Event is removed from the queue even on failure but enqueued again with the delay
func (qClient *QueueClient) Delete(msg QueueMessage) (err error) {
	queue := sqs.New(qClient.session)
	_, err = queue.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &qClient.Url,
		ReceiptHandle: &msg.ReceiptHandle,
	})
	return
}

// New creates QueueClinet optArgs [AwsKey, AwsSecret]
func InitQueue(url string, optArgs ...string) (*QueueClient, error) {
	var err error
	awsConfig := &aws.Config{
		Region: aws.String(Region),
	}
	if len(optArgs) == 2 {
		awsKey := optArgs[0]
		awsSecret := optArgs[1]
		awsConfig.Credentials = credentials.NewStaticCredentials(awsKey, awsSecret, "")
	} else {
		err = fmt.Errorf("opt Args length in invalid Formate AwsKey, AwsSecret")
		return nil, err
	}
	//Credentials:
	sess, err := session.NewSession(awsConfig)
	queueClient := &QueueClient{
		Url:     url,
		session: sess,
	}
	return queueClient, nil
}

// =========================== Private Functions =================================
