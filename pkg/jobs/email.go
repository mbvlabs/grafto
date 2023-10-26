package jobs

import (
	"context"
	"encoding/json"
)

const emailExecutorName = "email_Job_executor"

type EmailInstructions struct {
	To       string      `json:"to"`
	From     string      `json:"from"`
	TmplName string      `json:"tmpl_name"`
	Payload  interface{} `json:"payload"`
}

func CreateEmailJob(payload EmailInstructions) (Job, error) {
	instructions, err := json.Marshal(payload)
	if err != nil {
		return Job{}, err
	}

	return newJob(instructions, emailExecutorName), nil
}

type emailClient interface {
	Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error
}

type EmailExecutor struct {
	client emailClient
	name   string
}

func NewEmailJobExecutor(client emailClient) *EmailExecutor {
	return &EmailExecutor{
		client: client,
		name:   emailExecutorName,
	}
}

// Name implements Processor.
func (e *EmailExecutor) Name() string {
	return e.name
}

// Process implements Processor.
func (e *EmailExecutor) Process(ctx context.Context, msg []byte) error {
	var instructions EmailInstructions
	if err := json.Unmarshal(msg, &instructions); err != nil {
		return err
	}

	if err := e.client.Send(ctx,
		instructions.To, instructions.From, "Please confirm your email", instructions.TmplName,
		instructions.Payload); err != nil {
		return err
	}

	return nil
}

var _ Executor = (*EmailExecutor)(nil)
