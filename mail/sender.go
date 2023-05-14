package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(subject string, content string, to []string,
		cc []string, bcc []string, attachFiles []string) error
}

type GmailEmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailEmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailEmailSender{
		name,
		fromEmailAddress,
		fromEmailPassword,
	}
}

func (sender GmailEmailSender) SendEmail(subject string, content string,
	to []string, cc []string, bcc []string, attachFiles []string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, file := range attachFiles {
		if _, err := e.AttachFile(file); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", file, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)

	return e.Send(smtpServerAddress, smtpAuth)
}
