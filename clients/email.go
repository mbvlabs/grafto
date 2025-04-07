package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mbvlabs/grafto/config"
)

type EmailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

const (
	charSet   = "UTF-8"
	awsRegion = "eu-central-1"
)

var defaultSender = config.Cfg.DefaultSenderSignature

type Email struct {
	client *ses.SES
}

func NewEmail() Email {
	creds := credentials.NewEnvCredentials()
	conf := &aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: creds,
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		panic(err)
	}

	ses := ses.New(sess)

	return Email{
		ses,
	}
}

func (m Email) Send(
	ctx context.Context,
	payload EmailPayload,
) error {
	from := payload.From
	if payload.From == "" {
		from = defaultSender
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
					Charset: aws.String(charSet),
					Data:    aws.String(payload.HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(payload.TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(payload.Subject),
			},
		},
		Source: aws.String(from),
	}

	_, err := m.client.SendEmail(input)
	if err != nil {
		//nolint:errorlint //todo
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(
					ses.ErrCodeMailFromDomainNotVerifiedException,
					aerr.Error(),
				)
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(
					ses.ErrCodeConfigurationSetDoesNotExistException,
					aerr.Error(),
				)
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
