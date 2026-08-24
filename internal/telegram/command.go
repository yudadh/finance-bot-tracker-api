package telegram

import "strings"

type Command string

const (
	CommandStart  Command = "start"
	CommandHelp   Command = "help"
	CommandToday  Command = "today"
	CommandMonth  Command = "month"
	CommandCancel Command = "cancel"
	CommandReport Command = "report"
	CommandBudget Command = "budget"
	CommandInfo   Command = "info"
	CommandSetBudget Command = "setbudget"
	CommandUnknown Command = ""
)

func DetectCommand(text string) Command {
	normalizeText := strings.ToLower(strings.TrimSpace(text))
	switch normalizeText {
	case "start", "mulai":
		return CommandStart
	case "help", "bantuan":
		return CommandHelp
	case "today", "hari ini":
		return CommandToday
	case "month", "bulan ini":
		return CommandMonth
	case "cancel", "batal", "batalkan":
		return CommandCancel
	case "report", "laporan":
		return CommandReport
	case "budget":
		return CommandBudget
	case "info":
		return CommandInfo
	default:
		if strings.HasPrefix(normalizeText, "setbudget ") {
			return CommandSetBudget
		}
		return CommandUnknown
	}
}
