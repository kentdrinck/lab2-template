package handlers

import (
	"net/http"
	"strconv"

	"flight_service/database"
	"flight_service/models"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// PaginationResponse - структура для ответа с пагинацией
type PaginationResponse struct {
	Page          int                     `json:"current_page"`
	PageSize      int                     `json:"page_size"`
	TotalElements int64                   `json:"total_elements"`
	Items         []models.FlightResponse `json:"results"`
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

// GetFlightsByNumbers обрабатывает POST запрос, принимая список номеров рейсов
// и возвращая соответствующие данные FlightResponse.
func GetFlightsByNumbers(c *gin.Context) {
	// 1. Получение списка номеров рейсов из Query-параметра
	flightNumbers := c.QueryArray("flightNumber")

	// 2. Получение данных из базы данных
	var flights []models.Flight

	// Используем WHERE IN для выбора рейсов по списку номеров.
	// Вызов DB.Where() с массивом строк автоматически создаст SQL "IN (?)".
	err := database.DB.
		Where("flight_number IN (?)", flightNumbers).
		Preload("FromAirport").
		Preload("ToAirport").
		Find(&flights).Error

	// Обработка ошибки базы данных
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения перелетов по номерам"})
		return
	}

	// 3. Преобразование в формат ответа API
	flightResponses := make([]models.FlightResponse, len(flights))
	for i, flight := range flights {
		// Проверка на nil, если связь не была найдена (хотя Preload должен сработать)
		fromAirportName := ""
		fromAirportCity := ""
		fromAirportName = flight.FromAirport.Name
		fromAirportCity = flight.FromAirport.City

		toAirportName := ""
		toAirportCity := ""
		toAirportName = flight.ToAirport.Name
		toAirportCity = flight.ToAirport.City

		flightResponses[i] = models.FlightResponse{
			FlightNumber: flight.FlightNumber,
			// Объединяем Название и Город, как в вашем примере
			FromAirport: fromAirportName + " " + fromAirportCity,
			ToAirport:   toAirportName + " " + toAirportCity,
			// Форматируем время
			Date:  flight.Datetime.Format("2006-01-02 15:04"),
			Price: flight.Price,
		}
	}

	c.JSON(http.StatusOK, flightResponses)
}
