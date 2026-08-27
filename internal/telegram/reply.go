package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type CmdType string

const (
	CmdTypeTransaction CmdType = "transaction"
	CmdTypeBudget      CmdType = "budget"
)

func welcomeMessage() string {
	return `👋 Halo! Selamat datang di Berry Finance Tracker.

💸 Kirim transaksi seperti:
makan siang 25000`
}

func helpMessage() string {
	return `👋 Halo!
Aku bisa bantu mencatat dan memantau keuangan kamu.

📌 Command yang tersedia:

🚀 start / mulai - Memulai bot
❓ help / bantuan - Melihat bantuan
📅 today / hari ini - Ringkasan hari ini
🗓️ month / bulan ini - Ringkasan bulan ini
📊 report / laporan - Laporan keuangan
🎯 budget - Melihat budget
💰 setbudget [bulanan/monthly] [jumlah] - Mengatur budget bulanan
💰 setbudget [mingguan/weekly] [jumlah] - Mengatur budget mingguan
ℹ️ info - Informasi tracker
❌ cancel / batal / batalkan - Membatalkan proses

📝 Contoh:
setbudget bulanan 3000000`
}

func transactionSavedMessage(transaction *domain.Transaction, categoryName string) string {
	category := categoryName
	if category == "" {
		category = "Uncategorized"
	}

	return fmt.Sprintf(
		"✅ Transaksi tersimpan\n\n📌 Tipe: %s\n💸 Jumlah: Rp%d\n🏷️ Kategori: %s",
		transaction.Type,
		transaction.Amount,
		category,
	)
}

func budgetStatusMessage(budgetStatus *service.BudgetStatus) string {
	if budgetStatus == nil {
		return "⚠️ Budget belum ditemukan.\n\n💰 Gunakan setbudget [periode] [jumlah] untuk mengatur budget."
	}

	usedAmount := uint64(0)
	if budgetStatus.TotalTransactionAmount > 0 {
		usedAmount = uint64(budgetStatus.TotalTransactionAmount)
	}

	remainingAmount := int64(budgetStatus.BudgetAmount) - int64(usedAmount)
	usedPercentage := float64(0)
	if budgetStatus.BudgetAmount > 0 {
		usedPercentage = float64(usedAmount) / float64(budgetStatus.BudgetAmount) * 100
	}

	statusText := ""
	if remainingAmount < 0 {
		statusText = fmt.Sprintf(
			"🚨 Status: Melebihi budget Rp%s",
			formatIDR(uint64(-remainingAmount)),
		)
	} else {
		statusText = fmt.Sprintf(
			"✅ Status: Sisa budget Rp%s",
			formatIDR(uint64(remainingAmount)),
		)
	}

	return fmt.Sprintf(
		`🎯 Status Budget

📅 Periode: %s - %s
💰 Budget: Rp%s
💸 Pengeluaran: Rp%s
📊 Terpakai: %.1f%%
%s`,
		formatDate(budgetStatus.PeriodStart),
		formatDate(budgetStatus.PeriodEnd.Add(-time.Nanosecond)),
		formatIDR(budgetStatus.BudgetAmount),
		formatIDR(usedAmount),
		usedPercentage,
		statusText,
	)
}

func budgetCreatedMessage(budget *service.CreateBudgetResult) string {
	if budget == nil {
		return "✅ Budget berhasil dibuat."
	}

	return fmt.Sprintf(
		`✅ Budget berhasil dibuat

📌 Periode: %s
📅 Tanggal: %s - %s
💰 Jumlah: %s %s`,
		formatBudgetPeriodType(budget.PeriodType),
		formatDate(budget.PeriodStart),
		formatDate(budget.PeriodEnd.Add(-time.Nanosecond)),
		budget.Currency,
		formatIDR(budget.Amount),
	)
}

func budgetAlreadyExistsMessage() string {
	return "⚠️ Budget untuk periode ini sudah ada.\n\n🎯 Ketik budget untuk melihat status budget kamu."
}

func budgetAlertMessage(
	budgetAlert *service.CheckBudgetAlertResult,
	periodType domain.BudgetPeriodType,
) string {
	if budgetAlert == nil || !budgetAlert.ShouldAlert {
		return ""
	}

	usedAmount := uint64(0)
	if budgetAlert.TotalTransactionAmount > 0 {
		usedAmount = uint64(budgetAlert.TotalTransactionAmount)
	}

	remainingAmount := int64(budgetAlert.BudgetAmount) - int64(usedAmount)
	statusText := ""
	if remainingAmount < 0 {
		statusText = fmt.Sprintf("🚨 Status: Budget terlewati Rp%s.", formatIDR(uint64(-remainingAmount)))
	} else {
		statusText = fmt.Sprintf("✅ Status: Sisa budget Rp%s.", formatIDR(uint64(remainingAmount)))
	}

	return fmt.Sprintf(
		`⚠️ Reminder Budget %s

📊 Pemakaian: %d%% dari budget %s
🎯 Budget: Rp%s
💸 Pengeluaran: Rp%s
%s`,
		formatBudgetPeriodType(periodType),
		budgetAlert.UsedPercentage,
		formatBudgetPeriodType(periodType),
		formatIDR(budgetAlert.BudgetAmount),
		formatIDR(usedAmount),
		statusText,
	)
}

func formatBudgetPeriodType(periodType domain.BudgetPeriodType) string {
	switch periodType {
	case domain.BudgetPeriodTypeWeekly:
		return "Mingguan"
	case domain.BudgetPeriodTypeMonthly:
		return "Bulanan"
	default:
		return string(periodType)
	}
}

func formatDate(date time.Time) string {
	return date.Format("02 Jan 2006")
}

func formatIDR(amount uint64) string {
	value := strconv.FormatUint(amount, 10)
	if len(value) <= 3 {
		return value
	}

	var builder strings.Builder
	firstGroupLength := len(value) % 3
	if firstGroupLength == 0 {
		firstGroupLength = 3
	}

	builder.WriteString(value[:firstGroupLength])
	for i := firstGroupLength; i < len(value); i += 3 {
		builder.WriteString(".")
		builder.WriteString(value[i : i+3])
	}

	return builder.String()
}

func InternalErrorMessage() string {
	return "🚧 Mohon maaf, sedang terjadi masalah.\n\n🔄 Coba beberapa saat lagi."
}

func parseErrorMessage(cmdType CmdType) string {
	switch cmdType {
	case CmdTypeTransaction:
		return "⚠️ Maaf, saya belum bisa membaca transaksi itu.\n\n💸 Coba format seperti:\nmakan siang 25000"

	case CmdTypeBudget:
		return "⚠️ Maaf, saya belum bisa membaca budget itu.\n\n💰 Coba format seperti:\nsetbudget bulanan 1000000\n\n📌 Pilihan periode: mingguan / bulanan"

	default:
		return ""
	}
}

func unknownCommandMessage() string {
	return `❓ Maaf, perintah tidak diketahui.

📌 Ketik "help" untuk melihat daftar perintah.`
}
