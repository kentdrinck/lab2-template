package models

// =========================================================================================
// 1. DTO (Data Transfer Objects) на основе OpenAPI Схем
// =========================================================================================

// Enums
const (
	TicketStatusPaid      = "PAID"
	TicketStatusCanceled  = "CANCELED"
	PrivilegeStatusBronze = "BRONZE"
	PrivilegeStatusSilver = "SILVER"
	PrivilegeStatusGold   = "GOLD"
	OperationTypeFillIn   = "FILL_IN_BALANCE"
	OperationTypeDebit    = "DEBIT_THE_ACCOUNT"
	OperationTypeMoney    = "FILLED_BY_MONEY"
)

// FlightResponse (GET /api/v1/flights/items)
type FlightResponse struct {
	FlightNumber string `json:"flightNumber"`
	FromAirport  string `json:"fromAirport"`
	ToAirport    string `json:"toAirport"`
	Date         string `json:"date"`  // ISO 8601
	Price        int    `json:"price"` // Используем int, так как в примере целое число
}

// PaginationResponse (GET /api/v1/flights)
type PaginationResponse struct {
	Page          int              `json:"page"`
	PageSize      int              `json:"pageSize"`
	TotalElements int              `json:"totalElements"`
	Items         []FlightResponse `json:"items"`
}

// TicketResponse (GET /api/v1/tickets, GET /api/v1/tickets/{ticketUid})
type TicketResponse struct {
	TicketUID    string `json:"ticketUid"` // uuid
	FlightNumber string `json:"flightNumber"`
	FromAirport  string `json:"fromAirport"`
	ToAirport    string `json:"toAirport"`
	Date         string `json:"date"` // ISO 8601
	Price        int    `json:"price"`
	Status       string `json:"status"` // PAID, CANCELED
}

// PrivilegeShortInfo (nested)
type PrivilegeShortInfo struct {
	Balance string `json:"balance"` // В схеме string
	Status  string `json:"status"`  // BRONZE, SILVER, GOLD
}

// UserInfoResponse (GET /api/v1/me)
type UserInfoResponse struct {
	Tickets   []TicketResponse   `json:"tickets"`
	Privilege PrivilegeShortInfo `json:"privilege"`
}

// BalanceHistory (nested in PrivilegeInfoResponse)
type BalanceHistory struct {
	Date          string `json:"date"`          // ISO 8601
	BalanceDiff   string `json:"balanceDiff"`   // В схеме string
	TicketUID     string `json:"ticketUid"`     // uuid
	OperationType string `json:"operationType"` // FILL_IN_BALANCE, DEBIT_THE_ACCOUNT, FILLED_BY_MONEY
}

// PrivilegeInfoResponse (GET /api/v1/privilege)
type PrivilegeInfoResponse struct {
	Balance string           `json:"balance"` // В схеме string
	Status  string           `json:"status"`  // BRONZE, SILVER, GOLD
	History []BalanceHistory `json:"history"`
}

// TicketPurchaseRequest (POST /api/v1/tickets request body)
type TicketPurchaseRequest struct {
	FlightNumber    string `json:"flightNumber" binding:"required"`
	Price           int    `json:"price" binding:"required,min=1"`
	PaidFromBalance bool   `json:"paidFromBalance"`
}

// TicketPurchaseResponse (POST /api/v1/tickets response)
type TicketPurchaseResponse struct {
	TicketUID     string             `json:"ticketUid"`
	FlightNumber  string             `json:"flightNumber"`
	FromAirport   string             `json:"fromAirport"`
	ToAirport     string             `json:"toAirport"`
	Date          string             `json:"date"`   // ISO 8601
	Status        string             `json:"status"` // PAID, CANCELED
	Price         int                `json:"price"`
	PaidByMoney   int                `json:"paidByMoney"`
	PaidByBonuses int                `json:"paidByBonuses"`
	Privilege     PrivilegeShortInfo `json:"privilege"`
}

// ErrorDescription (nested in ValidationErrorResponse)
type ErrorDescription struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse (404 responses)
type ErrorResponse struct {
	Message string `json:"message"`
}

// ValidationErrorResponse (400 response)
type ValidationErrorResponse struct {
	Message string             `json:"message"`
	Errors  []ErrorDescription `json:"errors"`
}

// Параметры запроса для GET /api/v1/flights
type GetFlightsQueryParams struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
