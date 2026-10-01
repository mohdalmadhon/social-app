package mailer

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const (
	smtpHost = "smtp.gmail.com"
	smtpPort = "587"
)

type Config struct {
	EmailPassword string
	EmailAddress  string
}

type Mailer struct {
	cfg Config
}

type otpView struct {
	Digits  []string
	Code    string
	Minutes int
	Tries   int
}

var otpTemplate = template.Must(template.New("otp").Parse(otpHTML))

func New(cfg Config) *Mailer {
	cfg.EmailPassword = strings.ReplaceAll(cfg.EmailPassword, " ", "")

	return &Mailer{cfg: cfg}
}

func wrap76(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)

	var b strings.Builder

	for len(encoded) > 76 {
		b.WriteString(encoded[:76])
		b.WriteString("\r\n")
		encoded = encoded[76:]
	}

	b.WriteString(encoded)

	return b.String()
}

func randomID() string {
	b := make([]byte, 12)
	rand.Read(b)

	return hex.EncodeToString(b)
}

func (m *Mailer) build(to, subject, plain, html string) []byte {
	boundary := "social-" + randomID()
	from := mail.Address{Name: "social network", Address: m.cfg.EmailAddress}

	domain := "socialnetwork.app"

	if at := strings.LastIndex(m.cfg.EmailAddress, "@"); at >= 0 {
		domain = m.cfg.EmailAddress[at+1:]
	}

	var b bytes.Buffer

	fmt.Fprintf(&b, "From: %s\r\n", from.String())
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", randomID(), domain)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(wrap76([]byte(plain)))
	b.WriteString("\r\n")

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(wrap76([]byte(html)))
	b.WriteString("\r\n")

	fmt.Fprintf(&b, "--%s--\r\n", boundary)

	return b.Bytes()
}

func (m *Mailer) send(to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(smtpHost, smtpPort), 10*time.Second)
	if err != nil {
		return err
	}

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: smtpHost}); err != nil {
		return err
	}

	auth := smtp.PlainAuth("", m.cfg.EmailAddress, m.cfg.EmailPassword, smtpHost)

	if err := client.Auth(auth); err != nil {
		return err
	}

	if err := client.Mail(m.cfg.EmailAddress); err != nil {
		return err
	}

	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (m *Mailer) SendVerificationCode(to, code string, ttl time.Duration, tries int) error {
	view := otpView{
		Digits:  strings.Split(code, ""),
		Code:    code,
		Minutes: int(ttl.Minutes()),
		Tries:   tries,
	}

	var htmlBody bytes.Buffer

	if err := otpTemplate.Execute(&htmlBody, view); err != nil {
		return err
	}

	plain := fmt.Sprintf(
		"social network - VERIFICATION\r\n\r\nYour verification code is: %s\r\n\r\nIt expires in %d minutes and you have %d tries.\r\n\r\nDid not request this? Ignore this email.\r\n",
		code, view.Minutes, tries,
	)

	msg := m.build(to, "social network - Your verification code is "+code, plain, htmlBody.String())

	return m.send(to, msg)
}
