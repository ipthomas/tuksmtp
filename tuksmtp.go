package tuksmtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type NotifyEvent struct {
	Body     string
	From     string
	To       string
	Server   string
	Port     string
	Password string
}

func (i *NotifyEvent) Notify() error {
	var err error

	auth := smtp.PlainAuth("", i.From, i.Password, i.Server)
	conn, err := smtp.Dial(i.Server + ":" + i.Port)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("Connected to smtp server : %s", i.Server)
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
		return err
	}
	log.Printf("Set Email from : %s", i.From)
	if err = conn.Rcpt(i.To); err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("Set Email to : %s", i.To)
	wc, err := conn.Data()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer wc.Close()
	if _, err = fmt.Fprint(wc, i.Body); err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("Notification sent to %s", i.To)
	}

	return err
}
