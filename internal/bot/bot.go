// Package bot wraps the Telegram Bot API to send and edit messages.
package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot is a thin wrapper around the Telegram Bot API client.
type Bot struct {
	api    *tgbotapi.BotAPI
	chatID int64
}

// New creates a Bot using the provided token and target chat ID.
func New(token string, chatID int64) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, chatID: chatID}, nil
}

// SendMessage sends a new text message to the configured chat and returns the
// Telegram message ID assigned to it.
func (b *Bot) SendMessage(text string) (int, error) {
	msg := tgbotapi.NewMessage(b.chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	sent, err := b.api.Send(msg)
	if err != nil {
		return 0, err
	}
	return sent.MessageID, nil
}

// EditMessage replaces the text of an existing message (identified by
// messageID) in the configured chat.
func (b *Bot) EditMessage(messageID int, text string) error {
	edit := tgbotapi.NewEditMessageText(b.chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeHTML
	_, err := b.api.Send(edit)
	return err
}
