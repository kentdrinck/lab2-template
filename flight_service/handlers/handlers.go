package handlers

import (
	"net/http"
	"strconv"

	"flight_service/database"
	"flight_service/models"

	"github.com/gin-gonic/gin"
)

// PaginationResponse - структура для ответа с пагинацией
type PaginationResponse struct {
	Page          int                     `json:"page"`
	PageSize      int                     `json:"pageSize"`
	TotalElements int64                   `json:"totalElements"`
	Items         []models.FlightResponse `json:"items"`
}

// GetFlights обрабатывает GET запрос для получения списка перелетов с пагинацией
func GetFlights(c *gin.Context) {
	// 1. Парсинг параметров пагинации
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Ограничиваем максимальный размер страницы, чтобы избежать перегрузки
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// 2. Получение общего количества элементов
	var totalElements int64
	// Используем *database.DB для доступа к Gorm DB
	if err := database.DB.Model(&models.Flight{}).Count(&totalElements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения общего числа перелетов"})
		return
	}

	// 3. Получение данных с пагинацией и предзагрузкой связанных таблиц
	var flights []models.Flight
	err = database.DB.
		Limit(pageSize).
		Offset(offset).
		// Предзагрузка (Preload) FromAirport и ToAirport для выполнения JOIN-ов
		Preload("FromAirport").
		Preload("ToAirport").
		Find(&flights).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка перелетов"})
		return
	}

	// 4. Преобразование в формат ответа API
	flightResponses := make([]models.FlightResponse, len(flights))
	for i, flight := range flights {
		flightResponses[i] = models.FlightResponse{
			FlightNumber: flight.FlightNumber,
			FromAirport:  flight.FromAirport.Name + " " + flight.FromAirport.City, // Используем Название и Город
			ToAirport:    flight.ToAirport.Name + " " + flight.ToAirport.City,
			// Форматируем время в "2021-10-08 20:00"
			Date:  flight.Datetime.Format("2006-01-02 15:04"),
			Price: flight.Price,
		}
	}

	// 5. Отправка ответа
	response := PaginationResponse{
		Page:          page,
		PageSize:      pageSize,
		TotalElements: totalElements,
		Items:         flightResponses,
	}

	c.JSON(http.StatusOK, response)
}
