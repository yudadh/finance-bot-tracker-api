package telegram

import (
	"context"
	"log/slog"
	"time"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type transactionCreator interface {
	CreateFromText(ctx context.Context, transaction service.CreateTransactionFromTextInput) (*service.CreateTransactionFromTextResult, error)
}

type userFinderOrCreator interface {
	FindOrCreate(ctx context.Context, userInput *service.FindOrCreateUserInput) (*domain.User, error)
}

type BotHandler struct {
	transactionService transactionCreator
	userService        userFinderOrCreator
	logger             *slog.Logger
}

func NewBotHandler(
	transactionService transactionCreator,
	userService userFinderOrCreator,
	logger *slog.Logger,
) *BotHandler {
	return &BotHandler{
		transactionService: transactionService,
		userService:        userService,
		logger:             logger,
	}
}

func (h *BotHandler) HandleUpdate(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	text := update.Message.Text
	chatID := update.Message.Chat.ID

	switch DetectCommand(text) {
	case CommandStart:
		h.sendMessage(ctx, b, chatID, welcomeMessage())
		return
	case CommandHelp:
		h.sendMessage(ctx, b, chatID, helpMessage())
		return
	}

	userInput := &service.FindOrCreateUserInput{
		TelegramID:       update.Message.From.ID,
		TelegramUsername: update.Message.From.Username,
		FirstName:        update.Message.From.FirstName,
		LastName:         update.Message.From.LastName,
		LanguageCode:     update.Message.From.LanguageCode,
	}

	user, err := h.userService.FindOrCreate(ctx, userInput)
	if err != nil {
		h.sendMessage(ctx, b, chatID, InternalErrorMessage())
		return
	}

	transactionInput := service.CreateTransactionFromTextInput{
		UserID:   user.ID,
		Text:     text,
		Now:      time.Now(),
		Currency: "IDR",
	}

	result, err := h.transactionService.CreateFromText(ctx, transactionInput)
	if err != nil {
		h.sendMessage(ctx, b, chatID, parseErrorMessage())
		return
	}

	h.sendMessage(ctx, b, chatID, transactionSavedMessage(result.Transaction, result.CategoryName))

}

func (h *BotHandler) sendMessage(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
	text string,
) {
	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID, 
		Text: text,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	if err != nil && h.logger != nil {
		h.logger.Error("failed to send telegram message", "error", err)
	}
}
