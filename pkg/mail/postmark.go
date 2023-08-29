package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
}

// SendMail implements emailClient.
func (p *Postmark) SendMail(ctx context.Context, payload MailPayload) error {
	byt, err := json.Marshal(mailBody{
		From:     payload.From,
		To:       payload.To,
		Subject:  payload.Subject,
		HtmlBody: payload.HtmlBody,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/email", p.baseUrl), bytes.NewBuffer(byt))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", p.token)

	res, err := p.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return ErrNotAuthorized
	}

	if res.StatusCode != http.StatusOK {
		return ErrCouldNotSend
	}

	return nil
}
