package clients

import (
	"context"
	"net/textproto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	jwEmail "github.com/jordan-wright/email"
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

func (e Email) Send(
	ctx context.Context,
	payload EmailPayload,
) error {
	from := payload.From
	if payload.From == "" {
		from = defaultSender
	}

	// Create the unsubscribe URL with a token to identify the user
	unsubscribeURL := "https://mbvlabs.com/unsubscribe?token=" + "mkdasmdldsa"
	unsubscribeEmail := "unsubscribe@mbvlabs.com"

	baseEmail := &jwEmail.Email{
		To:      []string{payload.To},
		From:    from,
		Subject: payload.Subject,
		Text:    []byte(payload.TextBody),
		HTML:    []byte(payload.HtmlBody),
		Headers: textproto.MIMEHeader{
			"List-Unsubscribe": []string{
				"<" + unsubscribeURL + ">, <mailto:" + unsubscribeEmail + ">",
			},
			"List-Unsubscribe-Post": []string{"List-Unsubscribe=One-Click"},
		},
	}

	// ses.AddHeaderAction

	// input := &ses.SendEmailInput{
	// 	Destination: &ses.Destination{
	// 		CcAddresses: []*string{},
	// 		ToAddresses: []*string{
	// 			aws.String(payload.To),
	// 		},
	// 	},
	// 	Message: &ses.Message{
	// 		Body: &ses.Body{
	// 			Html: &ses.Content{
	// 				Charset: aws.String(charSet),
	// 				Data:    aws.String(payload.HtmlBody),
	// 			},
	// 			Text: &ses.Content{
	// 				Charset: aws.String(charSet),
	// 				Data:    aws.String(payload.TextBody),
	// 			},
	// 		},
	// 		Subject: &ses.Content{
	// 			Charset: aws.String(charSet),
	// 			Data:    aws.String(payload.Subject),
	// 		},
	// 	},
	// 	Source: aws.String(from),
	// }

	// req, _ := m.client.SendEmailRequest(input)
	// req.HTTPRequest.
	// if err != nil {
	// 	return err
	// }

	// Encode the email
	rawMessage, err := baseEmail.Bytes()
	if err != nil {
		return err
	}

	// Send the raw email with context
	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: rawMessage,
		},
	}

	_, err = e.client.SendRawEmailWithContext(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
