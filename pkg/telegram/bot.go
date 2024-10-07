package telegram

import (
	"fmt"
	"log"
	"time"

	"telegram_pvp_bot/pkg/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/rand"
)

// Определяем глобальные переменные

// Переменная бота
var bot *tgbotapi.BotAPI

// ID последнего сообщения от бота
var lastMessageID int

// Карта ходов игрока
var Moves = map[string]int{
	"👊":  0,
	"✌️": 1,
	"✋":  2,
}

// Матрица результата игры
// 0 - ничья, 1 - победа игрока, 2 - победа бота
var resultMatrix = [3][3]int{
	{0, 1, 2}, // Игрок выбирает 0
	{2, 0, 1}, // Игрок выбирает 1
	{1, 2, 0}, // Игрок выбирает 2
}

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
func sendMainMenu(chatID int64, welcomeText string) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎲 Играть", "play"),
			tgbotapi.NewInlineKeyboardButtonData("📈 Моя статистика", "stats"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, welcomeText+"Выберите опцию:")
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)
}

// Вывод игрового меню
func sendPlayMenu(chatID int64) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👊", "👊"),
			tgbotapi.NewInlineKeyboardButtonData("✌️", "✌️"),
			tgbotapi.NewInlineKeyboardButtonData("✋", "✋"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Выберите ход:")
	msg.ReplyMarkup = inlineKeyboard

	sentMsg, err := bot.Send(msg)
	if err == nil {
		lastMessageID = sentMsg.MessageID // Сохраняем идентификатор отправленного сообщения
	} else {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
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
	case "👊", "✌️", "✋":
		HandleMoveButton(callbackQuery)
	default:
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Произошла ошибка. Попробуйте ещё раз.")
		bot.Send(msg)
	}
}

// Обработчик команды /start
func HandleStartCommand(chatID int64) {
	sendMainMenu(chatID, "Добро пожаловать!\n\nЗдесь вы можете сыграть в игру камень-ножницы-бумага.\n")
	err := database.InsertUserData(chatID) // Используем функцию из database
	if err != nil {
		log.Fatalf("Ошибка при добавлении пользователя: %v", err)
	}
}

// Обработчик кнопки "Играть"
func HandlePlayButton(chatID int64) {
	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)

	sendPlayMenu(chatID)
}

// Обработчик хода игрока
func HandleMoveButton(callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID

	// Удаляем предыдущее сообщение, если оно существует
	DeleteLastMessage(chatID)

	move := callbackQuery.Data

	botMoveInt := generateRandomMove()
	botMove, isMapContainsKey := findKeyByValue(Moves, botMoveInt)

	if !isMapContainsKey {
		return
	}

	moveInt := Moves[move]

	result := resultMatrix[moveInt][botMoveInt]

	text := handleGameResult(chatID, botMove, result)

	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := bot.Send(msg)
	sendMainMenu(chatID, "Сыграть ещё раз?\n")
	if err == nil {
		lastMessageID = sentMsg.MessageID // Сохраняем идентификатор отправленного сообщения
	} else {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// Обработка результата игры
func handleGameResult(chatID int64, botMove string, gameCode int) string {
	user, err := database.GetUserStats(chatID) // Используем функцию из database
	if err != nil {
		log.Printf("Пользователь не найден: %v", err)
		return "Произошла ошибка. Попробуйте ещё раз."
	}

	text := ""
	suffix := fmt.Sprintf("\nХод бота: %s", botMove)

	switch gameCode {
	case 0: // Ничья
		text = "Ничья!" + suffix
		user.NumberOfDraws++
	case 1: // Победа игрока
		text = "Вы выиграли!" + suffix
		user.NumberOfWins++
	case 2: // Победа бота
		text = "Вы проиграли!" + suffix
		user.NumberOfDefeats++
	default:
		text = "Произошла ошибка. Попробуйте ещё раз."
	}

	err = database.UpdateUserStats(chatID, user)
	if err != nil {
		log.Printf("Данные пользователя не удалось обновить: %v", err)
		return "Произошла ошибка. Попробуйте ещё раз."
	}
	return text
}

// Генерация случайного хода бота
func generateRandomMove() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	randomNum := rand.Intn(3)
	return randomNum
}

// Функция для поиска ключа по значению в словаре
func findKeyByValue(myMap map[string]int, value int) (string, bool) {
	for key, val := range myMap {
		if val == value {
			return key, true
		}
	}
	return "", false
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
