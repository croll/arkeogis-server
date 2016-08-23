package mail

import (
	"net/smtp"

	config "github.com/croll/arkeogis-server/config"
)

func init() {
}

func Send(to []string, subject, message string) (err error) {

	auth := smtp.PlainAuth("", config.Main.Mail.User, config.Main.Mail.Password, config.Main.Mail.Host)

	err = smtp.SendMail(
		config.Main.Mail.Host,
		auth,
		config.Main.Mail.From,
		to,
		[]byte(message),
		//[]byte("This is the email body."),
	)
	return
}
