package telegram

import (
	"fmt"
	"log"

	"telegram_pvp_bot/pkg/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Определяем глобальные переменные
var bot *tgbotapi.BotAPI
var lastMessageID int

// Инициализация бота
func InitBot(token string) {
	var err error

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизация прошла: %s", bot.Self.UserName)
}

// Вывод главного меню
func sendMainMenu(chatID int64) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎲 Играть", "play"),
			tgbotapi.NewInlineKeyboardButtonData("📈 Моя статистика", "stats"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать!\n\nЗдесь вы можете сыграть в игру камень-ножницы-бумага.\nВыберите опцию:")
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)
}

// Функция для обработки команд
func HandleCommand(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		HandleStartCommand(update.Message.Chat.ID)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда.")
		bot.Send(msg)
	}
}

// Функция для обработки нажатий на inline-кнопки
func HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	switch callbackQuery.Data {
	case "play":
		HandlePlayButton(callbackQuery.Message.Chat.ID)
	case "stats":
		HandleStatsButton(callbackQuery.Message.Chat.ID)
	default:
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Произошла ошибка. Попробуйте ещё раз.")
		bot.Send(msg)
	}
}

// Обработчик команды /start
func HandleStartCommand(chatID int64) {
	sendMainMenu(chatID)
	err := database.InsertUserData(chatID) // Используем функцию из database
	if err != nil {
		log.Fatalf("Ошибка при добавлении пользователя: %v", err)
	}
}

// Обработчик кнопки "Играть"
func HandlePlayButton(chatID int64) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)

	msg := tgbotapi.NewMessage(chatID, "Игра началась :)")

	sentMsg, err := bot.Send(msg)
	if err == nil {
		lastMessageID = sentMsg.MessageID // Сохраняем идентификатор отправленного сообщения
	} else {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// Обработчик кнопки "Моя статистика"
func HandleStatsButton(chatID int64) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)
	user, err := database.GetUserStats(chatID) // Используем функцию из database
	if err != nil {
		log.Printf("Пользователь не найден: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить статистику.")
		bot.Send(msg)
		return
	}
	winningRatio := 0.0

	if user.NumberOfWins+user.NumberOfDefeats+user.NumberOfDraws > 0 {
		winningRatio = float64(user.NumberOfWins) / float64(user.NumberOfWins+user.NumberOfDefeats+user.NumberOfDraws)
	}

	text := fmt.Sprintf("📊 Ваша статистика:\n\n✅ Количество побед: %d\n❌ Количество поражений: %d\n🤝 Количество ничьих: %d\n#️⃣ Коэффициент побед: %.2f",
		user.NumberOfWins,
		user.NumberOfDefeats,
		user.NumberOfDraws,
		winningRatio)

	msg := tgbotapi.NewMessage(chatID, text)

	sentMsg, err := bot.Send(msg)
	if err == nil {
		lastMessageID = sentMsg.MessageID // Сохраняем идентификатор отправленного сообщения
	} else {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func GetUpdates() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return bot.GetUpdatesChan(u)
}

func SendMessage(msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// Функция для удаления последнего сообщения
func DeleteLastMessage(chatID int64) {
	if lastMessageID != 0 {
		// Создаем конфигурацию для удаления сообщения
		deleteConfig := tgbotapi.NewDeleteMessage(chatID, lastMessageID)
		_, err := bot.Send(deleteConfig)
		if err != nil {
			log.Printf("Ошибка при удалении сообщения: %v", err)
		}
		lastMessageID = 0 // Сбрасываем идентификатор после удаления
	}
}
