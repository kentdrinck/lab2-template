package database

import (
	"fmt"
	"log"
	"os"

	"ticket-service/models" // Замените на путь к вашим моделям

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	// Настройка подключения к PostgreSQL.
	// В реальном приложении эти данные должны быть в конфигурации (например, через viper или env-переменные).
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Автоматическая миграция (создание/обновление таблиц)
	// В продакшене лучше использовать полноценные миграции
	err = DB.AutoMigrate(&models.Ticket{})
	if err != nil {
		log.Fatalf("Ошибка автомиграции: %v", err)
	}

	log.Println("База данных успешно подключена и мигрирована!")
}
