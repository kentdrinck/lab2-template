package models

import (
	"time"
)

// Airport соответствует схеме таблицы airport
type Airport struct {
	ID      uint   `gorm:"primarykey"`
	Name    string `gorm:"size:255"`
	City    string `gorm:"size:255"`
	Country string `gorm:"size:255"`
}

// Flight соответствует схеме таблицы flight
type Flight struct {
	ID            uint      `gorm:"primarykey"`
	FlightNumber  string    `gorm:"size:20;not null"`
	Datetime      time.Time `gorm:"type:timestamp with time zone;not null"`
	FromAirportID uint
	ToAirportID   uint
	Price         int `gorm:"not null"`

	// Связи Gorm
	FromAirport Airport `gorm:"foreignKey:FromAirportID"`
	ToAirport   Airport `gorm:"foreignKey:ToAirportID"`
}

// FlightResponse - структура для ответа API, согласно вашему формату
type FlightResponse struct {
	FlightNumber string `json:"flightNumber"`
	FromAirport  string `json:"fromAirport"`
	ToAirport    string `json:"toAirport"`
	Date         string `json:"date"` // Будет форматирована как "2021-10-08 20:00"
	Price        int    `json:"price"`
}

// FlightNumbersRequest - DTO для тела POST-запроса
type FlightNumbersRequest struct {
	FlightNumbers []string `json:"flightNumbers" binding:"required,min=1"`
}

// FlightResponse и PaginationResponse используются из вашего примера.
// В данном контексте нам нужна только FlightResponse.
