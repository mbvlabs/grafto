package job

import (
	"context"
	"encoding/json"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/google/uuid"
)

const emailProcessorName = "email_job_processor"

type EmailJobMsg struct {
	To       string `json:"to"`
	From     string `json:"from"`
	HtmlTmpl string `json:"html_tmpl"`
	TextTmpl string `json:"text_tmpl"`
}

func CreateEmailJobMsg(payload EmailJobMsg) (Job, error) {
	instructions, err := json.Marshal(payload)
	if err != nil {
		return Job{}, err
	}

	return Job{
		ID:           uuid.New(),
		Instructions: instructions,
		processor:    emailProcessorName,
	}, nil
}

type EmailClient interface {
	Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error
}

type EmailJobProcessor struct {
	client EmailClient
	name   string
}

// Name implements Processor.
func (e *EmailJobProcessor) Name() string {
	return e.name
}

// Process implements Processor.
func (e *EmailJobProcessor) Process(ctx context.Context, msg []byte) error {
	var message EmailJobMsg
	if err := json.Unmarshal(msg, &message); err != nil {
		return err
	}

	telemetry.Logger.Info("processing email job", "msg", message)

	return nil
}

func NewEmailJobProcessor(client EmailClient) *EmailJobProcessor {
	return &EmailJobProcessor{
		client: client,
		name:   "email_job_processor",
	}
}

var _ Processor = (*EmailJobProcessor)(nil)
