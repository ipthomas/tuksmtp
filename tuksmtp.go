package tuksmtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/ipthomas/tukdbint"
)

type NotifyEvent struct {
	Subject       string
	Body          string
	From          string
	Server        string
	Port          string
	Password      string
	Subscriptions tukdbint.Subscriptions
	Event         tukdbint.Event
}

func (i *NotifyEvent) Notify() error {
	var err error
	for _, v := range i.Subscriptions.Subscriptions {
		if i.shouldNotify(v) {
			body := "ICB Workflow Event\n\n" + i.Body
			emailBody := fmt.Sprintf("Subject: %s\r\n\r\n%s", i.Subject, body)
			auth := smtp.PlainAuth("", i.From, i.Password, i.Server)
			conn, err := smtp.Dial(i.Server + ":" + i.Port)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			log.Printf("Set Email Body\n%s", emailBody)
			defer conn.Close()
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         i.Server,
			}
			if err = conn.StartTLS(tlsConfig); err != nil {
				log.Println(err.Error())
				return err
			}
			log.Printf("Opened TLS connection to %s", i.Server)
			if err = conn.Auth(auth); err != nil {
				log.Println(err.Error())
				return err
			}
			log.Printf("Successfully Authenticated as %s", i.From)
			if err = conn.Mail(i.From); err != nil {
				log.Println(err.Error())
				continue
			}
			if err = conn.Rcpt(v.Email); err != nil {
				log.Println(err.Error())
				continue
			}
			wc, err := conn.Data()
			if err != nil {
				log.Println(err.Error())
				continue
			}
			if _, err = fmt.Fprint(wc, emailBody); err != nil {
				log.Println(err.Error())
			} else {
				log.Printf("Notification sent to %s", v.Email)
			}
		}
	}
	return err
}
func (i *NotifyEvent) shouldNotify(sub tukdbint.Subscription) bool {
	var shouldReturn bool
	if sub.Email != "" {
		if sub.Pathway != "" {
			shouldReturn = (i.Event.Pathway == sub.Pathway) && (sub.NhsId == "" || sub.NhsId == i.Event.NhsId) && (sub.Expression == "" || i.Event.Expression == sub.Expression)
		} else {
			if sub.NhsId != "" {
				shouldReturn = (i.Event.NhsId == sub.NhsId) && (sub.Expression == "" || i.Event.Expression == sub.Expression)
			} else {
				if sub.Expression != "" {
					shouldReturn = i.Event.Expression == sub.Expression
				} else {
					shouldReturn = true
				}
			}
		}
	}
	log.Printf("Subscription Matched : %v", shouldReturn)
	return shouldReturn
}
