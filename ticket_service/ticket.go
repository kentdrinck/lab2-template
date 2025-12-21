package models

import (
	"github.com/google/uuid"
)

// StatusType определяет возможные статусы билета
type StatusType string

const (
	StatusPaid     StatusType = "PAID"
	StatusCanceled StatusType = "CANCELED"
)

// Ticket соответствует схеме таблицы ticket
type Ticket struct {
	ID           uint       `gorm:"primarykey"`
	TicketUID    uuid.UUID  `gorm:"type:uuid;unique;not null" json:"-"` // Скрыть в JSON, если не нужно
	Username     string     `gorm:"size:80;not null" json:"-"`
	FlightNumber string     `gorm:"size:20;not null" json:"flightNumber"`
	Price        int        `gorm:"not null" json:"price"`
	Status       StatusType `gorm:"type:varchar(20);not null" json:"status"`
}

// TicketResponse - структура для ответа API
type TicketResponse struct {
	TicketUID    uuid.UUID  `json:"ticketUid"`
	FlightNumber string     `json:"flightNumber"`
	Price        int        `json:"price"`
	Status       StatusType `json:"status"`
}
