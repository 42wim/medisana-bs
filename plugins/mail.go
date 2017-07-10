package plugins

import (
	"bytes"
	"fmt"
	"github.com/jovandeginste/medisana-bs/structs"
	"html/template"
	"log"
	"net/smtp"
	"sort"
	"strings"
	"time"
)

type Mail struct {
	Server        string
	SenderName    string
	SenderAddress string
	Recipients    map[int]structs.MailRecipient
	StartTLS      bool
	TemplateFile  string
	Subject       string
	Metrics       int
}

func (plugin Mail) Initialize() bool {
	log.Println("[MAIL PLUGIN] I am the Mail plugin")
	log.Printf("[MAIL PLUGIN]   - Server: %s\n", plugin.Server)
	log.Printf("[MAIL PLUGIN]   - StartTLS: %t\n", plugin.StartTLS)
	return true
}
func (plugin Mail) ParseData(person *structs.PersonMetrics) bool {
	log.Println("[MAIL PLUGIN] The mail plugin is parsing new data")
	plugin.sendMail(person)
	return true
}
func (mail Mail) sendMail(person *structs.PersonMetrics) {
	personId := person.Person
	recipient := mail.Recipients[personId]
	subject := mail.Subject

	metrics := make(structs.BodyMetrics, len(person.BodyMetrics))
	idx := 0
	for _, value := range person.BodyMetrics {
		metrics[idx] = value
		idx++
	}
	sort.Sort(metrics)
	lastMetrics := make(map[time.Time]structs.BodyMetric)
	for _, value := range metrics[len(metrics)-mail.Metrics:] {
		thetime := time.Unix(int64(value.Timestamp), 0)
		lastMetrics[thetime] = value
	}

	from := fmt.Sprintf("\"%s\" <%s>", mail.SenderName, mail.SenderAddress)
	to := recipient.Address

	var auth smtp.Auth
	auth = nil

	var msg string
	msg = msg + fmt.Sprintf("From: %s\r\n", from)
	msg = msg + fmt.Sprintf("To: %s\r\n", strings.Join(to, ";"))
	msg = msg + fmt.Sprintf("Subject: %s\r\n", subject)
	msg = msg + "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	msg = msg + "\r\n"
	parameters := struct {
		Name     string
		PersonId int
		Metrics  map[time.Time]structs.BodyMetric
	}{
		Name:     recipient.Name,
		PersonId: personId,
		Metrics:  lastMetrics,
	}
	body, err := ParseTemplate(mail.TemplateFile, parameters)
	if err != nil {
		log.Printf("[MAIL PLUGIN] An error occurred: %+v\n", err)
		return
	}
	msg = msg + body

	log.Printf("[MAIL PLUGIN] Sending mail to %s...\n", to)
	smtp.SendMail(mail.Server, auth, mail.SenderAddress, to, []byte(msg))
	log.Printf("[MAIL PLUGIN] Message was %d bytes.\n", len(msg))
}

func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	var result string
	t, err := template.ParseFiles(templateFileName)

	if err != nil {
		return result, err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return result, err
	}
	result = buf.String()
	return result, nil
}