package queue

import (
	"context"
	"encoding/json"
)

const emailExecutorName = "email"

type emailSender interface {
	Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error
}

type EmailExecutor struct {
	client emailSender
	name   string
}

func NewEmailExecutor(client emailSender) *EmailExecutor {
	return &EmailExecutor{
		client: client,
		name:   emailExecutorName,
	}
}

var _ Executor = (*EmailExecutor)(nil)

// Name implements Executor.
func (e *EmailExecutor) Name() string {
	return e.name
}

// Process implements Executor.
func (e *EmailExecutor) Process(ctx context.Context, msg []byte) error {
	var instructions EmailInstructions
	if err := json.Unmarshal(msg, &instructions); err != nil {
		return err
	}

	if err := e.client.Send(
		ctx, instructions.To, instructions.From, instructions.Subject, instructions.TmplName,
		instructions.Payload); err != nil {
		return err
	}

	return nil
}

type EmailInstructions struct {
	To       string      `json:"to"`
	From     string      `json:"from"`
	Subject  string      `json:"subject"`
	TmplName string      `json:"tmpl_name"`
	Payload  interface{} `json:"payload"`
}

func CreateEmailJob(payload EmailInstructions) (*Job, error) {
	instructions, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return createJob(instructions, emailExecutorName), nil
}