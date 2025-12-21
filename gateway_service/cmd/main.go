package main

import (
	"gateway_service/handlers"
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

	// Инициализация роутера Gin
	router := gin.Default()

	server := handlers.NewServer("http://localhost:8060", "http://localhost:8070")
	v1 := router.Group("/api/v1")
	v1.GET("/flights", server.GetFlights)
	v1.GET("/tickets", server.GetTickets)
	v1.POST("/tickets", server.PurchaseTicket)
	v1.GET("/tickets/:ticketUid", server.GetTicketByUID)
	v1.DELETE("/tickets/:ticketUid", server.RefundTicket)
	v1.GET("/me", server.GetUserInfo)
	v1.GET("/privilege", server.GetPrivilegeInfo)

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
