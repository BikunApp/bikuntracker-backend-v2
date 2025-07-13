package dto

import (
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type GetBusesResponse = []models.Bus

type CreateBusRequestBody struct {
	VehicleNo    string `json:"vehicle_no"`
	Imei         string `json:"imei"`
	IsActive     bool   `json:"is_active"`
	Color        string `json:"color"`
	CurrentHalte string `json:"current_halte,omitempty"`
	NextHalte    string `json:"next_halte,omitempty"`
}

type UpdateBusRequestBody struct {
	VehicleNo    *string `json:"vehicle_no,omitempty"`
	Imei         *string `json:"imei,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	Color        *string `json:"color,omitempty"`
	CurrentHalte *string `json:"current_halte,omitempty"`
	NextHalte    *string `json:"next_halte,omitempty"`
}

// Lap tracking DTOs
type LapEventData struct {
	EventType         string     `json:"event_type"` // "lap_start" or "lap_end"
	IMEI              string     `json:"imei"`
	LapID             int        `json:"lap_id"`
	LapNumber         int        `json:"lap_number"`
	RouteColor        string     `json:"route_color"`
	HalteVisitHistory string     `json:"halte_visit_history,omitempty"`
	StartTime         time.Time  `json:"start_time"`
	EndTime           *time.Time `json:"end_time,omitempty"`
	Duration          *float64   `json:"duration,omitempty"` // in seconds
	Timestamp         time.Time  `json:"timestamp"`
}

// Paginated response structure
type PaginatedResponse[T any] struct {
	Success     bool `json:"success"`
	Data        []T  `json:"data"`
	HasNext     bool `json:"hasNext"`
	TotalPages  int  `json:"totalPages"`
	CurrentPage int  `json:"currentPage"`
	TotalCount  int  `json:"totalCount"`
}

type GetLapHistoryResponse = []models.BusLapHistory

// Lap history filter DTO
type LapHistoryFilter struct {
	IMEI       *string    `json:"imei,omitempty"`        // Filter by specific bus IMEI
	BusID      *int       `json:"bus_id,omitempty"`      // Filter by bus ID
	RouteColor *string    `json:"route_color,omitempty"` // Filter by route color (blue, red, grey, etc.)
	FromDate   *time.Time `json:"from_date,omitempty"`   // Filter from start date
	ToDate     *time.Time `json:"to_date,omitempty"`     // Filter to end date
	StartTime  *string    `json:"start_time,omitempty"`  // Filter by time of day (HH:MM format)
	EndTime    *string    `json:"end_time,omitempty"`    // Filter by time of day (HH:MM format)
	Limit      *int       `json:"limit,omitempty"`       // Limit number of results
	Offset     *int       `json:"offset,omitempty"`      // Offset for pagination
	Page       *int       `json:"page,omitempty"`        // Page number (1-based)
}
