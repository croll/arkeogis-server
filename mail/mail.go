package mail

import (
	"bytes"
	"log"
	"net/smtp"

	config "github.com/croll/arkeogis-server/config"
)

func init() {
}

func Send() (err error) {
	c, err := smtp.Dial("mail.example.com:25")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	// Set the sender and recipient.
	c.Mail(config.Main.Mail.Sender)
	c.Rcpt("beve@croll.fr")
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	defer wc.Close()
	buf := bytes.NewBufferString("This is the email body.")
	if _, err = buf.WriteTo(wc); err != nil {
		log.Fatal(err)
	}

	return
}
