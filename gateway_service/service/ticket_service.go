package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "gateway_service/models"
	"io"
	"net/http"
	"net/url"
)

// =========================================================================================
// 1. DTO Внешних Сервисов (Упрощённые или аналогичные внешние структуры)
//    Для простоты, мы используем те же DTO, что и для API, но на практике они могут отличаться.
// =========================================================================================

// Внешний сервис может возвращать что-то, что мы затем преобразуем
type ExternalFlightDTO struct {
	Number string `json:"flight_number"`
	Origin string `json:"from"`
	Dest   string `json:"to"`
	Time   string `json:"departure_time"`
	Cost   int    `json:"cost"`
}

type ExternalPaginationDTO struct {
	CurrentPage int                 `json:"current_page"`
	PageSize    int                 `json:"page_size"`
	Total       int                 `json:"total_elements"`
	Results     []ExternalFlightDTO `json:"results"`
}

// =========================================================================================
// 2. Структура Service и Клиент для HTTP
// =========================================================================================

// FlightService - содержит методы для взаимодействия с внешними системами
type FlightService struct {
	// Базовый клиент для всех запросов
	client *http.Client
	// Базовый URL для сервиса рейсов
	flightBaseURL string
	// Базовый URL для сервиса билетов/бонусов
	ticketBaseURL string
}

// NewFlightService - Конструктор сервиса
func NewFlightService(flightURL, ticketURL string) *FlightService {
	return &FlightService{
		client:        &http.Client{},
		flightBaseURL: flightURL,
		ticketBaseURL: ticketURL,
	}
}

// PerformRequest - Универсальный метод для выполнения HTTP-запросов
// Он берёт на себя логику отправки, проверки статуса (кроме 4xx/5xx) и парсинга ответа.
func (s *FlightService) PerformRequest(
	method, url string,
	body interface{}, // Тело запроса (например, TicketPurchaseRequest)
	authHeaderValue string, // Значение X-User-Name
	targetDTO interface{}, // DTO для парсинга успешного ответа
) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("ошибка сериализации тела запроса: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeaderValue != "" {
		req.Header.Set("X-User-Name", authHeaderValue)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки HTTP-запроса: %w", err)
	}

	buf := bytes.Buffer{}
	r := io.NopCloser(io.TeeReader(resp.Body, &buf))

	// Если статус 2xx и DTO для парсинга предоставлен, парсим его
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && targetDTO != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(r).Decode(targetDTO); err != nil {
			return resp, fmt.Errorf("ошибка парсинга DTO ответа: %w", err)
		}
	} else if resp.StatusCode >= 400 {
		// Оставляем resp открытым, чтобы вызывающий код мог прочитать тело ошибки (4xx/5xx)
		return resp, fmt.Errorf("внешний сервис вернул статус %d", resp.StatusCode)
	}
	data := buf.Bytes()
	fmt.Println("FUCK", string(data))
	fmt.Println("FUCK", targetDTO)

	return resp, nil
}

// =========================================================================================
// 3. Методы Сервиса (реализация ручек)
// =========================================================================================

