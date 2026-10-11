package telegram

import (
	"context"
	"log/slog"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

)

type Reply interface {
	Text(
		ctx context.Context,
		chatID int64,
		text string,
	) error
	Inline(
		ctx context.Context,
		chatID int64,
		text string,
		keyboard models.InlineKeyboardMarkup,
	) error
}

type ReplyFactory struct {
	logger *slog.Logger
}

func NewReplyFactory(logger *slog.Logger) *ReplyFactory {
	return &ReplyFactory{logger: logger}
}

func (f *ReplyFactory) For(bot *tgBot.Bot) Reply {
	return &TelegramReply{
		bot: bot,
		logger: f.logger,
	}
}

type TelegramReply struct {
	bot *tgBot.Bot
	logger *slog.Logger
}

func (r *TelegramReply) Text(
	ctx context.Context, 
	chatID int64, 
	text string,
) error {
	_, err := r.bot.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text: text,
	})

	if err != nil {
		r.logger.ErrorContext(ctx, "failed to send telegram reply text", "error", err)
	}

	return err
}

func (r *TelegramReply) Inline(
	ctx context.Context,
	chatID int64,
	text string,
	keyboard models.InlineKeyboardMarkup,
) error {
	_, err := r.bot.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text: text,
		ReplyMarkup: keyboard,
	})

	if err != nil {
		r.logger.ErrorContext(ctx, "failed to send telegram reply inline keyboard", "error", err)
	}

	return err
}