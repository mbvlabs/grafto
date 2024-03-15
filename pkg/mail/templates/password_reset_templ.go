// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"fmt"
	"text/template"

	"github.com/vanng822/go-premailer/premailer"
)

const passwordResetTmplName = "password_reset_email"

type PasswordResetMail struct {
	ResetPasswordLink string
	UnsubscribeLink   string
}

var _ MailTemplateHandler = (*PasswordResetMail)(nil)

func (p PasswordResetMail) GenerateTextVersion() (string, error) {
	textFile, err := template.ParseFS(textTemplates, fmt.Sprintf("%s.txt", passwordResetTmplName))
	if err != nil {
		return "", err
	}

	var textBody bytes.Buffer
	if err := textFile.Execute(&textBody, PasswordResetMail{}); err != nil {
		return "", err
	}

	return textBody.String(), nil
}

func (p PasswordResetMail) GenerateHtmlVersion() (string, error) {
	var html bytes.Buffer
	if err := p.template().Render(context.Background(), &html); err != nil {
		return "", err
	}

	premailer, err := premailer.NewPremailerFromString(html.String(), premailer.NewOptions())
	if err != nil {
		return "", err
	}

	inlineHtml, err := premailer.Transform()
	if err != nil {
		return "", err
	}

	return inlineHtml, nil
}

func (p PasswordResetMail) Render(ctx context.Context, w io.Writer) error {
	return p.template().Render(ctx, w)
}

