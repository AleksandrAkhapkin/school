package mail

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tarasova-school/internal/types"
	"github.com/tarasova-school/internal/types/config"
	"html/template"
	"net/smtp"
)

type Request struct {
	from  string
	to    []string
	body  string
	email *config.ConfigForSendEmail
}

func NewRequest(to []string, cnf *config.ConfigForSendEmail) *Request {
	return &Request{to: to, email: cnf}
}

func (r *Request) parseTemplate(fileName string, data interface{}) error {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return errors.Wrap(err, "err while ParseFiles")
	}
	buffer := new(bytes.Buffer)
	dataStr := struct {
		Data string
	}{
		Data: data.(string),
	}

	if err = t.Execute(buffer, dataStr); err != nil {
		return errors.Wrap(err, "err while Execute")
	}
	r.body = buffer.String()
	return nil
}

func (r *Request) sendMail() error {
	body := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\r\n%s\r\n%s", r.email.EmailLogin, r.to[0], types.EmailRecoveryTitle, types.MIME, r.body)
	addr := fmt.Sprintf("%s:%s", r.email.EmailHost, r.email.EmailPort)
	if err := smtp.SendMail(addr, smtp.PlainAuth("", r.email.EmailLogin, r.email.EmailPass, r.email.EmailHost), r.email.EmailLogin, r.to, []byte(body)); err != nil {
		return errors.Wrap(err, "err while SendMail")
	}
	return nil
}

func (r *Request) Send(templateName string, data interface{}) error {
	err := r.parseTemplate(templateName, data)
	if err != nil {
		return errors.Wrap(err, "err while parseTemplate")
	}
	if err := r.sendMail(); err != nil {
		return errors.Wrap(err, "err while sendMail")
	}

	return nil
}
