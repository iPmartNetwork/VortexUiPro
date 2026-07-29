package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"vortexuipro/internal/database"
)

// EmailService handles SMTP email sending with templates.
type EmailService struct{}

// NewEmailService creates a new email service.
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

// LoadConfig reads SMTP settings from the database.
func (s *EmailService) LoadConfig() (*SMTPConfig, error) {
	host, _ := database.GetSetting("smtp_host")
	portStr, _ := database.GetSetting("smtp_port")
	username, _ := database.GetSetting("smtp_username")
	password, _ := database.GetSetting("smtp_password")
	from, _ := database.GetSetting("smtp_from")

	if host == "" {
		return nil, fmt.Errorf("SMTP not configured")
	}

	port := 587
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}
	if from == "" {
		from = username
	}

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		UseTLS:   port == 465,
	}, nil
}

// Send sends an email to the specified recipient.
func (s *EmailService) Send(to, subject, body string) error {
	cfg, err := s.LoadConfig()
	if err != nil {
		return err
	}

	msg := s.buildMessage(cfg.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	if cfg.UseTLS {
		return s.sendTLS(addr, cfg, msg)
	}
	return s.sendSTARTTLS(addr, cfg, msg)
}

// SendWithTemplate sends an email using a named template.
func (s *EmailService) SendWithTemplate(to, templateName string, data map[string]string) error {
	subject := "Notification from VortexUiPro"
	body := ""

	switch templateName {
	case "welcome":
		subject = "Welcome to VortexUiPro"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #8b5cf6, #06b6d4); padding: 30px; border-radius: 12px; text-align: center;">
    <h2 style="color: white; margin: 0;">Welcome to VortexUiPro</h2>
  </div>
  <div style="padding: 20px; background: #f9fafb; border-radius: 0 0 12px 12px;">
    <p>Hello <strong>%s</strong>,</p>
    <p>Your account has been created successfully.</p>
    <p><strong>Username:</strong> %s</p>
    <p><strong>Data Limit:</strong> %s</p>
    <p><strong>Expiry:</strong> %s</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
    <p style="color: #6b7280; font-size: 12px;">© 2026 VortexUiPro. All rights reserved.</p>
  </div>
</body>
</html>`, data["username"], data["username"], data["data_limit"], data["expiry"])

	case "expiry_warning":
		subject = "Subscription Expiry Warning"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #f59e0b, #ef4444); padding: 30px; border-radius: 12px; text-align: center;">
    <h2 style="color: white; margin: 0;">Subscription Expiring Soon</h2>
  </div>
  <div style="padding: 20px; background: #f9fafb; border-radius: 0 0 12px 12px;">
    <p>Hello <strong>%s</strong>,</p>
    <p>Your subscription will expire on <strong>%s</strong>.</p>
    <p>Please renew to avoid service interruption.</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
    <p style="color: #6b7280; font-size: 12px;">© 2026 VortexUiPro. All rights reserved.</p>
  </div>
</body>
</html>`, data["username"], data["expiry"])

	case "traffic_warning":
		subject = "Traffic Limit Warning"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #f59e0b, #ef4444); padding: 30px; border-radius: 12px; text-align: center;">
    <h2 style="color: white; margin: 0;">Traffic Limit Warning</h2>
  </div>
  <div style="padding: 20px; background: #f9fafb; border-radius: 0 0 12px 12px;">
    <p>Hello <strong>%s</strong>,</p>
    <p>You have used <strong>%s%%</strong> of your data limit.</p>
    <p>Used: %s / %s</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
    <p style="color: #6b7280; font-size: 12px;">© 2026 VortexUiPro. All rights reserved.</p>
  </div>
</body>
</html>`, data["username"], data["usage_pct"], data["used"], data["total"])
	}

	return s.Send(to, subject, body)
}

// SaveConfig saves SMTP settings to the database.
func (s *EmailService) SaveConfig(cfg *SMTPConfig) error {
	settings := map[string]string{
		"smtp_host":     cfg.Host,
		"smtp_port":     fmt.Sprintf("%d", cfg.Port),
		"smtp_username": cfg.Username,
		"smtp_password": cfg.Password,
		"smtp_from":     cfg.From,
	}
	for k, v := range settings {
		if err := database.SetSetting(k, v); err != nil {
			return fmt.Errorf("save smtp setting %s: %w", k, err)
		}
	}
	return nil
}

// ─── Internal ────────────────────────────────────────────────────────

func (s *EmailService) buildMessage(from, to, subject, body string) []byte {
	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s\r\n", from, to, subject, body))
}

func (s *EmailService) sendSTARTTLS(addr string, cfg *SMTPConfig, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	return s.authAndSend(client, cfg, msg)
}

func (s *EmailService) sendTLS(addr string, cfg *SMTPConfig, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	return s.authAndSend(client, cfg, msg)
}

func (s *EmailService) authAndSend(client *smtp.Client, cfg *SMTPConfig, msg []byte) error {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	toList := s.extractTo(msg)
	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	sent := map[string]bool{}
	for _, to := range toList {
		if sent[to] {
			continue
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", to, err)
		}
		sent[to] = true
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return nil
}

func (s *EmailService) extractTo(msg []byte) []string {
	var tos []string
	for _, line := range strings.Split(string(msg), "\r\n") {
		if strings.HasPrefix(line, "To: ") {
			for _, addr := range strings.Split(line[4:], ",") {
				tos = append(tos, strings.TrimSpace(addr))
			}
		}
	}
	return tos
}
