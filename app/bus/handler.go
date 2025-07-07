package bus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/FreeJ1nG/bikuntracker-backend/utils/middleware"
)

type handler struct {
	repo      interfaces.BusRepository
	service   interfaces.BusService
	container interfaces.BusContainer
}

func NewHandler(repo interfaces.BusRepository, service interfaces.BusService, container interfaces.BusContainer) *handler {
	return &handler{
		repo:      repo,
		service:   service,
		container: container,
	}
}

func (h *handler) GetBuses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	res, err := h.repo.GetBuses(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse[[]models.Bus](w, res)
}

func (h *handler) CreateBus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := utils.ParseRequestBody[dto.CreateBusRequestBody](r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.repo.CreateBus(ctx, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse[models.Bus](w, *res)
}

func (h *handler) UpdateBus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := utils.ParseRequestBody[dto.UpdateBusRequestBody](r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, status, err := middleware.GetRouteParam(r, "id")
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	// First update the database
	res, err := h.repo.UpdateBus(ctx, &models.WhereData{FieldName: "id", Value: id}, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If color is being updated, also update the runtime bus coordinates
	if body.Color != nil {
		err = h.container.UpdateRuntimeBusColor(res.Imei, *body.Color)
		if err != nil {
			log.Printf("Failed to update runtime bus color for IMEI %s: %v", res.Imei, err)
			// Don't fail the request if runtime update fails, just log it
		}
	}

	utils.EncodeSuccessResponse[models.Bus](w, *res)
}

func (h *handler) DeleteBus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	id, status, err := middleware.GetRouteParam(r, "id")
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	err = h.repo.DeleteBus(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeEmptySuccessResponse(w)
}

// Lap history handlers
func (h *handler) GetLapHistory(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse filter from query parameters
	filter, err := h.parseLapHistoryFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If IMEI is provided in URL path, use it as priority
	if imei, _, pathErr := middleware.GetRouteParam(r, "imei"); pathErr == nil {
		filter.IMEI = &imei
	}

	log.Printf("GetLapHistory called with filter: IMEI=%v, RouteColor=%v", filter.IMEI, filter.RouteColor)

	// Set default pagination if not provided
	if filter.Limit == nil {
		defaultLimit := 20
		filter.Limit = &defaultLimit
	}
	if filter.Page == nil {
		defaultPage := 1
		filter.Page = &defaultPage
		offset := 0
		filter.Offset = &offset
	}

	// Get data and total count
	res, err := h.service.GetFilteredLapHistory(ctx, filter)
	if err != nil {
		log.Printf("GetFilteredLapHistory error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount, err := h.service.GetFilteredLapHistoryCount(ctx, filter)
	if err != nil {
		log.Printf("GetFilteredLapHistoryCount error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("GetFilteredLapHistory returned %d records, total: %d", len(res), totalCount)

	// Calculate pagination info
	totalPages := (totalCount + *filter.Limit - 1) / *filter.Limit // Ceiling division
	currentPage := *filter.Page
	hasNext := currentPage < totalPages

	response := dto.PaginatedResponse[models.BusLapHistory]{
		Success:     true,
		Data:        res,
		HasNext:     hasNext,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		TotalCount:  totalCount,
	}

	utils.EncodeSuccessResponse[dto.PaginatedResponse[models.BusLapHistory]](w, response)
}

func (h *handler) GetActiveLap(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	imei, status, err := middleware.GetRouteParam(r, "imei")
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	res, err := h.repo.GetActiveLapByImei(ctx, imei)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res == nil {
		http.Error(w, "No active lap found", http.StatusNotFound)
		return
	}

	utils.EncodeSuccessResponse[models.BusLapHistory](w, *res)
}

// GetFilteredLapHistory provides a dedicated endpoint for filtered lap history queries
func (h *handler) GetFilteredLapHistory(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse filter from query parameters
	filter, err := h.parseLapHistoryFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set default pagination if not provided
	if filter.Limit == nil {
		defaultLimit := 20
		filter.Limit = &defaultLimit
	}
	if filter.Page == nil {
		defaultPage := 1
		filter.Page = &defaultPage
		offset := 0
		filter.Offset = &offset
	}

	// Get data and total count
	res, err := h.service.GetFilteredLapHistory(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount, err := h.service.GetFilteredLapHistoryCount(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination info
	totalPages := (totalCount + *filter.Limit - 1) / *filter.Limit
	currentPage := *filter.Page
	hasNext := currentPage < totalPages

	response := dto.PaginatedResponse[models.BusLapHistory]{
		Success:     true,
		Data:        res,
		HasNext:     hasNext,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		TotalCount:  totalCount,
	}

	utils.EncodeSuccessResponse[dto.PaginatedResponse[models.BusLapHistory]](w, response)
}

// Debug endpoint to create test lap data
func (h *handler) CreateTestLapData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Create some test lap history data
	now := time.Now()
	testLap := &models.BusLapHistory{
		BusID:      1,
		IMEI:       "123456789",
		LapNumber:  1,
		StartTime:  now.Add(-2 * time.Hour),
		RouteColor: "blue",
	}

	res, err := h.repo.CreateLapHistory(ctx, testLap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// End the lap
	endTime := now.Add(-1 * time.Hour)
	res, err = h.repo.UpdateLapHistory(ctx, res.ID, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse[models.BusLapHistory](w, *res)
}

// Debug endpoint to check database table
func (h *handler) CheckLapHistoryTable(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Try to get the count of records
	count, err := h.repo.GetLapHistoryCount(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Table might not exist: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"table_exists": true,
		"record_count": count,
	}

	utils.EncodeSuccessResponse[map[string]interface{}](w, response)
}

// Helper method to parse lap history filter from query parameters
func (h *handler) parseLapHistoryFilter(r *http.Request) (dto.LapHistoryFilter, error) {
	var filter dto.LapHistoryFilter
	query := r.URL.Query()

	// Parse IMEI
	if imei := query.Get("imei"); imei != "" {
		filter.IMEI = &imei
	}

	// Parse Bus ID
	if busIDStr := query.Get("bus_id"); busIDStr != "" {
		busID, err := strconv.Atoi(busIDStr)
		if err != nil {
			return filter, err
		}
		filter.BusID = &busID
	}

	// Parse Route Color
	if routeColor := query.Get("route_color"); routeColor != "" {
		filter.RouteColor = &routeColor
	}

	// Parse From Date
	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			return filter, err
		}
		filter.FromDate = &fromDate
	}

	// Parse To Date
	if toDateStr := query.Get("to_date"); toDateStr != "" {
		toDate, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			return filter, err
		}
		// Set to end of day
		toDate = toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.ToDate = &toDate
	}

	// Parse Start Time
	if startTime := query.Get("start_time"); startTime != "" {
		// Validate format HH:MM
		if !isValidTimeFormat(startTime) {
			return filter, fmt.Errorf("invalid start_time format (expected HH:MM)")
		}
		filter.StartTime = &startTime
	}

	// Parse End Time
	if endTime := query.Get("end_time"); endTime != "" {
		// Validate format HH:MM
		if !isValidTimeFormat(endTime) {
			return filter, fmt.Errorf("invalid end_time format (expected HH:MM)")
		}
		filter.EndTime = &endTime
	}

	// Parse Limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return filter, err
		}
		filter.Limit = &limit
	}

	// Parse Offset
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return filter, fmt.Errorf("invalid offset: must be a non-negative integer")
		}
		filter.Offset = &offset
	}

	// Parse Page (1-based)
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return filter, fmt.Errorf("invalid page: must be a positive integer")
		}
		filter.Page = &page

		// Convert page to offset if limit is provided
		if filter.Limit != nil {
			offset := (page - 1) * (*filter.Limit)
			filter.Offset = &offset
		}
	}

	return filter, nil
}

// Helper function to validate time format HH:MM
func isValidTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])

	return err1 == nil && err2 == nil &&
		hour >= 0 && hour <= 23 &&
		minute >= 0 && minute <= 59
}
