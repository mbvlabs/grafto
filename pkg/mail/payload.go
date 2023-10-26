package mail

type MailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

type ConfirmPassword struct {
	Token string
}

type WeeklyStatusReport struct {
	NewUsers int
}
