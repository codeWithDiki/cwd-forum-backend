package mail

import "github.com/jordan-wright/email"

type SMTPMailer struct {
	Mailer *email.Email
}

func NewMailer(
	email *email.Email,
) *SMTPMailer {
	return &SMTPMailer{
		Mailer: email,
	}
}
