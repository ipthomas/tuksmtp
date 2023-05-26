package tuksmtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"

	"github.com/ipthomas/tukdbint"
)

type NotifyEvent struct {
	Pathway       string
	Expression    string
	NHSId         string
	User          string
	Org           string
	Role          string
	Subject       string
	From          string
	Server        string
	Port          string
	Password      string
	Subscriptions tukdbint.Subscriptions
}

var emailTemplate = "{{define 'notifcation'}}Notification of {{.Pathway}} Workflow {{.Expression}} Event for NHS ID {{.NHSId}}\nEvent Created by User {{.User}} at Organisation {{.Org}} in the Role of {{.Role}}\n{{end}}"

func (i *NotifyEvent) Notify() error {
	var err error
	var buf bytes.Buffer
	var tmplt *template.Template
	if tmplt, err = template.New("emailTemplate").Parse(emailTemplate); err == nil {
		if err = tmplt.Execute(&buf, i); err == nil {
			body := "ICB Workflow Event\n\n" + buf.String()
			for _, v := range i.Subscriptions.Subscriptions {
				if v.Email != "" {
					emailBody := fmt.Sprintf("Subject: %s\r\n\r\n%s", i.Subject, body)
					auth := smtp.PlainAuth("", i.From, i.Password, i.Server)
					conn, err := smtp.Dial(i.Server + ":" + i.Port)
					if err != nil {
						log.Println(err.Error())
						return err
					}
					defer conn.Close()
					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						ServerName:         i.Server,
					}
					if err = conn.StartTLS(tlsConfig); err == nil {
						if err = conn.Auth(auth); err == nil {
							if err = conn.Mail(i.From); err == nil {
								if err = conn.Rcpt(v.Email); err == nil {
									if wc, err := conn.Data(); err != nil {
										log.Println(err.Error())
									} else {
										if _, err = fmt.Fprint(wc, emailBody); err == nil {
											log.Printf("Notification sent to %s", v.Email)
										}
									}
								}
							}
						}
					}
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}
	} else {
		log.Println(err.Error())
	}
	return err
}
