package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── Bot Command Types ───────────────────────────────────────────────

// BotCommand represents a Telegram bot command.
type BotCommand struct {
	Command     string
	Description string
	Handler     func(chatID int64, args string) string
}

// ─── Enhanced Telegram Bot ──────────────────────────────────────────

// EnhancedBot extends TelegramBot with command handling and event subscriptions.
type EnhancedBot struct {
	*TelegramBot

	mu       sync.RWMutex
	commands []BotCommand

	// Webhook / polling
	webhookURL string
	polling    bool

	// Event handlers
	onTrafficAlert func(chatID int64, username string, usagePercent float64, used, limit int64)
	onExpiryWarn   func(chatID int64, username string, daysLeft int)
	onUserCreate   func(chatID int64, username string)
	onPayment      func(chatID int64, username string, amount float64, currency string)
}

// NewEnhancedBot creates an enhanced Telegram bot.
func NewEnhancedBot(token string) *EnhancedBot {
	eb := &EnhancedBot{
		TelegramBot: NewTelegramBot(token),
		commands:    make([]BotCommand, 0),
	}
	eb.RegisterDefaultCommands()
	return eb
}

// RegisterDefaultCommands registers the default set of bot commands.
func (eb *EnhancedBot) RegisterDefaultCommands() {
	cmds := []BotCommand{
		{
			Command:     "/start",
			Description: "Show welcome message and available commands",
			Handler:     eb.handleStart,
		},
		{
			Command:     "/help",
			Description: "Show available commands",
			Handler:     eb.handleHelp,
		},
		{
			Command:     "/status",
			Description: "Show panel status overview",
			Handler:     eb.handleStatus,
		},
		{
			Command:     "/users",
			Description: "Show user count and basic stats",
			Handler:     eb.handleUsers,
		},
		{
			Command:     "/traffic",
			Description: "Show total traffic usage",
			Handler:     eb.handleTraffic,
		},
		{
			Command:     "/sysinfo",
			Description: "Show system information (CPU, memory)",
			Handler:     eb.handleSysInfo,
		},
	}

	for _, cmd := range cmds {
		eb.RegisterCommand(cmd)
	}
}

// RegisterCommand adds a new command handler.
func (eb *EnhancedBot) RegisterCommand(cmd BotCommand) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for i, c := range eb.commands {
		if c.Command == cmd.Command {
			eb.commands[i] = cmd
			return
		}
	}
	eb.commands = append(eb.commands, cmd)
}

// StartWithPolling starts the bot with long-polling for updates.
func (eb *EnhancedBot) StartWithPolling() {
	if !eb.enabled {
		return
	}
	eb.polling = true
	go eb.pollLoop()
	log.Println("Enhanced Telegram bot started with polling")
}

// SetWebhookURL sets the webhook URL for receiving updates.
func (eb *EnhancedBot) SetWebhookURL(url string) {
	eb.webhookURL = url
}

// ProcessUpdate processes an incoming Telegram update (from webhook or polling).
func (eb *EnhancedBot) ProcessUpdate(update map[string]any) {
	// Extract message
	message, ok := update["message"].(map[string]any)
	if !ok {
		return
	}

	// Extract chat ID
	chat, ok := message["chat"].(map[string]any)
	if !ok {
		return
	}
	chatID, _ := chat["id"].(float64)

	// Extract text
	text, ok := message["text"].(string)
	if !ok || text == "" {
		return
	}

	eb.processCommand(int64(chatID), text)
}

// processCommand parses and executes a bot command.
func (eb *EnhancedBot) processCommand(chatID int64, text string) {
	text = strings.TrimSpace(text)

	// Check for commands
	if !strings.HasPrefix(text, "/") {
		// Not a command, try sending help
		if chatID != 0 {
			eb.SendMessage(fmt.Sprintf("%d", chatID), eb.handleHelp(chatID, ""))
		}
		return
	}

	// Parse command and arguments
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	cmdName := strings.ToLower(parts[0])
	args := strings.Join(parts[1:], " ")

	eb.mu.RLock()
	for _, cmd := range eb.commands {
		if cmd.Command == cmdName || strings.HasPrefix(cmdName, cmd.Command+"@") {
			eb.mu.RUnlock()
			response := cmd.Handler(chatID, args)
			if response != "" && chatID != 0 {
				eb.SendMessage(fmt.Sprintf("%d", chatID), response)
			}
			return
		}
	}
	eb.mu.RUnlock()

	// Unknown command
	if chatID != 0 {
		eb.SendMessage(fmt.Sprintf("%d", chatID), fmt.Sprintf(
			"❌ Unknown command: <code>%s</code>\n\nUse /help to see available commands.", cmdName,
		))
	}
}

