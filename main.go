package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"

	"github.com/ipthomas/tukdbint"
)

type SMNP struct {
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

type TUK_DB_Interface interface {
	newNotifyEvent() error
}

func NewNotifyEvent(i TUK_DB_Interface) error {
	return i.newNotifyEvent()
}

func (i *SMNP) newNotifyEvent() error {
	log.SetFlags(log.Lshortfile)
	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, i); err != nil {
		log.Println(err.Error())
		return err
	}
	body := "ICB Workflow Event\n\n" + buf.String()
	subject := i.Subject
	from := i.From
	smtpServer := i.Server
	smtpPort := i.Port
	smtpPassword := i.Password

	for _, v := range i.Subscriptions.Subscriptions {
		if v.Email != "" {
			if err := SendEmail(from, v.Email, subject, smtpServer, smtpPassword, smtpPort, body); err != nil {
				log.Printf("Error sending email to %s: Error: %s", v.Email, err.Error())
			} else {
				log.Printf("Notification sent to %s", v.Email)
			}
		}
	}
	return nil
}
func SendEmail(from, to, subject, smtpServer, smtpPassword, smtpPort, body string) error {
	emailBody := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
	auth := smtp.PlainAuth("", from, smtpPassword, smtpServer)
	conn, err := smtp.Dial(smtpServer + ":" + smtpPort)
	if err != nil {
		return err
	}
	defer conn.Close()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}
	if err := conn.StartTLS(tlsConfig); err != nil {
		return err
	}
	if err := conn.Auth(auth); err != nil {
		return err
	}
	if err := conn.Mail(from); err != nil {
		return err
	}
	if err := conn.Rcpt(to); err != nil {
		return err
	}
	wc, err := conn.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	_, err = fmt.Fprint(wc, emailBody)
	if err != nil {
		return err
	}
	return nil
}
