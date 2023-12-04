package queue

import (
	"context"
	"encoding/json"
	"time"
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

// Name implements Executor.
func (e *emailExecutor) Name() string {
	return e.name
}

// Process implements Executor.
func (e *emailExecutor) process(ctx context.Context, msg []byte) error {
	var instructions emailJob
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

var _ Executor = (*emailExecutor)(nil)

type emailJob struct {
	To       string      `json:"to"`
	From     string      `json:"from"`
	Subject  string      `json:"subject"`
	TmplName string      `json:"tmpl_name"`
	Payload  interface{} `json:"payload"`
	job      *job
}

var _ jobCreator = (*emailJob)(nil)

func NewEmailJob(to, from, subject, tmplName string, payload any) (*emailJob, error) {
	ej := emailJob{
		To:       to,
		From:     from,
		Subject:  subject,
		TmplName: tmplName,
		Payload:  payload,
	}

	instructions, err := json.Marshal(ej)
	if err != nil {
		return nil, err
	}

	job, err := newJob(jobInstructions{
		instructions: instructions,
		executor:     emailExecutorName,
	})
	if err != nil {
		return nil, err
	}

	ej.job = job

	return &ej, nil
}

func (e *emailJob) Schedule(t time.Time) error {
	// check if t is in the past
	e.job.scheduledFor = t
	return nil
}

func (e *emailJob) build() *job {
	return e.job
}
