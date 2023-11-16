package queue

import (
	"context"
	"encoding/json"
)

const emailExecutorName string = "email_executor"

type emailSender interface {
	Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error
}

type emailExecutor struct {
	client emailSender
	name   string
}

func NewEmailExecutor(client emailSender) *emailExecutor {
	return &emailExecutor{
		client,
		emailExecutorName,
	}
}

var _ Executor = (*emailExecutor)(nil)

// Name implements Executor.
func (e *emailExecutor) Name() string {
	return e.name
}

// Process implements Executor.
func (e *emailExecutor) process(ctx context.Context, msg []byte) error {
	var instructions EmailJob
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

type EmailJob struct {
	To       string      `json:"to"`
	From     string      `json:"from"`
	Subject  string      `json:"subject"`
	TmplName string      `json:"tmpl_name"`
	Payload  interface{} `json:"payload"`
}

func NewEmailJob(payload EmailJob) (*Job, error) {
	instructions, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	jobInstructions := jobInstructions{
		instructions: instructions,
		executor:     emailExecutorName,
	}

	return newJob(jobInstructions)
}
