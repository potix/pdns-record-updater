package notifier

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
	"github.com/potix/pdns-record-updater/contexter"
	"net"
	"crypto/tls"
	"net/mail"
	"net/smtp"
	"strings"
	"os"
	"fmt"
)

// Notifier is notifier
type Notifier struct {
	hostname string
	notifierContext *contexter.Notifier
}

func (n *Notifier) sendMail(mailContext *contexter.Mail, replacer *strings.Replacer, subject string, body string) {
	from := mail.Address{"", mailContext.From}
	toList, err := mail.ParseAddressList(mailContext.To)
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not parse mail address list (%v)", mailContext.To)))
	}
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mailContext.From)
	message += fmt.Sprintf("To: %s\r\n", mailContext.To)
	message += fmt.Sprintf("Subject: %s\r\n", replacer.Replace(subject))
	message += "\r\n" + replacer.Replace(body)

	host, _, _ := net.SplitHostPort(mailContext.HostPort)

	var auth smtp.Auth
	if strings.ToUpper(mailContext.AuthType) == "PLAIN" {
		auth = smtp.PlainAuth("", mailContext.Username, mailContext.Password, host)
	} else if strings.ToUpper(mailContext.AuthType) == "CRAM-MD5" {
		auth = smtp.CRAMMD5Auth(mailContext.Username, mailContext.Password)
	}

	var conn net.Conn
	if mailContext.UseTLS {
		tlsContext := &tls.Config {
			ServerName: host,
			InsecureSkipVerify: mailContext.TLSSkipVerify,
		}
		conn, err = tls.Dial("tcp", mailContext.HostPort, tlsContext)
		if err != nil {
			belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not connect mail host with tls (%v)", mailContext.HostPort)))
		}
	} else {
		conn, err = net.Dial("tcp", mailContext.HostPort)
		if err != nil {
			belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not connect mail host (%v)", mailContext.HostPort)))
		}
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not create smtp client (%v)", mailContext.HostPort)))
	}

	if mailContext.UseStartTLS {
		tlsconfig := &tls.Config {
			ServerName: host,
			InsecureSkipVerify: mailContext.TLSSkipVerify,
		}
		if err := client.StartTLS(tlsconfig); err != nil {
			belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not start tls (%v)", mailContext.HostPort)))
		}
	    }

	if err = client.Auth(auth); err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not authentication (%v) (%v)", mailContext.Username, mailContext.Password)))
	}

	if err = client.Mail(from.Address); err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not send MAIL command (%v)", from.Address)))
	}

	var emails []string
	for _,  to := range toList {
		emails = append(emails, to.Address)
	}
	recept := strings.Join(emails, ",")
	if err = client.Rcpt(recept); err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not send RCPT command (%v)", recept)))
	}

	w, err := client.Data()
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not send DATA command")))
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not write message (%v)", message)))
	}

	err = w.Close()
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not close message writer")))
	}

	err = client.Quit()
	if err != nil {
		belog.Notice("%v", errors.Wrap(err, fmt.Sprintf("can not send QUIT command")))
	}
}

// Notify is Notify
func (n *Notifier) Notify(replacer *strings.Replacer, subject string, body string) {
	for _, mailContext := range n.notifierContext.Mail {
		go n.sendMail(mailContext, replacer, subject, body)
	}
}

// New is create notifier
func New(notifierContext *contexter.Notifier) (n *Notifier) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &Notifier{
		hostname : hostname,
		notifierContext: notifierContext,
	}
}
