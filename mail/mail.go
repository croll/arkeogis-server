package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"strings"

	config "github.com/croll/arkeogis-server/config"
	"github.com/croll/arkeogis-server/translate"
	"github.com/juju/errors"
)

var isTranslatableStringRegexp = regexp.MustCompile(`^[A-Z]+\.[A-Z]+\.T+_+[A-Z]+$`)

// Send sends a mail. Amazing no ?
func Send(to []string, subject, body string, langIsoCode ...string) (err error) {

	lang := "en"

	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)

	if langIsoCode[0] != "" {
		lang = langIsoCode[0]
	}

	if subject == "" {
		return errors.New("Subject can't be empty")
	}

	if body == "" {
		return errors.New("Message can't be empty")
	}

	if isTranslatableStringRegexp.MatchString(subject) {
		subject = translate.T(lang, subject)
	}

	if isTranslatableStringRegexp.MatchString(body) {
		body = translate.T(lang, body)
	}

	/*
		message := "From: " + config.Main.Mail.From + "\n" +
			"To: " + "beve@croll.fr" + "\n" +
			"Subject: " + subject + "\n\n" +
			body

		auth := smtp.PlainAuth("", config.Main.Mail.User, config.Main.Mail.Password, config.Main.Mail.Host)

		err = smtp.SendMail(config.Main.Mail.Host, auth, config.Main.Mail.From, to, []byte(message))
		fmt.Println(err)
	*/ // the basics

	// setup the remote smtpserver & auth info
	auth := smtp.PlainAuth("", config.Main.Mail.User, config.Main.Mail.Password, config.Main.Mail.Host)

	// setup a map for the headers
	header := make(map[string]string)
	header["From"] = config.Main.Mail.From
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject

	// setup the message
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// create the smtp connection
	c, err := smtp.Dial(config.Main.Mail.Host)
	if err != nil {
		return errors.Annotate(err, "Dial failed")
	}

	// set some TLS options, so we can make sure a non-verified cert won't stop us sending
	host, _, _ := net.SplitHostPort(config.Main.Mail.Host)

	tlc := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = c.StartTLS(tlc); err != nil {
		return errors.Annotate(err, "Start tls handshake failed")
	}

	// auth stuff
	if err = c.Auth(auth); err != nil {
		return errors.Annotate(err, "Auth failed")
	}

	// To && From
	if err = c.Mail(config.Main.Mail.From); err != nil {
		return errors.Annotate(err, "Setting 'From' header failed")
	}
	if err = c.Rcpt(strings.Join(to, ",")); err != nil {
		return errors.Annotate(err, "Setting 'To' header failed")
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return errors.Annotate(err, "Setting data failed")
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return errors.Annotate(err, "Writing message failed")
	}
	err = w.Close()
	if err != nil {
		return errors.Annotate(err, "Closing connexion failed")
	}
	c.Quit()
	return
}
