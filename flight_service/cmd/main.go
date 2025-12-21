package main

import (
	"flight_service/database"
	"flight_service/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Импортируем godotenv
)

func main() {
	err := godotenv.Load()
	if err != nil {
		// Ошибка может быть, если .env не существует.
		// Можно изменить на log.Println, если вы хотите разрешить запуск без .env
		// (при условии, что переменные заданы в реальном окружении).
		log.Println("Не удалось загрузить файл .env. Использование системных переменных окружения.")
	}
	// Настройка Gin: устанавливаем режим Release для продакшена
	// gin.SetMode(gin.ReleaseMode)

	// Инициализация базы данных
	// Перед запуском убедитесь, что установлены переменные окружения DB_HOST, DB_USER и т.д.
	// или замените os.Getenv("...") на ваши настройки в database/db.go
	database.InitDB()

	// Инициализация роутера Gin
	router := gin.Default()

	router.GET("/flights", handlers.GetFlights)
	router.GET("/flights-by-number/", handlers.GetFlightsByNumbers)

	// Запуск сервера. Порт берется из переменной окружения или по умолчанию 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на порту :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
