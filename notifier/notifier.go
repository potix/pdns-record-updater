package watcher

import (
	"strings"
	"sync"
)


// Notifier is notifier
type Notifier struct {
	notifierConfig *configurator.Notifier
}

func (n notifier) sendMail(mail configurator.Mail, t *Time, record *configurator.Record, oldAlive uint32, newAlive uint32) {
        replacer := strings.NewReplacer(
                "%(time)", time.Format("2006-01-02 15:04:05"),
                "%(name)", record.Name,
                "%(type)", record.Type,
                "%(content)", record.Content,
                "%(oldAlive)", (oldAlive != 0),
                "%(newAlive)", (newAlive != 0))


	from := mail.Address{"", mail.From}
	toList, err := mail.ParseAddressList(mail.To)
	if err != nil {
		//log
	}
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mail.From)
	message += fmt.Sprintf("To: %s\r\n", mail.To)
	message += fmt.Sprintf("Subject: %s\r\n", replacer.Replace(mail.Subject))
	message += "\r\n" + replacer.Replace(mail.Body)

	host, _, _ := net.SplitHostPort(mail.HostPort)

	var auth Auth
	if strings.ToUpper(mail.AuthType) == "PLAIN" {
		auth = PlainAuth("", mail.Username, mail.Password, host)
	} else if strings.ToUpper(mail.AuthType) == "CRAM-MD5" {
		auth = CRAMMD5Auth(mail.Username, mail.Password)
	}

	if useSSL {
		// TLS config
		tlsConfig := &tls.Config {
			ServerName: host,
		}
		tlsConfig.InsecureSkipVerify = Mail.TLSSkipVerify
		conn, err := tls.Dial("tcp", mail.HostPort, tlsconfig)
		if err != nil {
			// log
		}
	} else {
		conn, err := net.Dial("tcp", mail.HostPort)
		if err != nil {
			// log
		}

	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		// log
	}

	if UseStarttls {
		tlsconfig := &tls.Config {
			ServerName: host,
		}
		tlsConfig.InsecureSkipVerify = Mail.TLSSkipVerify
		if err := client.StartTLS(tlsconfig); err != nil {
			// log
		}
	    }

	// Auth
	if err = client.Auth(auth); err != nil {
		// log
	}

	if err = client.Mail(from.Address); err != nil {
		// log
	}

	var emails []string
	for _,  to := range toList {
		emails = append(emails, to.Address)
	}
	recept := strings.Join(emails, ",")
	if err = client.Rcpt(recept); err != nil {
		// log
	}

	// Data
	w, err := client.Data()
	if err != nil {
		//log
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	if err != nil {
		// log
	}
}

func (n *Notifier) notifyMain(t *Time, record *configurator.Record, oldAlive uint32, newAlive uint32) {
	// send mail
	for _, mail := range notifierConfig.Mail {
		err := sendMail(mail, t, record, oldAlive, newAlive)
	}
}

// Notify is Notify
func (n *Notifier) Notify(record *configurator.Record, oldAlive uint32, newAlive uint32) {
	go n.notifyMain(&time.Now(), record, oldAlive, newAlive)
}

// New is create notifier
func New(config *configurator.Config) (n *Notifier) {
	return &notifier{
		notifierConfig: config.Notifier,
	}
}