// pollLoop polls Telegram for updates.
func (eb *EnhancedBot) pollLoop() {
	offset := 0
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		select {
		case <-eb.stopCh:
			return
		default:
		}

		url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=25", eb.token, offset)
		resp, err := client.Get(url)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		var result struct {
			OK     bool                   `json:"ok"`
			Result []map[string]any       `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		resp.Body.Close()

		if !result.OK {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range result.Result {
			eb.ProcessUpdate(update)
			if id, ok := update["update_id"].(float64); ok {
				offset = int(id) + 1
			}
		}
	}
}

// ─── Command Handlers ────────────────────────────────────────────────

func (eb *EnhancedBot) handleStart(chatID int64, args string) string {
	return fmt.Sprintf(`🤖 <b>VortexUiPro Bot</b>

Welcome to the VortexUiPro management bot! I can help you monitor your panel and receive notifications.

<b>Available Commands:</b>
/help — Show all commands
/status — Panel status overview
/users — User statistics
/traffic — Traffic usage
/sysinfo — System information

<b>Notifications:</b>
You'll automatically receive alerts for:
• 🚨 Traffic limit warnings
• ⏰ Expiry reminders
• ✅ Payment notifications
• 🔴 System alerts

Use the panel settings to configure your notification preferences.`)
}

func (eb *EnhancedBot) handleHelp(chatID int64, args string) string {
	eb.mu.RLock()
	cmdList := make([]BotCommand, len(eb.commands))
	copy(cmdList, eb.commands)
	eb.mu.RUnlock()

	var b strings.Builder
	b.WriteString("🤖 <b>VortexUiPro Bot Commands</b>\n\n")

	for _, cmd := range cmdList {
		b.WriteString(fmt.Sprintf("<code>%-15s</code> %s\n", cmd.Command, cmd.Description))
	}

	b.WriteString("\n<i>💡 Send any message to see this help.</i>")
	return b.String()
}

func (eb *EnhancedBot) handleStatus(chatID int64, args string) string {
	return `📊 <b>Panel Status</b>

<i>Status information would be fetched from the panel database.</i>

• Panel: VortexUiPro
• Version: 0.0.1
• Status: ✅ Online

<i>Run this command in the panel admin chat for detailed stats.</i>`
}

func (eb *EnhancedBot) handleUsers(chatID int64, args string) string {
	return `👥 <b>User Statistics</b>

<i>Connect to the panel database for live user stats.</i>

• Total Users: —
• Active Users: —
• Expired Users: —

<i>Use the panel admin interface for detailed user management.</i>`
}

func (eb *EnhancedBot) handleTraffic(chatID int64, args string) string {
	return `📈 <b>Traffic Usage</b>

<i>Traffic data available when connected to an active core.</i>

• Total Upload: —
• Total Download: —
• Today's Usage: —

<i>Use the panel dashboard for real-time traffic charts.</i>`
}

func (eb *EnhancedBot) handleSysInfo(chatID int64, args string) string {
	return `🖥️ <b>System Information</b>

<i>System information is available when the monitor is active.</i>

• CPU: —
• Memory: —
• Uptime: —

<i>Check the panel settings for system monitoring configuration.</i>`
}

// ─── Setup Commands for Webhook ──────────────────────────────────────

// SetupWebhook configures the bot to receive updates via webhook.
func (eb *EnhancedBot) SetupWebhook(webhookURL string) error {
	eb.webhookURL = webhookURL

	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", eb.token)
	payload := map[string]string{
		"url": webhookURL,
	}

	data, _ := json.Marshal(payload)
	resp, err := eb.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.OK {
		return fmt.Errorf("telegram webhook setup failed: %s", result.Description)
	}

	log.Printf("Telegram webhook set to: %s", webhookURL)
	return nil
}

// DeleteWebhook removes the webhook configuration.
func (eb *EnhancedBot) DeleteWebhook() error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", eb.token)
	resp, err := eb.client.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetWebhookInfo returns the current webhook status.
func (eb *EnhancedBot) GetWebhookInfo() (map[string]any, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", eb.token)
	resp, err := eb.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool           `json:"ok"`
		Result map[string]any `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────

// ParseChatID parses a chat ID from string.
func ParseChatID(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// FormatHTML escapes HTML special characters for Telegram messages.
func FormatHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// ParseCommandArgs parses command arguments in the format key=value.
func ParseCommandArgs(args string) map[string]string {
	result := make(map[string]string)
	re := regexp.MustCompile(`(\w+)=("[^"]*"|\S+)`)

	matches := re.FindAllStringSubmatch(args, -1)
	for _, match := range matches {
		if len(match) == 3 {
			value := strings.Trim(match[2], "\"")
			result[match[1]] = value
		}
	}
	return result
}

// SplitMessage splits a long message into chunks of max 4096 chars (Telegram limit).
func SplitMessage(text string) []string {
	const maxLen = 4096
	var parts []string

	for len(text) > 0 {
		if len(text) <= maxLen {
			parts = append(parts, text)
			break
		}

		// Try to split at newline
		idx := strings.LastIndex(text[:maxLen], "\n")
		if idx < 0 {
			idx = maxLen
		}
		parts = append(parts, text[:idx])
		text = text[idx:]
	}
	return parts
}
