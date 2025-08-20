package clients

import (
	"context"
	"net/textproto"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	jwEmail "github.com/jordan-wright/email"
	appConfig "github.com/mbvlabs/grafto/config"
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

var defaultSender = appConfig.Cfg.App.DefaultSenderSignature

type Email struct {
	client *sesv2.Client
}

func NewEmail(ctx context.Context) Email {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		panic(err)
	}

	client := sesv2.NewFromConfig(cfg)

	return Email{
		client,
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

	input := &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: rawMessage,
			},
		},
	}

	_, err = e.client.SendEmail(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
