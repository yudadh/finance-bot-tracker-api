package telegram

import (
	"context"
	"errors"
	"log/slog"
	"time"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type transactionCreator interface {
	CreateFromText(ctx context.Context, transaction service.CreateTransactionFromTextInput) (*service.CreateTransactionFromTextResult, error)
}

type userFinderOrCreator interface {
	FindOrCreate(ctx context.Context, userInput *service.FindOrCreateUserInput) (*domain.User, error)
}

type budgetFinderAndCreator interface {
	SetBudget(ctx context.Context, budgetInput service.CreateBudgetInput) (*service.CreateBudgetResult, error)
	GetBudgetStatus(ctx context.Context, user service.BudgetStatusInput) (*service.BudgetStatus, error)
	CheckBudgetAlert(ctx context.Context, input service.BudgetStatusInput) (*service.CheckBudgetAlertResult, error)
	UpdateBudgetTimestamp(
		ctx context.Context,
		budgetID uint64,
		threshold uint8,
		now time.Time,
	) error
}

type BotHandler struct {
	transactionService transactionCreator
	userService        userFinderOrCreator
	budgetService      budgetFinderAndCreator
	logger             *slog.Logger
}

func NewBotHandler(
	transactionService transactionCreator,
	userService userFinderOrCreator,
	budgetService budgetFinderAndCreator,
	logger *slog.Logger,
) *BotHandler {
	return &BotHandler{
		transactionService: transactionService,
		userService:        userService,
		budgetService: budgetService,
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

	h.SendTimezoneSelection(ctx, b, chatID)

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

	switch DetectCommand(text) {
	case CommandStart:
		h.sendMessage(ctx, b, chatID, welcomeMessage())
		return

	case CommandHelp:
		h.sendMessage(ctx, b, chatID, helpMessage())
		return

	case CommandBudget:
		loc, err := time.LoadLocation(user.Timezone)
		if err != nil {
			h.sendMessage(ctx, b, chatID, InternalErrorMessage())
			return
		}

		budgetStatus, err := h.budgetService.GetBudgetStatus(ctx, service.BudgetStatusInput{
			UserID:     user.ID,
			PeriodType: domain.BudgetPeriodTypeMonthly,
			Location:   loc,
		})

		if err != nil {
			h.sendMessage(ctx, b, chatID, InternalErrorMessage())
			return
		}

		h.sendMessage(ctx, b, chatID, budgetStatusMessage(budgetStatus))
		return

	case CommandSetBudget:
		res, err := parser.ParseBudget(text)

		if err != nil {
			if errors.Is(err, parser.ErrInvalidBudgetText) {
				h.sendMessage(ctx, b, chatID, parseErrorMessage(CmdTypeBudget))
				return
			}

			h.sendMessage(ctx, b, chatID, InternalErrorMessage())
			return
		}

		loc, err := time.LoadLocation(user.Timezone)

		if err != nil {
			h.sendMessage(ctx, b, chatID, InternalErrorMessage())
			return
		}

		budget, err := h.budgetService.SetBudget(ctx, service.CreateBudgetInput{
			UserID:     user.ID,
			PeriodType: res.PeriodType,
			Now:        time.Now(),
			Amount:     uint64(res.Amount),
			Currency:   "IDR",
			Loc:        loc,
		})
		
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				h.sendMessage(ctx, b, chatID, budgetAlreadyExistsMessage())
				return
			}

			h.sendMessage(ctx, b, chatID, InternalErrorMessage())
			return
		}

		h.sendMessage(ctx, b, chatID, budgetCreatedMessage(budget))
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
		h.sendMessage(ctx, b, chatID, parseErrorMessage(CmdTypeTransaction))
		return
	}

	h.sendMessage(ctx, b, chatID, transactionSavedMessage(result.Transaction, result.CategoryName))

}

func (h *BotHandler) sendMessage(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
	text string,
) error {
	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	if err != nil && h.logger != nil {
		h.logger.Error("failed to send telegram message", "error", err)
	}
	
	return err
}

const (
	callbackTimezoneWIB  = "timezone:wib"
	callbackTimezoneWITA = "timezone:wita"
	callbackTimezoneWIT  = "timezone:wit"
)

func (h *BotHandler) SendTimezoneSelection(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
) error {
	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text:   "Pilih zona waktu kamu:",
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "WIB",
						CallbackData: callbackTimezoneWIB,
					},
					{
						Text:         "WITA",
						CallbackData: callbackTimezoneWITA,
					},
					{
						Text:         "WIT",
						CallbackData: callbackTimezoneWIT,
					},
				},
			},
		},
	})

	return err
}
