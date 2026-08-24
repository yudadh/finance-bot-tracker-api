package telegram

import (
	"context"
	"log/slog"

	tgBot "github.com/go-telegram/bot"
)

type Bot struct {
	client *tgBot.Bot
	logger *slog.Logger
}

func NewBot(token string, handler *BotHandler, logger *slog.Logger) (*Bot, error) {
	bot, err := tgBot.New(token, tgBot.WithDefaultHandler(handler.HandleUpdate))
	if err != nil {
		logger.Error("error when initializing bot", "error", err.Error())
		return nil, err
	}

	return &Bot{
		client: bot,
		logger: logger,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	if b.logger != nil {
		b.logger.Info("starting telegram bot")
	}
	b.client.Start(ctx)
}