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
	"time"
	"os"
	"fmt"
)

// Notifier is notifier
type Notifier struct {
	hostname string
	notifierContext *contexter.Notifier
}

func (n *Notifier) sendMail(mailContext *contexter.Mail, t time.Time, domain string, record *contexter.Record, targetResult string, oldAlive uint32, newAlive uint32) (error) {
        replacer := strings.NewReplacer(
		"%(hostname)", n.hostname,
                "%(time)", t.Format("2006-01-02 15:04:05"),
                "%(zone)", domain,
                "%(name)", record.Name,
                "%(type)", record.Type,
                "%(content)", record.Content,
                "%(oldAlive)", fmt.Sprintf("%v", (oldAlive != 0)),
                "%(newAlive)", fmt.Sprintf("%v", (newAlive != 0)),
                "%(detail)", targetResult)

	from := mail.Address{"", mailContext.From}
	toList, err := mail.ParseAddressList(mailContext.To)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not parse mail address list (%v)", mailContext.To))
	}
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mailContext.From)
	message += fmt.Sprintf("To: %s\r\n", mailContext.To)
	subject := mailContext.Subject
	if subject == "" {
		subject = "%(hostname) %(zone) %(name) %(content): old alive = %(oldAlive) -> new alive = %(newAlive)"
	}
	message += fmt.Sprintf("Subject: %s\r\n", replacer.Replace(subject))
	body := mailContext.Body
	if body == "" {
		body = "hostname: %(hostname)\nzone: %(zone)\nrecord: %(name) %(type) %(content)\n%(time) old alive = %(oldAlive) -> new alive = %(newAlive)\n%(detail)"
	}
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
			return errors.Wrap(err, fmt.Sprintf("can not connect mail host with tls (%v)", mailContext.HostPort))
		}
	} else {
		conn, err = net.Dial("tcp", mailContext.HostPort)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not connect mail host (%v)", mailContext.HostPort))
		}
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not create smtp client (%v)", mailContext.HostPort))
	}

	if mailContext.UseStartTLS {
		tlsconfig := &tls.Config {
			ServerName: host,
			InsecureSkipVerify: mailContext.TLSSkipVerify,
		}
		if err := client.StartTLS(tlsconfig); err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not start tls (%v)", mailContext.HostPort))
		}
	    }

	if err = client.Auth(auth); err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not authentication (%v) (%v)", mailContext.Username, mailContext.Password))
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

	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not send DATA command"))
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not write message (%v)", message))
	}

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not close message writer"))
	}

	err = client.Quit()
	if err != nil {
		belog.Notice("%v", errors.Wrap(err, fmt.Sprintf("can not send QUIT command")))
	}

	return nil
}

func (n *Notifier) notifyMain(t time.Time, domain string, record *contexter.Record, targetResult string, oldAlive uint32, newAlive uint32) {
	// send mail
	for _, mailContext := range n.notifierContext.Mail {
		err := n.sendMail(mailContext, t, domain, record, targetResult, oldAlive, newAlive)
		if err != nil {
			belog.Error("%v", err)
		}
	}
}

// Notify is Notify
func (n *Notifier) Notify(domain string, record *contexter.Record, targetResult string, oldAlive uint32, newAlive uint32) {
	go n.notifyMain(time.Now(), domain, record, targetResult, oldAlive, newAlive)
}

// New is create notifier
func New(context *contexter.Context) (n *Notifier) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &Notifier{
		hostname : hostname,
		notifierContext: context.Notifier,
	}
}
