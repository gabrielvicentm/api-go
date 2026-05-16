package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type AdminPasswordResetNotifier interface {
	SendAdminPasswordReset(ctx context.Context, nome, email, token string, expiresAt time.Time) error
}

type SMTPAdminPasswordResetNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

func NewAdminPasswordResetNotifierFromEnv() *SMTPAdminPasswordResetNotifier {
	port, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SMTP_PORT")))
	if err != nil || port <= 0 {
		port = 587
	}

	return &SMTPAdminPasswordResetNotifier{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     port,
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		from:     strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		fromName: strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}
}

func (n *SMTPAdminPasswordResetNotifier) SendAdminPasswordReset(_ context.Context, nome, email, token string, expiresAt time.Time) error {
	if n == nil {
		return nil
	}

	if n.host == "" || n.from == "" {
		log.Printf("password reset token for %s <%s>: %s (expira em %s)", nome, email, token, expiresAt.Format(time.RFC3339))
		return nil
	}

	subject := "Redefinicao de senha administrativa"
	body := fmt.Sprintf(
		"Ola %s,\r\n\r\nUse o token abaixo para redefinir sua senha no painel administrativo:\r\n\r\n%s\r\n\r\nValido ate %s.\r\n\r\nSe voce nao solicitou este reset, ignore esta mensagem.\r\n",
		strings.TrimSpace(nome),
		token,
		expiresAt.Local().Format("02/01/2006 15:04"),
	)

	fromHeader := n.from
	if n.fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", n.fromName, n.from)
	}

	message := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + email,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	var auth smtp.Auth
	if n.username != "" && n.password != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}

	return smtp.SendMail(addr, auth, n.from, []string{email}, []byte(message))
}
