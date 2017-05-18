package watcher

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
	"github.com/potix/pdns-record-updater/configurator"
	"net"
	"crypto/tls"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
	"fmt"
)


// Notifier is notifier
type Notifier struct {
	notifierConfig *configurator.Notifier
}

func (n *Notifier) sendMail(mailConfig *configurator.Mail, t time.Time, record *configurator.Record, oldAlive uint32, newAlive uint32) (error) {
        replacer := strings.NewReplacer(
                "%(time)", t.Format("2006-01-02 15:04:05"),
                "%(name)", record.Name,
                "%(type)", record.Type,
                "%(content)", record.Content,
                "%(oldAlive)", fmt.Sprintf("%v", (oldAlive != 0)),
                "%(newAlive)", fmt.Sprintf("%v", (newAlive != 0)))

	from := mail.Address{"", mailConfig.From}
	toList, err := mail.ParseAddressList(mailConfig.To)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not parse mail address list (%v)", mailConfig.To))
	}
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mailConfig.From)
	message += fmt.Sprintf("To: %s\r\n", mailConfig.To)
	subject := mailConfig.Subject
	if subject == "" {
		subject = "%(name) %(type) %(content): old alive = %(oldAlive) -> new alive = %(newAlive)"
	}
	message += fmt.Sprintf("Subject: %s\r\n", replacer.Replace(subject))
	body := mailConfig.Body
	if body == "" {
		subject = "%(time) %(name) %(type) %(content): old alive = %(oldAlive) -> new alive = %(newAlive)"
	}
	message += "\r\n" + replacer.Replace(body)

	host, _, _ := net.SplitHostPort(mailConfig.HostPort)

	var auth smtp.Auth
	if strings.ToUpper(mailConfig.AuthType) == "PLAIN" {
		auth = smtp.PlainAuth("", mailConfig.Username, mailConfig.Password, host)
	} else if strings.ToUpper(mailConfig.AuthType) == "CRAM-MD5" {
		auth = smtp.CRAMMD5Auth(mailConfig.Username, mailConfig.Password)
	}

	var conn net.Conn
	if mailConfig.UseTLS {
		// TLS config
		tlsConfig := &tls.Config {
			ServerName: host,
			InsecureSkipVerify: mailConfig.TLSSkipVerify,
		}
		conn, err = tls.Dial("tcp", mailConfig.HostPort, tlsConfig)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not connect mail host with tls (%v)", mailConfig.HostPort))
		}
	} else {
		conn, err = net.Dial("tcp", mailConfig.HostPort)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not connect mail host (%v)", mailConfig.HostPort))
		}
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not create smtp client (%v)", mailConfig.HostPort))
	}

	if mailConfig.UseStartTLS {
		tlsconfig := &tls.Config {
			ServerName: host,
			InsecureSkipVerify: mailConfig.TLSSkipVerify,
		}
		if err := client.StartTLS(tlsconfig); err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not start tls (%v)", mailConfig.HostPort))
		}
	    }

	// Auth
	if err = client.Auth(auth); err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not authentication (%v) (%v)", mailConfig.Username, mailConfig.Password))
	}

	if err = client.Mail(from.Address); err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not send MAIL command (%v)", from.Address))
	}

	var emails []string
	for _,  to := range toList {
		emails = append(emails, to.Address)
	}
	recept := strings.Join(emails, ",")
	if err = client.Rcpt(recept); err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not send RCPT command (%v)", recept))
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not send DATA command"))
	}

	// write message
	_, err = w.Write([]byte(message))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not write message (%v)", message))
	}

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not close message writer"))
	}

	// quit
	err = client.Quit()
	if err != nil {
		belog.Notice("%v", errors.Wrap(err, fmt.Sprintf("can not send QUIT command")))
	}

	return nil
}

func (n *Notifier) notifyMain(t time.Time, record *configurator.Record, oldAlive uint32, newAlive uint32) {
	// send mail
	for _, mailConfig := range n.notifierConfig.Mail {
		err := n.sendMail(mailConfig, t, record, oldAlive, newAlive)
		if err != nil {
			belog.Error("%v", err)
		}
	}
}

// Notify is Notify
func (n *Notifier) Notify(record *configurator.Record, oldAlive uint32, newAlive uint32) {
	go n.notifyMain(time.Now(), record, oldAlive, newAlive)
}

// New is create notifier
func New(config *configurator.Config) (n *Notifier) {
	return &Notifier{
		notifierConfig: config.Notifier,
	}
}
