package handlers

import (
	"net/http"

	"ticket-service/database"
	"ticket-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// extractUsername проверяет наличие заголовка X-User-Name
// func extractUsername(c *gin.Context) (string, bool) {
// 	username := c.GetHeader("X-User-Name")
// 	if username == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"message": "Требуется заголовок X-User-Name."})
// 		return "", false
// 	}
// 	return username, true
// }

func extractUsername(c *gin.Context) (string, bool) {
	username, _ := c.GetQuery("user")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Требуется указать user."})
		return "", false
	}
	return username, true
}

// convertToResponse преобразует модель Ticket в структуру ответа
func convertToResponse(ticket models.Ticket) models.TicketResponse {
	return models.TicketResponse{
		TicketUID:    ticket.TicketUID,
		FlightNumber: ticket.FlightNumber,
		Price:        ticket.Price,
		Status:       ticket.Status,
	}
}

// GetTicketsForUser обрабатывает GET {{baseUrl}}/api/v1/tickets
// Возвращает все билеты, принадлежащие пользователю из заголовка.
func GetTicketsForUser(c *gin.Context) {
	username, ok := extractUsername(c)
	if !ok {
		return
	}

	var tickets []models.Ticket
	// Выбираем все билеты, где username совпадает с заголовком
	result := database.DB.Where("username = ?", username).Find(&tickets)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка получения билетов"})
		return
	}

	// Преобразование в формат ответа API
	responses := make([]models.TicketResponse, len(tickets))
	for i, t := range tickets {
		responses[i] = convertToResponse(t)
	}

	c.JSON(http.StatusOK, responses)
}

// GetTicketByUid обрабатывает GET {{baseUrl}}/api/v1/tickets/{{ticketUid}}
// Возвращает конкретный билет, проверяя его принадлежность пользователю.
func GetTicketByUid(c *gin.Context) {

	ticketUidParam := c.Param("ticketUid")
	ticketUID, err := uuid.Parse(ticketUidParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Неверный формат ticketUid"})
		return
	}

	var ticket models.Ticket

	// Ищем по TicketUID и обязательно по Username
	result := database.DB.
		Where("ticket_uid = ?", ticketUID).
		First(&ticket)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Возвращаем 404, если билет не найден или не принадлежит пользователю
			c.JSON(http.StatusNotFound, gin.H{"message": "Билет не найден или не принадлежит пользователю"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка поиска билета"})
		return
	}

	c.JSON(http.StatusOK, convertToResponse(ticket))
}
