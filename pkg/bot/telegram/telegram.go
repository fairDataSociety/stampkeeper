package telegram

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fairDataSociety/stampkeeper/pkg/api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	chatId = -645549805
)

type Bot struct {
	bot *tgbotapi.BotAPI
	api *api.Handler
}

func (b *Bot) Notify(message string) error {
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := b.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) List() error {
	list := b.api.List()
	data, err := json.MarshalIndent(list, "", "\t")
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, string(data))
	_, err = b.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string) error {
	if err := b.api.Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval); err != nil {
		return err
	}
	return nil
}

func (b *Bot) Unwatch(batchId string) error {
	if err := b.api.Unwatch(batchId); err != nil {
		return err
	}
	return nil
}

func NewBot(api *api.Handler) (*Bot, error) {
	token := os.Getenv("TG_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("bot token not available")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{bot: bot, api: api}, nil
}
