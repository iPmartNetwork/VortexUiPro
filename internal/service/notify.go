package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Notification Types ──────────────────────────────────────────────

type NotificationType string

const (
	NotifyTrafficAlert  NotificationType = "traffic_alert"
	NotifyExpiryWarning NotificationType = "expiry_warning"
	NotifySystemAlert   NotificationType = "system_alert"
	NotifyUserCreated   NotificationType = "user_created"
	NotifyPayment       NotificationType = "payment_received"
)

// ─── Telegram Bot Service ──────────────────────────────────────────

// TelegramBot sends notifications to users via Telegram.
type TelegramBot struct {
	mu       sync.RWMutex
	token    string
	client   *http.Client
	enabled  bool
	stopCh   chan struct{}
	queue    []TelegramMessage
}

// TelegramMessage represents a message to send via Telegram.
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// NewTelegramBot creates a new Telegram bot service.
func NewTelegramBot(token string) *TelegramBot {
	return &TelegramBot{
		token:   token,
		client:  &http.Client{Timeout: 15 * time.Second},
		enabled: token != "",
		stopCh:  make(chan struct{}),
		queue:   make([]TelegramMessage, 0, 100),
	}
}

// Start begins processing the message queue.
func (b *TelegramBot) Start() {
	if !b.enabled {
		log.Println("Telegram bot disabled (no token)")
		return
	}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				b.processQueue()
			case <-b.stopCh:
				b.processQueue() // flush remaining
				return
			}
		}
	}()
	log.Println("Telegram bot started")
}

// Stop gracefully stops the bot.
func (b *TelegramBot) Stop() {
	close(b.stopCh)
}

// SendMessage queues a message to be sent to a specific chat.
func (b *TelegramBot) SendMessage(chatID, text string) {
	if !b.enabled {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queue = append(b.queue, TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	})
}

// SendToAll sends a message to all admin Telegram chats.
func (b *TelegramBot) SendToAll(text string) {
	if !b.enabled {
		return
	}
	// Get all notification channels
	channels, err := database.ListNotificationChannels()
	if err != nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range channels {
		if ch.Type == "telegram" && ch.Enabled && ch.ChatID != "" {
			b.queue = append(b.queue, TelegramMessage{
				ChatID:    ch.ChatID,
				Text:      text,
				ParseMode: "HTML",
			})
		}
	}
}

// processQueue sends all queued messages.
func (b *TelegramBot) processQueue() {
	b.mu.Lock()
	if len(b.queue) == 0 {
		b.mu.Unlock()
		return
	}
	batch := b.queue
	b.queue = make([]TelegramMessage, 0, 100)
	b.mu.Unlock()

	for _, msg := range batch {
		if err := b.send(msg); err != nil {
			log.Printf("Telegram send error: %v", err)
			// Re-queue failed messages
			b.mu.Lock()
			b.queue = append(b.queue, msg)
			b.mu.Unlock()
		}
	}
}

// send sends a single message via the Telegram Bot API.
func (b *TelegramBot) send(msg TelegramMessage) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

// ─── Notification Templates ──────────────────────────────────────────

// FormatTrafficAlert formats a traffic usage alert message.
func FormatTrafficAlert(username string, usagePercent float64, usedBytes, limitBytes int64) string {
	return fmt.Sprintf(
		"🚨 <b>Traffic Alert</b>\n\n"+
			"User: <code>%s</code>\n"+
			"Usage: <b>%.1f%%</b>\n"+
			"Used: <b>%d MB</b> / %d MB\n\n"+
			"⚠️ Please top up your account to avoid service interruption.",
		username, usagePercent, usedBytes/1024/1024, limitBytes/1024/1024,
	)
}

// FormatExpiryWarning formats an expiry warning message.
func FormatExpiryWarning(username string, daysLeft int, expiryTime int64) string {
	expiryDate := time.UnixMilli(expiryTime).Format("2006-01-02")
	return fmt.Sprintf(
		"⏰ <b>Expiry Warning</b>\n\n"+
			"User: <code>%s</code>\n"+
			"Days left: <b>%d</b>\n"+
			"Expires: <b>%s</b>\n\n"+
			"🔄 Please renew your subscription to continue using the service.",
		username, daysLeft, expiryDate,
	)
}

// FormatPaymentNotification formats a payment received message.
func FormatPaymentNotification(username string, amount int64, currency string, planName string) string {
	return fmt.Sprintf(
		"✅ <b>Payment Received</b>\n\n"+
			"User: <code>%s</code>\n"+
			"Amount: <b>%d %s</b>\n"+
			"Plan: <b>%s</b>\n\n"+
			"🙏 Thank you for your purchase!",
		username, amount, currency, planName,
	)
}

// FormatSystemAlert formats a system alert message.
func FormatSystemAlert(alertType, message string) string {
	return fmt.Sprintf(
		"🔴 <b>System Alert: %s</b>\n\n%s\n\n⏰ %s",
		alertType, message, time.Now().Format("2006-01-02 15:04:05"),
	)
}

// ─── Notification Service ────────────────────────────────────────────

// NotificationService manages all notification channels.
type NotificationService struct {
	bot *TelegramBot
}

// NewNotificationService creates a new notification service.
func NewNotificationService(bot *TelegramBot) *NotificationService {
	return &NotificationService{bot: bot}
}

// NotifyTraffic sends a traffic alert notification.
func (ns *NotificationService) NotifyTraffic(username, chatID string, usagePercent float64, used, limit int64) {
	msg := FormatTrafficAlert(username, usagePercent, used, limit)
	if chatID != "" {
		ns.bot.SendMessage(chatID, msg)
	}
	ns.bot.SendToAll(msg)
}

// NotifyExpiry sends an expiry warning notification.
func (ns *NotificationService) NotifyExpiry(username, chatID string, daysLeft int, expiryTime int64) {
	msg := FormatExpiryWarning(username, daysLeft, expiryTime)
	if chatID != "" {
		ns.bot.SendMessage(chatID, msg)
	}
	ns.bot.SendToAll(msg)
}

// NotifyPayment sends a payment notification.
func (ns *NotificationService) NotifyPayment(username, chatID string, amount int64, currency, planName string) {
	msg := FormatPaymentNotification(username, amount, currency, planName)
	if chatID != "" {
		ns.bot.SendMessage(chatID, msg)
	}
	ns.bot.SendToAll(msg)
}

// NotifySystem sends a system alert.
func (ns *NotificationService) NotifySystem(alertType, message string) {
	msg := FormatSystemAlert(alertType, message)
	ns.bot.SendToAll(msg)
}
