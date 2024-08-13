package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/queue/jobs"
	"github.com/mbv-labs/grafto/views/emails"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type EmailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

type EmailClient interface {
	SendEmail(ctx context.Context, payload EmailPayload) error
}

type QueueClient interface {
	Insert(
		ctx context.Context,
		args river.JobArgs,
		opts *river.InsertOpts,
	) (*rivertype.JobRow, error)
}

type Email struct {
	cfg         config.Config
	client      EmailClient
	queueClient QueueClient
}

func NewEmailSvc(
	cfg config.Config,
	client EmailClient,
	queueClient QueueClient,
) Email {
	return Email{
		cfg,
		client,
		queueClient,
	}
}

func (e *Email) SendUserSignupWelcome(
	ctx context.Context,
	email string,
	activationTkn string,
	putOnQueue bool,
) error {
	newsletterEmail := emails.UserSignupWelcome{
		ConfirmationLink: fmt.Sprintf(
			"%s/verify-email?token=%s",
			e.cfg.GetFullDomain(),
			activationTkn,
		),
	}

	textVersion, err := newsletterEmail.GenerateTextVersion()
	if err != nil {
		slog.ErrorContext(
			ctx,
			"could not generate text version of UserSignupWelcomeEmail",
			"error",
			err,
		)
		return err
	}

	htmlVersion, err := newsletterEmail.GenerateHtmlVersion()
	if err != nil {
		slog.ErrorContext(
			ctx,
			"could not generate html version of UserSignupWelcomeEmail",
			"error",
			err,
		)
		return err
	}

	subject := "Grafto | Action Required"

	if putOnQueue {
		_, err := e.queueClient.Insert(ctx, jobs.EmailJobArgs{
			To:          email,
			From:        e.cfg.App.DefaultSenderSignature,
			Subject:     subject,
			TextVersion: textVersion,
			HtmlVersion: htmlVersion,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	}

	return e.client.SendEmail(ctx, EmailPayload{
		To:       email,
		From:     e.cfg.App.DefaultSenderSignature,
		Subject:  subject,
		HtmlBody: htmlVersion,
		TextBody: textVersion,
	})
}

func (e *Email) SendPasswordReset(
	ctx context.Context,
	email string,
	resetLink string,
	putOnQueue bool,
) error {
	newsletterEmail := emails.PasswordReset{
		ResetPasswordLink: resetLink,
	}

	textVersion, err := newsletterEmail.GenerateTextVersion()
	if err != nil {
		slog.ErrorContext(
			ctx,
			"could not generate text version of PasswordReset",
			"error",
			err,
		)
		return err
	}

	htmlVersion, err := newsletterEmail.GenerateHtmlVersion()
	if err != nil {
		slog.ErrorContext(
			ctx,
			"could not generate html version of PasswordReset",
			"error",
			err,
		)
		return err
	}

	subject := "Grafto | Reset Password Request"

	if putOnQueue {
		_, err := e.queueClient.Insert(ctx, jobs.EmailJobArgs{
			To:          email,
			From:        e.cfg.App.DefaultSenderSignature,
			Subject:     subject,
			TextVersion: textVersion,
			HtmlVersion: htmlVersion,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	}

	return e.client.SendEmail(ctx, EmailPayload{
		To:       email,
		From:     e.cfg.App.DefaultSenderSignature,
		Subject:  subject,
		HtmlBody: htmlVersion,
		TextBody: textVersion,
	})
}

func (e *Email) Send(
	ctx context.Context,
	to,
	from,
	subject,
	textVersion,
	htmlVersion string,
) error {
	return e.client.SendEmail(ctx, EmailPayload{
		To:       to,
		From:     from,
		Subject:  subject,
		HtmlBody: htmlVersion,
		TextBody: textVersion,
	})
}
