package connector

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"os"
	"strings"
)

// ============ Constants =============

const (
	CharSet        = "UTF-8" // The character encoding for the email.
	NoReplyEmailId = "no-reply@happay.in"
)

const MatrixGroupEmail = "matrix-defenders@googlegroups.com" // matrix developer group mail

var AwsSesRegion = os.Getenv("SSM_PS_RG")
//var AwsSesAccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
//var AwsSesSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

// ============ Structs =============

type EmailDet struct {
	Sender    string   `json:"sender"`
	Recipient []string `json:"recipient"`
	Subject   string   `json:"subject"`
	HtmlBody  string   `json:"htmlbody"`
	TextBody  string   `json:"textbody"`
}

// =========== Exposed (public) Methods - can be called from external packages ============

func (emailDet EmailDet) SendEmail() (err error) {
	if err = emailDet.CheckIfValidRecipients(); err != nil {
		return
	}

	sess, err := session.NewSession(sesConfig)
	if err != nil {
		reason := fmt.Sprintf("error while creating the SES session : %s", err)
		err = errors.New(reason)
		return
	}
	// creates a new AWS SES session
	emailProvider := ses.New(sess)
	// creates the email input
	emailInput := emailDet.createMailerInput()
	// sends the email
	_, err = emailProvider.SendEmail(emailInput)
	if err != nil {
		return
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

var sesConfig = &aws.Config{
	Region:      aws.String(AwsSesRegion),
}

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
// TODO: Should we return error to the calling function also
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