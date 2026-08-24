package telegram

import (
	"fmt"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func welcomeMessage() string {
	return `Halo! selamat datang di berry finance tracker, kirim transaksi seperti: 
makan siang 25000`
}

func helpMessage() string {
	return `Halo! 👋
		Aku bisa bantu mencatat dan memantau keuangan kamu.

		Command yang tersedia:

		• start / mulai - Memulai bot
		• help / bantuan - Melihat bantuan
		• today / hari ini - Ringkasan hari ini
		• month / bulan ini - Ringkasan bulan ini
		• report / laporan - Laporan keuangan
		• budget - Melihat budget
		• setbudget [jumlah] - Mengatur budget bulanan
		• info - Informasi tracker
		• cancel / batal / batalkan - Membatalkan proses

		Contoh:
		setbudget 3000000`
}

func transactionSavedMessage(transaction *domain.Transaction, categoryName string) string {
	category := categoryName
	if category == "" {
		category = "Uncategorized"
	}

	return fmt.Sprintf(
		"Tersimpan: %s Rp%d untuk %s",
		transaction.Type,
		transaction.Amount,
		category,
	)
}

func InternalErrorMessage() string {
	return "Mohon maaf sedang terjadi masalah, coba beberapa saat lagi"
}

func parseErrorMessage() string {
	return "Maaf, saya belum bisa membaca transaksi itu. Coba format seperti: makan siang 25000"
}

func unknownCommandMessage() string {
	return `Maaf perintah tidak diketahui, ketik "help" untuk mengecek list perintah`
}