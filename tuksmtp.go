package tuksmtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/ipthomas/tukdbint"
	"github.com/ipthomas/tukutil"
)

type NotifyEvent struct {
	SubscriberURL string
	ConsumerURL   string
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
		if stmt, shouldNotify := i.shouldNotify(v); shouldNotify {
			i.ConsumerURL = i.ConsumerURL + "?act=select&user=" + v.User + "&org=" + v.Org + "&role=" + v.Role + "&config=xdw&pathway=" + i.Event.Pathway + "&nhs=" + i.Event.NhsId + "&vers=" + tukutil.GetStringFromInt(i.Event.Version) + "&_format=html"
			i.SubscriberURL = i.SubscriberURL + "?act=select&topic=EMAIL&user=" + v.User + "&org=" + v.Org + "&role=" + v.Role + "&email=" + v.Email + "&_format=html"
			i.Body = stmt + i.Body
			wfurl := fmt.Sprintf("\r\n\r\nClick this link to view Workflow Details\r\n %s \r\n", i.ConsumerURL)
			subsurl := fmt.Sprintf("\r\n\r\nClick this link to manage your Subscriptions to ICB Notifications\r\n %s \r\n", i.SubscriberURL)
			emailBody := fmt.Sprintf("Subject: %s\r\n\r\n%s", i.Subject, i.Body)
			emailBody = emailBody + wfurl + subsurl
			log.Printf("Set Email Content\n%s", emailBody)

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
				continue
			}
			log.Printf("Set Email from : %s", i.From)
			if err = conn.Rcpt(v.Email); err != nil {
				log.Println(err.Error())
				continue
			}
			log.Printf("Set Email to : %s", v.Email)
			wc, err := conn.Data()
			if err != nil {
				log.Println(err.Error())
				continue
			}
			defer wc.Close()
			if _, err = fmt.Fprint(wc, emailBody); err != nil {
				log.Println(err.Error())
			} else {
				log.Printf("Notification sent to %s", v.Email)
			}
		}
	}
	return err
}

func (i *NotifyEvent) shouldNotify(sub tukdbint.Subscription) (string, bool) {
	var stmt = "You have subscribed to receive ICB Workflow Notifications"
	var shouldReturn bool
	if sub.Email != "" {
		if sub.Pathway != "" {
			shouldReturn = (i.Event.Pathway == sub.Pathway) && (sub.NhsId == "" || sub.NhsId == i.Event.NhsId) && (sub.Expression == "" || i.Event.Expression == sub.Expression)
			stmt = stmt + " for " + sub.Pathway + " " + sub.NhsId + " " + sub.Expression + " Events"
		} else {
			if sub.NhsId != "" {
				stmt = stmt + " for " + sub.NhsId + " " + sub.Expression + " Events"
				shouldReturn = (i.Event.NhsId == sub.NhsId) && (sub.Expression == "" || i.Event.Expression == sub.Expression)
			} else {
				if sub.Expression != "" {
					stmt = stmt + " for " + sub.Expression + " Events"
					shouldReturn = i.Event.Expression == sub.Expression
				} else {
					shouldReturn = true
				}
			}
		}
	}
	stmt = stmt + "\n\n"
	log.Printf("Subscription Matched : %v", shouldReturn)
	return stmt, shouldReturn
}

// package main

// import (
// 	"crypto/tls"
// 	"fmt"
// 	"net/smtp"
// )

// func SendEmail(from, to, subject, body string) error {
// 	// SMTP server configuration
// 	smtpServer := "smtp.gmail.com"
// 	smtpPort := "587"
// 	smtpUsername := "admin@tiani-spirit.co.uk"
// 	smtpPassword := "cTjh_M7b7XWZa9_8qZMR"

// 	// Compose the email
// 	emailBody := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

// 	// Establish a connection to the SMTP server
// 	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)
// 	conn, err := smtp.Dial(smtpServer + ":" + smtpPort)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	// StartTLS to initiate a secure (encrypted) connection
// 	tlsConfig := &tls.Config{
// 		InsecureSkipVerify: true,
// 		ServerName:         smtpServer,
// 	}
// 	if err := conn.StartTLS(tlsConfig); err != nil {
// 		return err
// 	}

// 	// Authenticate
// 	if err := conn.Auth(auth); err != nil {
// 		return err
// 	}

// 	// Set the sender and recipient
// 	if err := conn.Mail(from); err != nil {
// 		return err
// 	}
// 	if err := conn.Rcpt(to); err != nil {
// 		return err
// 	}

// 	// Send the email body
// 	wc, err := conn.Data()
// 	if err != nil {
// 		return err
// 	}
// 	defer wc.Close()

// 	_, err = fmt.Fprint(wc, emailBody)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func main() {
// 	from := "admin@tiani-spirit.co.uk"
// 	to := "ian.thomas@tiani-spirit.co.uk"
// 	subject := "Hello, World!"
// 	body := "This is the body of the email."

// 	err := SendEmail(from, to, subject, body)
// 	if err != nil {
// 		fmt.Println("Error sending email:", err)
// 	} else {
// 		fmt.Println("Email sent successfully!")
// 	}
// }
