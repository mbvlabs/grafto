package awsses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mbvlabs/grafto/services"
)

type AwsSimpleEmailService struct {
	client  *ses.SES
	sender  string
	charSet string
}

// SendEmail implements mailClient.
func (a *AwsSimpleEmailService) SendEmail(
	ctx context.Context,
	payload services.EmailPayload,
) error {
	from := payload.From
	if payload.From == "" {
		from = a.sender
	}
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(payload.To),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(a.charSet),
					Data:    aws.String(payload.HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(a.charSet),
					Data:    aws.String(payload.TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(a.charSet),
				Data:    aws.String(payload.Subject),
			},
		},
		Source: aws.String(from),
	}

	_, err := a.client.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return err
	}

	return nil
}

func New() AwsSimpleEmailService {
	creds := credentials.NewEnvCredentials()
	conf := &aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: creds,
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		panic(err)
	}

	// TODO: accept these as arguments
	sender := "nopreply@grafto.com"
	charSet := "UTF-8"

	// Create an SES session.
	svc := ses.New(sess)
	return AwsSimpleEmailService{
		svc,
		sender,
		charSet,
	}
}

var _ services.EmailClient = (*AwsSimpleEmailService)(nil)
