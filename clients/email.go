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

// TODO: complete this flow
type Unsubscribe struct {
	Email string
	Link  string
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

func (e Email) SendTransaction(
	ctx context.Context,
	payload EmailPayload,
) error {
	return e.send(ctx, payload, Unsubscribe{})
}

func (e Email) SendMarketing(
	ctx context.Context,
	payload EmailPayload,
	unsub Unsubscribe,
) error {
	return e.send(ctx, payload, unsub)
}

func (e Email) send(
	ctx context.Context,
	payload EmailPayload,
	unsub Unsubscribe,
) error {
	from := payload.From
	if payload.From == "" {
		from = defaultSender
	}

	baseEmail := &jwEmail.Email{
		To:      []string{payload.To},
		From:    from,
		Subject: payload.Subject,
		Text:    []byte(payload.TextBody),
		HTML:    []byte(payload.HtmlBody),
	}

	if unsub.Email != "" && unsub.Link != "" {
		baseEmail.Headers = textproto.MIMEHeader{
			"List-Unsubscribe": []string{
				"<" + unsub.Link + ">, <mailto:" + unsub.Email + ">",
			},
			"List-Unsubscribe-Post": []string{"List-Unsubscribe=One-Click"},
		}
	}

	rawMessage, err := baseEmail.Bytes()
	if err != nil {
		return err
	}

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
