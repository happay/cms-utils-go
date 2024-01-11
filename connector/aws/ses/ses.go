package ses

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/go-gomail/gomail"
	"github.com/happay/cms-utils-go/v2/connector/aws/cred"
	utilSession "github.com/happay/cms-utils-go/v2/connector/aws/session"
	"github.com/happay/cms-utils-go/v2/logger"
)

// ============ Constants =============

const (
	CharSet = "UTF-8" // The character encoding for the email.
)

// ============ Structs =============

type EmailDet struct {
	Sender      string   `json:"sender"`
	Recipient   []string `json:"recipient"`
	Subject     string   `json:"subject"`
	HtmlBody    string   `json:"htmlbody"`
	TextBody    string   `json:"textbody"`
	Attachments []string `json:"attachments"`
}

type EmailClient struct {
	cred.Cred
	session   *session.Session
	sesClient *ses.SES
}

// =========== Exposed (public) Methods - can be called from external packages ============

// SendEmail sends the email using the client and with the data specified in the EmailDet
func (emailClient *EmailClient) SendEmail(emailDet EmailDet) (err error) {
	if err = emailDet.CheckIfValidRecipients(); err != nil {
		return
	}
	// creates the email input
	emailInput := emailDet.createMailerInput()
	// sends the email
	_, err = emailClient.sesClient.SendEmail(emailInput)
	if err != nil {
		return
	}
	return
}

// SendEmailWithAttachments sends email with attachments
func (emailClient *EmailClient) SendEmailWithAttachments(emailDet EmailDet) (err error) {
	if err = emailDet.CheckIfValidRecipients(); err != nil {
		return
	}
	emailInput := emailDet.createRawInput()
	_, err = emailClient.sesClient.SendRawEmail(emailInput)
	return
}

func (emailClient *EmailClient) AddEmailIdToSuppressionList(emailId, reason string) (err error) {
	sesClient := sesv2.New(emailClient.session)
	suppressedDestInput := sesv2.PutSuppressedDestinationInput{
		EmailAddress: &emailId,
		Reason:       &reason,
	}
	_, err = sesClient.PutSuppressedDestination(&suppressedDestInput)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("unable to add email address to account suppression list: %s, err: %s",
			emailId, err))
		return
	}
	return
}

func (emailClient *EmailClient) GetEmailIdDetailsFromSuppressionList(emailId string) (result map[string]interface{}, err error) {
	sesClient := sesv2.New(emailClient.session)
	suppressedDestInput := sesv2.GetSuppressedDestinationInput{
		EmailAddress: &emailId,
	}
	suppressedOutput, err := sesClient.GetSuppressedDestination(&suppressedDestInput)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("unable to get email id from account suppression list: %s, err: %s",
			emailId, err))
		return
	}
	result = make(map[string]interface{})
	result["emailId"] = emailId
	result["reason"] = suppressedOutput.SuppressedDestination.Reason
	return
}

func (emailClient *EmailClient) GetListOfEmailIdsOnSuppressionList(startDate, endDate, nextToken string, reason []*string,
	pageSize int64) (result map[string]interface{}, err error) {
	sesClient := sesv2.New(emailClient.session)
	startDateTimeStamp, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("error while parsing start date: %s, err: %s", startDate, err))
		return
	}
	endDateTimeStamp, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("error while parsing end date: %s, err: %s", endDate, err))
		return
	}
	suppressionListDetails := sesv2.ListSuppressedDestinationsInput{
		StartDate: &startDateTimeStamp,
		EndDate:   &endDateTimeStamp,
		Reasons:   reason,
		PageSize:  &pageSize,
	}
	if nextToken != "" {
		suppressionListDetails.NextToken = &nextToken
	}
	suppressedListOutput, err := sesClient.ListSuppressedDestinations(&suppressionListDetails)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("unable to get the list of email ids from account suppression list - err: %s", err))
		return
	}
	result = make(map[string]interface{}, 0)
	response := make([]map[string]interface{}, 0)
	for _, suppressedOutput := range suppressedListOutput.SuppressedDestinationSummaries {
		details := make(map[string]interface{})
		details["emailId"] = suppressedOutput.EmailAddress
		details["reason"] = suppressedOutput.Reason
		response = append(response, details)
	}
	result["data"] = response
	result["next_token"] = suppressedListOutput.NextToken
	return
}

