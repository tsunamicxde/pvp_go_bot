package main

import (
	"log"
	"os"

	"telegram_pvp_bot/pkg/database"
	"telegram_pvp_bot/pkg/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Загружаем переменные окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	// Инициализация базы данных
	db := database.InitDB()
	defer db.Close()

	// Инициализация бота
	botToken := os.Getenv("BOT_TOKEN") // Получаем токен из переменной окружения
	if botToken == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}
	telegram.InitBot(botToken)

	// Обработка обновлений
	updates := telegram.GetUpdates()
	for update := range updates {
		if update.CallbackQuery != nil {
			telegram.HandleCallbackQuery(update.CallbackQuery)
			continue
		}
		if update.Message != nil {
			if update.Message.IsCommand() {
				telegram.HandleCommand(&update)
			} else {
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда.")
				telegram.SendMessage(msg)
			}
		}
	}
}