// GetFlights fetches flights from external service and maps to PaginationResponse DTO
func (s *FlightService) GetFlights(page, size int) (*PaginationResponse, error) {
	url := fmt.Sprintf("%s/flights?page=%d&size=%d", s.flightBaseURL, page, size)

	// DTO для парсинга ответа внешнего сервиса
	var externalResponse ExternalPaginationDTO

	// Выполняем GET запрос. X-User-Name не требуется.
	resp, err := s.PerformRequest(http.MethodGet, url, nil, "", &externalResponse)
	if err != nil {
		// Если статус 4xx/5xx, возвращаем ошибку для обработки контроллером
		if resp != nil {
			return nil, fmt.Errorf("ошибка внешнего сервиса (статус %d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	// --- Преобразование ExternalDTO в OpenAPI DTO ---
	apiFlights := make([]FlightResponse, len(externalResponse.Results))
	for i, f := range externalResponse.Results {
		apiFlights[i] = FlightResponse{
			FlightNumber: f.Number,
			FromAirport:  f.Origin,
			ToAirport:    f.Dest,
			Date:         f.Time,
			Price:        f.Cost,
		}
	}

	// Возвращаем DTO для контроллера
	return &PaginationResponse{
		Page:          externalResponse.CurrentPage,
		PageSize:      externalResponse.PageSize,
		TotalElements: externalResponse.Total,
		Items:         apiFlights,
	}, nil
}

// PurchaseTicket sends a request to the ticket service
func (s *FlightService) PurchaseTicket(username string, req *TicketPurchaseRequest) (*TicketPurchaseResponse, error) {
	url := fmt.Sprintf("%s/v1/tickets", s.ticketBaseURL)

	// DTO для парсинга ответа (в данном случае, оно совпадает с исходящим DTO)
	var apiResponse TicketPurchaseResponse

	// Выполняем POST запрос.
	resp, err := s.PerformRequest(http.MethodPost, url, req, username, &apiResponse)

	if err != nil {
		// Если это ошибка 4xx/5xx, мы можем прочитать тело ошибки для более детального ответа
		if resp != nil && (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound) {

			// Пытаемся распарсить тело как ValidationErrorResponse или ErrorResponse
			errorBody, _ := io.ReadAll(resp.Body)

			// Для Bad Request
			if resp.StatusCode == http.StatusBadRequest {
				var validationErr ValidationErrorResponse
				if json.Unmarshal(errorBody, &validationErr) == nil {
					return nil, fmt.Errorf("validation error (%d): %s", resp.StatusCode, validationErr.Message)
				}
			}

			// Для других 4xx/5xx (общее)
			var generalErr ErrorResponse
			if json.Unmarshal(errorBody, &generalErr) == nil {
				return nil, fmt.Errorf("service error (%d): %s", resp.StatusCode, generalErr.Message)
			}
		}

		return nil, fmt.Errorf("ошибка при покупке билета: %w", err)
	}

	// Возвращаем DTO для контроллера
	return &apiResponse, nil
}

// GetUserTickets fetches all tickets for a user
func (s *FlightService) GetUserTickets(username string) ([]TicketResponse, error) {
	values := url.Values{}
	values.Add("user", username)

	fullUrl := fmt.Sprintf("%s/tickets?%s", s.ticketBaseURL, values.Encode())

	var apiResponse []TicketResponse

	resp, err := s.PerformRequest(http.MethodGet, fullUrl, nil, username, &apiResponse)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("ошибка получения билетов (%d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	values = url.Values{}
	for _, v := range apiResponse {
		values.Add("flightNumber", v.FlightNumber)
	}

	var apiResponse2 []FlightResponse

	fullUrl = fmt.Sprintf("%s/flights-by-number?%s", s.flightBaseURL, values.Encode())
	resp, err = s.PerformRequest(http.MethodGet, fullUrl, nil, "", &apiResponse2)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("ошибка получения билетов (%d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	numberToStruct := make(map[string]FlightResponse)
	for _, v := range apiResponse2 {
		numberToStruct[v.FlightNumber] = v
	}

	for i, v := range apiResponse {
		apiResponse[i].FromAirport = numberToStruct[v.FlightNumber].FromAirport
		apiResponse[i].ToAirport = numberToStruct[v.FlightNumber].ToAirport
		apiResponse[i].Date = numberToStruct[v.FlightNumber].Date
	}

	return apiResponse, nil
}

// GetTicketByUID fetches a specific ticket
func (s *FlightService) GetTicketByUID(username, ticketUID string) (*TicketResponse, error) {
	fullUrl := fmt.Sprintf("%s/tickets/%s", s.ticketBaseURL, ticketUID)

	var apiResponse TicketResponse

	resp, err := s.PerformRequest(http.MethodGet, fullUrl, nil, username, &apiResponse)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("билет не найден")
		}
		return nil, fmt.Errorf("ошибка при получении билета: %w", err)
	}

	values := url.Values{}
	values.Add("flightNumber", apiResponse.FlightNumber)

	var apiResponse2 []FlightResponse

	fullUrl = fmt.Sprintf("%s/flights-by-number?%s", s.flightBaseURL, values.Encode())
	resp, err = s.PerformRequest(http.MethodGet, fullUrl, nil, "", &apiResponse2)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("ошибка получения билетов (%d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	apiResponse.Date = apiResponse2[0].Date
	apiResponse.FromAirport = apiResponse2[0].FromAirport
	apiResponse.ToAirport = apiResponse2[0].ToAirport

	return &apiResponse, nil
}

// RefundTicket refunds a specific ticket
func (s *FlightService) RefundTicket(username, ticketUID string) (int, error) {
	url := fmt.Sprintf("%s/v1/tickets/%s", s.ticketBaseURL, ticketUID)

	// DELETE запрос не возвращает тело в 204 No Content, targetDTO = nil
	resp, err := s.PerformRequest(http.MethodDelete, url, nil, username, nil)

	if err != nil {
		// Если 404, то возвращаем код 404
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return http.StatusNotFound, fmt.Errorf("билет не найден для возврата")
		}
		// Для всех остальных ошибок
		return http.StatusInternalServerError, fmt.Errorf("ошибка при возврате билета: %w", err)
	}

	// Успешно, возвращаем код 204 No Content
	return http.StatusNoContent, nil
}

// GetUserInfo fetches user and privilege info
func (s *FlightService) GetUserInfo(username string) (*UserInfoResponse, error) {
	url := fmt.Sprintf("%s/v1/me", s.ticketBaseURL)

	var apiResponse UserInfoResponse

	resp, err := s.PerformRequest(http.MethodGet, url, nil, username, &apiResponse)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("ошибка получения информации о пользователе (%d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	return &apiResponse, nil
}

// GetPrivilegeInfo fetches user's bonus account status
func (s *FlightService) GetPrivilegeInfo(username string) (*PrivilegeInfoResponse, error) {
	url := fmt.Sprintf("%s/v1/privilege", s.ticketBaseURL)

	var apiResponse PrivilegeInfoResponse

	resp, err := s.PerformRequest(http.MethodGet, url, nil, username, &apiResponse)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("ошибка получения бонусного счета (%d): %w", resp.StatusCode, err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	return &apiResponse, nil
}
