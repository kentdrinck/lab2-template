package main

import (
	"log"
	"os"

	"ticket-service/database"
	"ticket-service/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// ... ПРИМЕЧАНИЕ: Добавьте в database/db.go models.Ticket в AutoMigrate
/* func InitDB() {
	// ... подключение
	DB.AutoMigrate(&models.Ticket{}) // Новая миграция
	// ...
}
*/

func main() {
	// 1. Загрузка переменных окружения из .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить файл .env. Использование системных переменных окружения.")
	}

	// 2. Инициализация базы данных
	database.InitDB()

	router := gin.Default()

	// Группа роутов API версии 1

	router.GET("/tickets", handlers.GetTicketsForUser)
	router.GET("/tickets/:ticketUid", handlers.GetTicketByUid)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервис билетов запущен на порту :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
