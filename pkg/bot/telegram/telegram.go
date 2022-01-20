package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/fairDataSociety/stampkeeper/pkg/api"
	"github.com/fairDataSociety/stampkeeper/pkg/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	chatId = -645549805
)

type Bot struct {
	bot *tgbotapi.BotAPI
	api *api.Handler

	logger logging.Logger
}

func (b *Bot) Notify(message string) error {
	return b.send(message)
}

func (b *Bot) List() error {
	list := b.api.List()
	listText := ""
	for _, v := range list {
		status := "inactive"
		item := v.(map[string]interface{})
		if item["active"] == true {
			status = "active"
		}
		listText += fmt.Sprintf("\"%s\" (%s) is %v\n", item["name"], item["batch"], status)
	}
	return b.send(listText)
}

func (b *Bot) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string) error {
	if err := b.api.Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval); err != nil {
		return err
	}
	return b.send("started watching batch")
}

func (b *Bot) Unwatch(batchId string) error {
	if err := b.api.Unwatch(batchId); err != nil {
		return err
	}
	return b.send("stopped watching batch")
}

func (b *Bot) send(message string) error {
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := b.bot.Send(msg)
	if err != nil {
		b.logger.Errorf("failed to send message %s", err.Error())
		return err
	}
	return nil
}

func NewBot(ctx context.Context, token string, api *api.Handler, logger logging.Logger) (*Bot, error) {
	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	bot := &Bot{
		bot:    telegramBot,
		api:    api,
		logger: logger,
	}
	updates := telegramBot.GetUpdatesChan(u)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updates:
				if update.Message == nil { // ignore any non-Message updates
					continue
				}
				if !update.Message.IsCommand() { // ignore any non-command Messages
					bot.logger.Debug(update.Message.Text, " is not a command")
					continue
				}
				if update.Message.Chat.ID != chatId {
					bot.logger.Debug("unknown chat")
					continue
				}

				// Extract the command from the Message.
				switch update.Message.Command() {
				case "help":
					helpMessage := `
/version - Stampkeeper version
/list - List of watched stamps
/watch - Watch a batch

	/watch customName batchId balanceEndpoint minBalance topupBalance interval
	
	Please choose a name without spaces

/unwatch - Stop watching a batch
	
	/unwatch batchId

/help - General usage instruction`
					if err := bot.send(helpMessage); err != nil {
						bot.logger.Errorf("failed to send message %s", err.Error())
						continue
					}

				case "list":
					if err := bot.List(); err != nil {
						bot.logger.Errorf("failed to send message %s", err.Error())
						continue
					}

				case "watch":
					args := strings.Split(update.Message.Text, " ")
					if len(args) != 7 {
						if err := bot.send("invalid arguments. aborting"); err != nil {
							bot.logger.Errorf("failed to send message %s", err.Error())
							continue
						}
						continue
					}
					if err := bot.Watch(args[1], args[2], args[3], args[4], args[5], args[6]); err != nil {
						bot.logger.Errorf("failed to watch %s", err.Error())
						if err := bot.send(fmt.Sprintf("failed to watch %s. aborting", err.Error())); err != nil {
							bot.logger.Errorf("failed to send message %s", err.Error())
							continue
						}
						continue
					}

				case "unwatch":
					args := strings.Split(update.Message.Text, " ")
					if len(args) != 2 {
						if err := bot.send("invalid arguments. aborting"); err != nil {
							bot.logger.Errorf("failed to send message %s", err.Error())
							continue
						}
						continue
					}
					if err := bot.Unwatch(args[1]); err != nil {
						bot.logger.Errorf("failed to unwatch %s", err.Error())
						if err := bot.send("invalid arguments. aborting"); err != nil {
							bot.logger.Errorf("failed to send message %s", err.Error())
							continue
						}
					}

				case "version":
					if err := bot.send("v0.0.1"); err != nil {
						bot.logger.Errorf("failed to send message %s", err.Error())
						continue
					}

				default:
					if err := bot.send("I don't know that command"); err != nil {
						bot.logger.Errorf("failed to send message %s", err.Error())
						continue
					}
				}
			}
		}
	}()

	return bot, nil
}
