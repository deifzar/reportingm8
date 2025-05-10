package email8

import (
	"deifzar/reportingm8/pkg/configparser"
	"net/smtp"
	"strconv"
)

type Email8 struct {
	server   string
	port     int
	address  string
	user     string
	password string
	auth     smtp.Auth
	sender   string
}

func NewEmail8() (Email8Interface, error) {
	v, err := configparser.InitConfigParser()
	if err != nil {
		return &Email8{}, err
	}
	// smtp config
	server := v.GetString("REPORTINGM8.SMTP.server")
	port := v.GetInt("REPORTINGM8.SMTP.port")
	address := server + ":" + strconv.Itoa(port)
	user := v.GetString("REPORTINGM8.SMTP.username")
	pass := v.GetString("REPORTINGM8.SMTP.password")
	sender := v.GetString("REPORTINGM8.SMTP.emailsender")
	auth := smtp.PlainAuth("", user, pass, server)

	return &Email8{server: server, port: port, address: address, user: user, password: pass, auth: auth, sender: sender}, nil
}

func (m *Email8) SendEmail(recipients []string, message []byte) error {
	return smtp.SendMail(m.address, m.auth, m.sender, recipients, message)
}
