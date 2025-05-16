package jobs

const emailJobKind string = "email_job"

type EmailJobArgs struct {
	Type        string `json:"type"`
	To          string `json:"to"`
	From        string `json:"from"`
	Subject     string `json:"subject"`
	TextVersion string `json:"text_version"`
	HtmlVersion string `json:"html_version"`
}

func (EmailJobArgs) Kind() string { return emailJobKind }
