package mail

type MailPayload struct {
	To       string
	From     string
	Subject  string
	HtmlBody string
	TextBody string
}

type ForgottonPassword struct {
	SiteUrl string
	Token   string
}

type ConfirmPassword struct {
	Token string
}
