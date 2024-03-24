package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Postmark struct {
	client  http.Client
	token   string
	baseUrl string
}

func NewPostmark(token string) Postmark {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	return Postmark{
		client,
		token,
		"https://api.postmarkapp.com",
	}
}

var _ mailClient = (*Postmark)(nil)

type mailBody struct {
	From     string `json:"From"`
	To       string `json:"To"`
	Subject  string `json:"Subject"`
	HtmlBody string `json:"HtmlBody"`
	TextBody string `json:"TextBody"`
}

// SendMail implements emailClient.
func (p *Postmark) SendMail(ctx context.Context, payload MailPayload) error {
	byt, err := json.Marshal(mailBody{
		From:     payload.From,
		To:       payload.To,
		Subject:  payload.Subject,
		HtmlBody: payload.HtmlBody,
		TextBody: payload.TextBody,
	})
	if err != nil {
		slog.Error("could not marshal email payload", "error", err)
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/email", p.baseUrl),
		bytes.NewBuffer(byt),
	)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", p.token)

	res, err := p.client.Do(req)
	if err != nil {
		slog.Error("could not send email", "error", err)
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		slog.Error("received unauthorized status code", "error", err)
		return ErrNotAuthorized
	}

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		slog.Error(
			"received non ok status code",
			"error",
			err,
			"status",
			res.StatusCode,
			"body",
			string(body),
		)
		return ErrCouldNotSend
	}

	return nil
}
