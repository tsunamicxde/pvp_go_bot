package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"telegram_pvp_bot/pkg/user"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Глобальная переменная базы данных
var db *sql.DB

// Подключение к базе данных
func InitDB() *sql.DB {
	var err error
	// Загружаем переменные окружения из файла .env
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	// Получаем параметры подключения из переменных окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	fmt.Println("Успешное подключение к базе данных")

	// Вызов функции createUsersTable()
	CreateUsersTable()

	return db
}

// Функция для создания таблицы Users
func CreateUsersTable() {
	// SQL-запрос на создание таблицы
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id BIGINT UNIQUE NOT NULL,
		number_of_wins INT DEFAULT 0 NOT NULL,
		number_of_defeats INT DEFAULT 0 NOT NULL,
		number_of_draws INT DEFAULT 0 NOT NULL
	);`

	// Выполнение SQL-запроса
	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %v", err)
	}

	fmt.Println("Таблица успешно создана")
}

// Добавление пользователя в БД
func InsertUserData(chatID int64) error {
	// SQL-запрос на получение пользователя
	query := "SELECT user_id FROM users WHERE user_id = $1"
	row := db.QueryRow(query, chatID)

	var userID int64
	err := row.Scan(&userID)

	// Если пользователь не найден - добавляем в БД
	if err != nil {
		if err == sql.ErrNoRows {
			query = "INSERT INTO users (user_id) VALUES ($1)"
			_, err = db.Exec(query, chatID)
		}
	}
	return err
}

// Получение статистики пользователя
func GetUserStats(chatID int64) (*user.User, error) {
	// Запрос на получение значений атрибутов таблицы с указанным user_id
	query := "SELECT user_id, number_of_wins, number_of_defeats, number_of_draws FROM users WHERE user_id = $1"
	row := db.QueryRow(query, chatID)

	// Создание объекта User с полученными параметрами
	var u user.User
	err := row.Scan(&u.UserID, &u.NumberOfWins, &u.NumberOfDefeats, &u.NumberOfDraws)

	// Проверка на наличие пользователя в БД
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь с ID %d не найден", chatID)
		}
		// Возвращаем ошибку
		return nil, err
	}

	// Возвращаем объект User
	return &u, nil
}

func UpdateUserStats(chatID int64, u1 *user.User) error {
	// Запрос на обновление значений атрибутов таблицы с указанным user_id
	query := "UPDATE users SET number_of_wins = $1, number_of_defeats = $2, number_of_draws = $3 WHERE user_id = $4"
	_, err := db.Exec(query, u1.NumberOfWins, u1.NumberOfDefeats, u1.NumberOfDraws, chatID)
	return err
}