func (n PasswordResetMail) template() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html xmlns=\"http://www.w3.org/1999/xhtml\"><head><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><meta name=\"x-apple-disable-message-reformatting\"><meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\"><meta name=\"color-scheme\" content=\"light dark\"><meta name=\"supported-color-schemes\" content=\"light dark\"><title></title><style type=\"text/css\" rel=\"stylesheet\" media=\"all\">\n    /* Base ------------------------------ */\n    \n    @import url(\"https://fonts.googleapis.com/css?family=Nunito+Sans:400,700&display=swap\");\n    body {\n      width: 100% !important;\n      height: 100%;\n      margin: 0;\n      -webkit-text-size-adjust: none;\n    }\n    \n    a {\n      color: #3869D4;\n    }\n    \n    a img {\n      border: none;\n    }\n    \n    td {\n      word-break: break-word;\n    }\n    \n    .preheader {\n      display: none !important;\n      visibility: hidden;\n      mso-hide: all;\n      font-size: 1px;\n      line-height: 1px;\n      max-height: 0;\n      max-width: 0;\n      opacity: 0;\n      overflow: hidden;\n    }\n    /* Type ------------------------------ */\n    \n    body,\n    td,\n    th {\n      font-family: \"Nunito Sans\", Helvetica, Arial, sans-serif;\n    }\n    \n    h1 {\n      margin-top: 0;\n      color: #333333;\n      font-size: 22px;\n      font-weight: bold;\n      text-align: left;\n    }\n    \n    h2 {\n      margin-top: 0;\n      color: #333333;\n      font-size: 16px;\n      font-weight: bold;\n      text-align: left;\n    }\n    \n    h3 {\n      margin-top: 0;\n      color: #333333;\n      font-size: 14px;\n      font-weight: bold;\n      text-align: left;\n    }\n    \n    td,\n    th {\n      font-size: 16px;\n    }\n    \n    p,\n    ul,\n    ol,\n    blockquote {\n      margin: .4em 0 1.1875em;\n      font-size: 16px;\n      line-height: 1.625;\n    }\n    \n    p.sub {\n      font-size: 13px;\n    }\n    /* Utilities ------------------------------ */\n    \n    .align-right {\n      text-align: right;\n    }\n    \n    .align-left {\n      text-align: left;\n    }\n    \n    .align-center {\n      text-align: center;\n    }\n    \n    .u-margin-bottom-none {\n      margin-bottom: 0;\n    }\n    /* Buttons ------------------------------ */\n    \n    .button {\n      background-color: #3869D4;\n      border-top: 10px solid #3869D4;\n      border-right: 18px solid #3869D4;\n      border-bottom: 10px solid #3869D4;\n      border-left: 18px solid #3869D4;\n      display: inline-block;\n      color: #FFF;\n      text-decoration: none;\n      border-radius: 3px;\n      box-shadow: 0 2px 3px rgba(0, 0, 0, 0.16);\n      -webkit-text-size-adjust: none;\n      box-sizing: border-box;\n    }\n    \n    .button--green {\n      background-color: #22BC66;\n      border-top: 10px solid #22BC66;\n      border-right: 18px solid #22BC66;\n      border-bottom: 10px solid #22BC66;\n      border-left: 18px solid #22BC66;\n    }\n    \n    .button--red {\n      background-color: #FF6136;\n      border-top: 10px solid #FF6136;\n      border-right: 18px solid #FF6136;\n      border-bottom: 10px solid #FF6136;\n      border-left: 18px solid #FF6136;\n    }\n    \n    @media only screen and (max-width: 500px) {\n      .button {\n        width: 100% !important;\n        text-align: center !important;\n      }\n    }\n    /* Attribute list ------------------------------ */\n    \n    .attributes {\n      margin: 0 0 21px;\n    }\n    \n    .attributes_content {\n      background-color: #F4F4F7;\n      padding: 16px;\n    }\n    \n    .attributes_item {\n      padding: 0;\n    }\n    /* Related Items ------------------------------ */\n    \n    .related {\n      width: 100%;\n      margin: 0;\n      padding: 25px 0 0 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n    }\n    \n    .related_item {\n      padding: 10px 0;\n      color: #CBCCCF;\n      font-size: 15px;\n      line-height: 18px;\n    }\n    \n    .related_item-title {\n      display: block;\n      margin: .5em 0 0;\n    }\n    \n    .related_item-thumb {\n      display: block;\n      padding-bottom: 10px;\n    }\n    \n    .related_heading {\n      border-top: 1px solid #CBCCCF;\n      text-align: center;\n      padding: 25px 0 10px;\n    }\n    /* Discount Code ------------------------------ */\n    \n    .discount {\n      width: 100%;\n      margin: 0;\n      padding: 24px;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n      background-color: #F4F4F7;\n      border: 2px dashed #CBCCCF;\n    }\n    \n    .discount_heading {\n      text-align: center;\n    }\n    \n    .discount_body {\n      text-align: center;\n      font-size: 15px;\n    }\n    /* Social Icons ------------------------------ */\n    \n    .social {\n      width: auto;\n    }\n    \n    .social td {\n      padding: 0;\n      width: auto;\n    }\n    \n    .social_icon {\n      height: 20px;\n      margin: 0 8px 10px 8px;\n      padding: 0;\n    }\n    /* Data table ------------------------------ */\n    \n    .purchase {\n      width: 100%;\n      margin: 0;\n      padding: 35px 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n    }\n    \n    .purchase_content {\n      width: 100%;\n      margin: 0;\n      padding: 25px 0 0 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n    }\n    \n    .purchase_item {\n      padding: 10px 0;\n      color: #51545E;\n      font-size: 15px;\n      line-height: 18px;\n    }\n    \n    .purchase_heading {\n      padding-bottom: 8px;\n      border-bottom: 1px solid #EAEAEC;\n    }\n    \n    .purchase_heading p {\n      margin: 0;\n      color: #85878E;\n      font-size: 12px;\n    }\n    \n    .purchase_footer {\n      padding-top: 15px;\n      border-top: 1px solid #EAEAEC;\n    }\n    \n    .purchase_total {\n      margin: 0;\n      text-align: right;\n      font-weight: bold;\n      color: #333333;\n    }\n    \n    .purchase_total--label {\n      padding: 0 15px 0 0;\n    }\n    \n    body {\n      background-color: #F2F4F6;\n      color: #51545E;\n    }\n    \n    p {\n      color: #51545E;\n    }\n    \n    .email-wrapper {\n      width: 100%;\n      margin: 0;\n      padding: 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n      background-color: #F2F4F6;\n    }\n    \n    .email-content {\n      width: 100%;\n      margin: 0;\n      padding: 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n    }\n    /* Masthead ----------------------- */\n    \n    .email-masthead {\n      padding: 25px 0;\n      text-align: center;\n    }\n    \n    .email-masthead_logo {\n      width: 94px;\n    }\n    \n    .email-masthead_name {\n      font-size: 16px;\n      font-weight: bold;\n      color: #A8AAAF;\n      text-decoration: none;\n      text-shadow: 0 1px 0 white;\n    }\n    /* Body ------------------------------ */\n    \n    .email-body {\n      width: 100%;\n      margin: 0;\n      padding: 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n    }\n    \n    .email-body_inner {\n      width: 570px;\n      margin: 0 auto;\n      padding: 0;\n      -premailer-width: 570px;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n      background-color: #FFFFFF;\n    }\n    \n    .email-footer {\n      width: 570px;\n      margin: 0 auto;\n      padding: 0;\n      -premailer-width: 570px;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n      text-align: center;\n    }\n    \n    .email-footer p {\n      color: #A8AAAF;\n    }\n    \n    .body-action {\n      width: 100%;\n      margin: 30px auto;\n      padding: 0;\n      -premailer-width: 100%;\n      -premailer-cellpadding: 0;\n      -premailer-cellspacing: 0;\n      text-align: center;\n    }\n    \n    .body-sub {\n      margin-top: 25px;\n      padding-top: 25px;\n      border-top: 1px solid #EAEAEC;\n    }\n    \n    .content-cell {\n      padding: 45px;\n    }\n    /*Media Queries ------------------------------ */\n    \n    @media only screen and (max-width: 600px) {\n      .email-body_inner,\n      .email-footer {\n        width: 100% !important;\n      }\n    }\n    \n    @media (prefers-color-scheme: dark) {\n      body,\n      .email-body,\n      .email-body_inner,\n      .email-content,\n      .email-wrapper,\n      .email-masthead,\n      .email-footer {\n        background-color: #333333 !important;\n        color: #FFF !important;\n      }\n      p,\n      ul,\n      ol,\n      blockquote,\n      h1,\n      h2,\n      h3,\n      span,\n      .purchase_item {\n        color: #FFF !important;\n      }\n      .attributes_content,\n      .discount {\n        background-color: #222 !important;\n      }\n      .email-masthead_name {\n        text-shadow: none !important;\n      }\n    }\n    \n    :root {\n      color-scheme: light dark;\n      supported-color-schemes: light dark;\n    }\n    </style><!--[if mso]>\n    <style type=\"text/css\">\n      .f-fallback  {\n        font-family: Arial, sans-serif;\n      }\n    </style>\n  <![endif]--></head><body><span class=\"preheader\">Use this link to reset your password. The link is only valid for 24 hours.</span><table class=\"email-wrapper\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td align=\"center\"><table class=\"email-content\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td class=\"email-masthead\"><a href=\"https://example.com\" class=\"f-fallback email-masthead_name\">Grafto</a></td></tr><!-- Email Body --><tr><td class=\"email-body\" width=\"570\" cellpadding=\"0\" cellspacing=\"0\"><table class=\"email-body_inner\" align=\"center\" width=\"570\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><!-- Body content --><tr><td class=\"content-cell\"><div class=\"f-fallback\"><h1>Hi,</h1><p>You recently requested to reset your password for your Grafto account. Use the button below to reset it. <strong>This password reset is only valid for the next 24 hours.</strong></p><!-- Action --><table class=\"body-action\" align=\"center\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td align=\"center\"><table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\" role=\"presentation\"><tr><td align=\"center\"><a href=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 templ.SafeURL = templ.SafeURL(n.ResetPasswordLink)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var2)))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" class=\"f-fallback button button--green\" target=\"_blank\">Reset your password</a></td></tr></table></td></tr></table><p>If you did not request a password reset, please ignore this email or  <a href=\"support@mbv-labs.com\">contact support</a> you have questions.</p><p>Thanks,<br>The Grafto team</p><!-- Sub copy --><table class=\"body-sub\" role=\"presentation\"><tr><td><p class=\"f-fallback sub\">If you’re having trouble with the button above, copy and paste the URL below into your web browser.</p><p class=\"f-fallback sub\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(n.ResetPasswordLink)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `pkg/mail/templates/password_reset.templ`, Line: 545, Col: 63}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p></td></tr></table></div></td></tr></table></td></tr><tr>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = footer(n.UnsubscribeLink).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tr></table></td></tr></table></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