func (emailClient *EmailClient) RemoveEmailIdFromSuppressionList(emailId string) (err error) {
	sesClient := sesv2.New(emailClient.session)
	suppressedDestInput := sesv2.DeleteSuppressedDestinationInput{
		EmailAddress: &emailId,
	}
	_, err = sesClient.DeleteSuppressedDestination(&suppressedDestInput)
	if err != nil {
		logger.GetLoggerV3().Error(fmt.Sprintf("unable to delete email id from account suppression list: %s, err: %s",
			emailId, err))
		return
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

func (emailDet EmailDet) createMailerInput() (emailInput *ses.SendEmailInput) {
	recipients := make([]*string, len(emailDet.Recipient))
	for idx, recipientEmail := range emailDet.Recipient {
		recipients[idx] = aws.String(recipientEmail)
	}
	emailInput = &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
			CcAddresses: []*string{},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(emailDet.HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(emailDet.TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(emailDet.Subject),
			},
		},
		Source: aws.String(emailDet.Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}
	return
}

// CheckIfValidRecipients checks if the recepients have a valid email address, otherwise skip sending mails
// NOTE: While skipping, it is not raising any error for now, just logs the information
func (emailDet *EmailDet) CheckIfValidRecipients() (err error) {
	for _, recipientEmail := range emailDet.Recipient {
		if !govalidator.IsEmail(recipientEmail) {
			err = fmt.Errorf("invalid recpient mail: %s", recipientEmail)
		} else if strings.HasSuffix(recipientEmail, "abc.xyz.iin") { // email ends in this domain
			err = fmt.Errorf("invalid mail domain: %s", recipientEmail)
		} else if govalidator.IsNumeric(strings.Split(recipientEmail, "@")[0]) {
			err = fmt.Errorf("invalid recpient mail (local-part): %s", recipientEmail)
		}
		if err != nil {
			return
		}
	}
	return
}

// New creates new Email Client for SES
func (emailClient *EmailClient) New() (err error) {
	var config aws.Config
	config.Region = aws.String(emailClient.Region)
	if emailClient.Key != "" && emailClient.Secret != "" {
		config.Credentials = credentials.NewStaticCredentials(emailClient.Key, emailClient.Secret, "")
	}

	sess, err := utilSession.GetSession(&config)
	if err != nil {
		reason := fmt.Sprintf("error while creating the SES session : %s", err)
		err = errors.New(reason)
		return
	}
	emailClient.sesClient = ses.New(sess)
	return
}

func (emailDet EmailDet) createRawInput() (emailInput *ses.SendRawEmailInput) {
	recipients := make([]*string, len(emailDet.Recipient))
	for idx, recipientEmail := range emailDet.Recipient {
		recipients[idx] = aws.String(recipientEmail)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", emailDet.Sender)
	msg.SetHeader("To", strings.Join(emailDet.Recipient, ","))
	msg.SetHeader("Subject", emailDet.Subject)
	msg.SetBody("text/html", emailDet.HtmlBody)
	for _, fileLocation := range emailDet.Attachments {
		msg.Attach(fileLocation)
	}

	var emailRaw bytes.Buffer
	msg.WriteTo(&emailRaw)

	message := ses.RawMessage{Data: emailRaw.Bytes()}
	emailInput = &ses.SendRawEmailInput{
		Source:       aws.String(emailDet.Sender),
		Destinations: recipients,
		RawMessage:   &message,
	}
	return
}
