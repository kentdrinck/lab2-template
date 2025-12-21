package handlers

import (
	. "gateway_service/models"
	. "gateway_service/service"
	"net/http"

	"github.com/gin-gonic/gin"

	"log"
)

// Server - Теперь содержит ссылку на Service
type Server struct {
	FlightService *FlightService // Новый слой Service
}

func NewServer(flightURL, ticketURL string) *Server {
	return &Server{
		FlightService: NewFlightService(flightURL, ticketURL),
	}
}

func ensureUserNameHeader(c *gin.Context) (string, bool) {
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Header 'X-User-Name' is required"})
		return "", false
	}
	return userName, true
}

// GetFlights обрабатывает GET /api/v1/flights (+)
func (s *Server) GetFlights(c *gin.Context) {
	var params GetFlightsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		// Устанавливаем разумные дефолты или возвращаем 400
		params.Page = 0
		params.Size = 10
	}
	if params.Size == 0 {
		params.Size = 100
	}
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Size > 100 {
		params.Size = 100
	}

	// Вызов слоя Service
	response, err := s.FlightService.GetFlights(params.Page, params.Size)

	if err != nil {
		log.Printf("Error fetching flights: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error while fetching flights"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTickets обрабатывает GET /api/v1/tickets (+)
func (s *Server) GetTickets(c *gin.Context) {
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	// Вызов слоя Service
	response, err := s.FlightService.GetUserTickets(userName)

	if err != nil {
		log.Printf("Error fetching user tickets for %s: %v", userName, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error while fetching tickets"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PurchaseTicket обрабатывает POST /api/v1/tickets
func (s *Server) PurchaseTicket(c *gin.Context) {
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	var purchaseReq TicketPurchaseRequest
	if err := c.ShouldBindJSON(&purchaseReq); err != nil {
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Message: "Ошибка валидации запроса: " + err.Error(),
			Errors:  []ErrorDescription{{Field: "request", Error: "invalid format"}},
		})
		return
	}

	// Вызов слоя Service
	response, err := s.FlightService.PurchaseTicket(userName, &purchaseReq)

	if err != nil {
		log.Printf("Error purchasing ticket for %s: %v", userName, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error during purchase"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTicketByUID обрабатывает GET /api/v1/tickets/{ticketUid} (+)
func (s *Server) GetTicketByUID(c *gin.Context) {
	ticketUID := c.Param("ticketUid")
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	// Вызов слоя Service
	response, err := s.FlightService.GetTicketByUID(userName, ticketUID)

	if err != nil {
		// if status == http.StatusNotFound {
		// 	c.JSON(http.StatusNotFound, ErrorResponse{Message: "Билет не найден"})
		// 	return
		// }
		log.Printf("Error fetching ticket %s: %v", ticketUID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefundTicket обрабатывает DELETE /api/v1/tickets/{ticketUid}
func (s *Server) RefundTicket(c *gin.Context) {
	ticketUID := c.Param("ticketUid")
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	// Вызов слоя Service
	status, err := s.FlightService.RefundTicket(userName, ticketUID)

	if err != nil {
		if status == http.StatusNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: "Билет не найден для возврата"})
			return
		}
		log.Printf("Error refunding ticket %s: %v", ticketUID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error"})
		return
	}

	c.Status(status) // 204 No Content
}

// GetUserInfo обрабатывает GET /api/v1/me
func (s *Server) GetUserInfo(c *gin.Context) {
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	// Вызов слоя Service
	response, err := s.FlightService.GetUserInfo(userName)

	if err != nil {
		log.Printf("Error fetching user info for %s: %v", userName, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error while fetching user info"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPrivilegeInfo обрабатывает GET /api/v1/privilege
func (s *Server) GetPrivilegeInfo(c *gin.Context) {
	userName, ok := ensureUserNameHeader(c)
	if !ok {
		return
	}

	// Вызов слоя Service
	response, err := s.FlightService.GetPrivilegeInfo(userName)

	if err != nil {
		log.Printf("Error fetching privilege info for %s: %v", userName, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal service error while fetching privilege info"})
		return
	}

	c.JSON(http.StatusOK, response)
}
