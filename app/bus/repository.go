package bus

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetBuses(ctx context.Context) (res []models.Bus, err error) {
	rows, err := r.db.Query(ctx, `SELECT id, vehicle_no, imei, is_active, color, bus_number, plate_number, current_halte, next_halte, created_at, updated_at FROM bus;`)
	if err != nil {
		err = fmt.Errorf("unable to execute SQL query to get buses: %w", err)
		return
	}
	res = make([]models.Bus, 0)
	for rows.Next() {
		var bus models.Bus
		var currentHalte, nextHalte, busNumber, plateNumber sql.NullString
		if err := rows.Scan(
			&bus.Id,
			&bus.VehicleNo,
			&bus.Imei,
			&bus.IsActive,
			&bus.Color,
			&busNumber,
			&plateNumber,
			&currentHalte,
			&nextHalte,
			&bus.CreatedAt,
			&bus.UpdatedAt,
		); err != nil {
			err = fmt.Errorf("unable to scan SQL result: %w", err)
			return res, err
		}
		if busNumber.Valid {
			bus.BusNumber = busNumber.String
		}
		if plateNumber.Valid {
			bus.PlateNumber = plateNumber.String
		}
		if currentHalte.Valid {
			bus.CurrentHalte = currentHalte.String
		}
		if nextHalte.Valid {
			bus.NextHalte = nextHalte.String
		}
		res = append(res, bus)
	}

	return
}

func (r *repository) CreateBus(ctx context.Context, data dto.CreateBusRequestBody) (res *models.Bus, err error) {
	row := r.db.QueryRow(
		ctx,
		`INSERT INTO bus (imei, vehicle_no, is_active, color, bus_number, plate_number) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, vehicle_no, imei, is_active, color, bus_number, plate_number, current_halte, next_halte, created_at, updated_at;`,
		data.Imei,
		data.VehicleNo,
		data.IsActive,
		data.Color,
		data.BusNumber,
		data.PlateNumber,
	)

	var createdBus models.Bus
	var currentHalte, nextHalte, busNumber, plateNumber sql.NullString
	if err = row.Scan(
		&createdBus.Id,
		&createdBus.VehicleNo,
		&createdBus.Imei,
		&createdBus.IsActive,
		&createdBus.Color,
		&busNumber,
		&plateNumber,
		&currentHalte,
		&nextHalte,
		&createdBus.CreatedAt,
		&createdBus.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("unable to execute create bus SQL: %w", err)
		return
	}
	if busNumber.Valid {
		createdBus.BusNumber = busNumber.String
	}
	if plateNumber.Valid {
		createdBus.PlateNumber = plateNumber.String
	}
	if currentHalte.Valid {
		createdBus.CurrentHalte = currentHalte.String
	}
	if nextHalte.Valid {
		createdBus.NextHalte = nextHalte.String
	}
	res = &createdBus
	return
}

func (r *repository) UpdateBus(ctx context.Context, whereData *models.WhereData, data dto.UpdateBusRequestBody) (res *models.Bus, err error) {
	sqlStr, params, err := utils.GetPartialUpdateSQL("bus", data, whereData)
	if err != nil {
		return
	}
	row := r.db.QueryRow(
		ctx,
		sqlStr+" RETURNING id, vehicle_no, imei, is_active, color, bus_number, plate_number, current_halte, next_halte, created_at, updated_at",
		params...,
	)
	var updatedBus models.Bus
	var currentHalte, nextHalte, busNumber, plateNumber sql.NullString
	if err = row.Scan(
		&updatedBus.Id,
		&updatedBus.VehicleNo,
		&updatedBus.Imei,
		&updatedBus.IsActive,
		&updatedBus.Color,
		&busNumber,
		&plateNumber,
		&currentHalte,
		&nextHalte,
		&updatedBus.CreatedAt,
		&updatedBus.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			err = fmt.Errorf("no bus found with %s = %v", whereData.FieldName, whereData.Value)
			return
		}
		err = fmt.Errorf("unable to execute update bus SQL: %w", err)
		return
	}
	if busNumber.Valid {
		updatedBus.BusNumber = busNumber.String
	}
	if plateNumber.Valid {
		updatedBus.PlateNumber = plateNumber.String
	}
	if currentHalte.Valid {
		updatedBus.CurrentHalte = currentHalte.String
	}
	if nextHalte.Valid {
		updatedBus.NextHalte = nextHalte.String
	}
	res = &updatedBus
	return
}

func (r *repository) DeleteBus(ctx context.Context, id string) (err error) {
	_, err = r.db.Exec(ctx, "DELETE FROM bus WHERE id = $1", id)
	return
}

func (r *repository) InsertBuses(ctx context.Context, data []models.Bus) (err error) {
	batch := &pgx.Batch{}
	for _, bus := range data {
		batch.Queue("INSERT INTO bus (vehicle_no, imei, is_active, color, bus_number, plate_number) VALUES ($1, $2, $3, $4, $5, $6)", bus.VehicleNo, bus.Imei, bus.IsActive, bus.Color, bus.BusNumber, bus.PlateNumber)
	}
	err = r.db.SendBatch(ctx, batch).Close()
	if err != nil {
		err = fmt.Errorf("unable to batch insert buses: %w", err)
		return
	}
	return
}

// Lap history repository methods
func (r *repository) CreateLapHistory(ctx context.Context, lapHistory *models.BusLapHistory) (*models.BusLapHistory, error) {
	row := r.db.QueryRow(
		ctx,
		`INSERT INTO bus_lap_history (bus_id, imei, lap_number, start_time, route_color, halte_visit_history) 
		 VALUES ($1, $2, $3, $4, $5, $6) 
		 RETURNING id, bus_id, imei, lap_number, start_time, end_time, route_color, halte_visit_history, created_at, updated_at`,
		lapHistory.BusID,
		lapHistory.IMEI,
		lapHistory.LapNumber,
		lapHistory.StartTime,
		lapHistory.RouteColor,
		lapHistory.HalteVisitHistory,
	)

	var created models.BusLapHistory
	var endTime sql.NullTime
	var halteVisitHistory sql.NullString
	err := row.Scan(
		&created.ID,
		&created.BusID,
		&created.IMEI,
		&created.LapNumber,
		&created.StartTime,
		&endTime,
		&created.RouteColor,
		&halteVisitHistory,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create lap history: %w", err)
	}

	if endTime.Valid {
		created.EndTime = &endTime.Time
	}

	if halteVisitHistory.Valid {
		created.HalteVisitHistory = halteVisitHistory.String
	}

	return &created, nil
}

func (r *repository) UpdateLapHistory(ctx context.Context, id int, endTime interface{}) (*models.BusLapHistory, error) {
	row := r.db.QueryRow(
		ctx,
		`UPDATE bus_lap_history SET end_time = $1, updated_at = now() 
		 WHERE id = $2 
		 RETURNING id, bus_id, imei, lap_number, start_time, end_time, route_color, halte_visit_history, created_at, updated_at`,
		endTime,
		id,
	)

	var updated models.BusLapHistory
	var endTimeNull sql.NullTime
	var halteVisitHistory sql.NullString
	err := row.Scan(
		&updated.ID,
		&updated.BusID,
		&updated.IMEI,
		&updated.LapNumber,
		&updated.StartTime,
		&endTimeNull,
		&updated.RouteColor,
		&halteVisitHistory,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to update lap history: %w", err)
	}

	if endTimeNull.Valid {
		updated.EndTime = &endTimeNull.Time
	}

	if halteVisitHistory.Valid {
		updated.HalteVisitHistory = halteVisitHistory.String
	}

	return &updated, nil
}

func (r *repository) UpdateLapHistoryWithColor(ctx context.Context, id int, endTime interface{}, routeColor string) (*models.BusLapHistory, error) {
	row := r.db.QueryRow(
		ctx,
		`UPDATE bus_lap_history SET end_time = $1, route_color = $2, updated_at = now() 
		 WHERE id = $3 
		 RETURNING id, bus_id, imei, lap_number, start_time, end_time, route_color, halte_visit_history, created_at, updated_at`,
		endTime,
		routeColor,
		id,
	)

	var updated models.BusLapHistory
	var endTimeNull sql.NullTime
	var halteVisitHistory sql.NullString
	err := row.Scan(
		&updated.ID,
		&updated.BusID,
		&updated.IMEI,
		&updated.LapNumber,
		&updated.StartTime,
		&endTimeNull,
		&updated.RouteColor,
		&halteVisitHistory,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to update lap history with color: %w", err)
	}

	if endTimeNull.Valid {
		updated.EndTime = &endTimeNull.Time
	}

	if halteVisitHistory.Valid {
		updated.HalteVisitHistory = halteVisitHistory.String
	}

	return &updated, nil
}

func (r *repository) UpdateLapHistoryHalteVisits(ctx context.Context, id int, halteVisitHistory string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE bus_lap_history SET halte_visit_history = $1, updated_at = now() 
		 WHERE id = $2`,
		halteVisitHistory,
		id,
	)
	if err != nil {
		return fmt.Errorf("unable to update lap history halte visits: %w", err)
	}
	return nil
}

func (r *repository) GetActiveLapByImei(ctx context.Context, imei string) (*models.BusLapHistory, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT blh.id, blh.bus_id, blh.imei, blh.lap_number, blh.start_time, blh.end_time, blh.route_color, blh.halte_visit_history, blh.created_at, blh.updated_at,
		        b.vehicle_no, b.bus_number, b.plate_number, b.is_active, b.color
		 FROM bus_lap_history blh 
		 JOIN bus b ON blh.bus_id = b.id
		 WHERE blh.imei = $1 AND blh.end_time IS NULL 
		 ORDER BY blh.start_time DESC 
		 LIMIT 1`,
		imei,
	)

	var lap models.BusLapHistory
	var endTime sql.NullTime
	var halteVisitHistory, busNumber, plateNumber sql.NullString
	err := row.Scan(
		&lap.ID,
		&lap.BusID,
		&lap.IMEI,
		&lap.LapNumber,
		&lap.StartTime,
		&endTime,
		&lap.RouteColor,
		&halteVisitHistory,
		&lap.CreatedAt,
		&lap.UpdatedAt,
		&lap.VehicleNo,
		&busNumber,
		&plateNumber,
		&lap.IsActive,
		&lap.Color,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No active lap found
		}
		return nil, fmt.Errorf("unable to get active lap: %w", err)
	}

	if endTime.Valid {
		lap.EndTime = &endTime.Time
	}

	if halteVisitHistory.Valid {
		lap.HalteVisitHistory = halteVisitHistory.String
	}

	if busNumber.Valid {
		lap.BusNumber = busNumber.String
	}

	if plateNumber.Valid {
		lap.PlateNumber = plateNumber.String
	}

	return &lap, nil
}

func (r *repository) GetLapHistoryByImei(ctx context.Context, imei string) ([]models.BusLapHistory, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT blh.id, blh.bus_id, blh.imei, blh.lap_number, blh.start_time, blh.end_time, blh.route_color, blh.halte_visit_history, blh.created_at, blh.updated_at,
		        b.vehicle_no, b.bus_number, b.plate_number, b.is_active, b.color
		 FROM bus_lap_history blh 
		 JOIN bus b ON blh.bus_id = b.id
		 WHERE blh.imei = $1 
		 ORDER BY blh.start_time DESC`,
		imei,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get lap history: %w", err)
	}
	defer rows.Close()

	var laps []models.BusLapHistory
	// Initialize empty slice to avoid null response
	laps = make([]models.BusLapHistory, 0)

	for rows.Next() {
		var lap models.BusLapHistory
		var endTime sql.NullTime
		var halteVisitHistory, busNumber, plateNumber sql.NullString
		err := rows.Scan(
			&lap.ID,
			&lap.BusID,
			&lap.IMEI,
			&lap.LapNumber,
			&lap.StartTime,
			&endTime,
			&lap.RouteColor,
			&halteVisitHistory,
			&lap.CreatedAt,
			&lap.UpdatedAt,
			&lap.VehicleNo,
			&busNumber,
			&plateNumber,
			&lap.IsActive,
			&lap.Color,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan lap history: %w", err)
		}

		if endTime.Valid {
			lap.EndTime = &endTime.Time
		}

		if halteVisitHistory.Valid {
			lap.HalteVisitHistory = halteVisitHistory.String
		}

		if busNumber.Valid {
			lap.BusNumber = busNumber.String
		}

		if plateNumber.Valid {
			lap.PlateNumber = plateNumber.String
		}

		laps = append(laps, lap)
	}

	return laps, nil
}

func (r *repository) GetFilteredLapHistory(ctx context.Context, filter dto.LapHistoryFilter) ([]models.BusLapHistory, error) {
	query := `SELECT blh.id, blh.bus_id, blh.imei, blh.lap_number, blh.start_time, blh.end_time, blh.route_color, blh.halte_visit_history, blh.created_at, blh.updated_at,
			         b.vehicle_no, b.bus_number, b.plate_number, b.is_active, b.color
			  FROM bus_lap_history blh 
			  JOIN bus b ON blh.bus_id = b.id 
			  WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Add filters dynamically
	if filter.IMEI != nil {
		query += fmt.Sprintf(" AND blh.imei = $%d", argIndex)
		args = append(args, *filter.IMEI)
		argIndex++
	}

	if filter.BusID != nil {
		query += fmt.Sprintf(" AND blh.bus_id = $%d", argIndex)
		args = append(args, *filter.BusID)
		argIndex++
	}

	if filter.RouteColor != nil {
		query += fmt.Sprintf(" AND blh.route_color = $%d", argIndex)
		args = append(args, *filter.RouteColor)
		argIndex++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND blh.start_time >= $%d", argIndex)
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND blh.start_time <= $%d", argIndex)
		args = append(args, *filter.ToDate)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM blh.start_time) * 60 + EXTRACT(MINUTE FROM blh.start_time) >= $%d", argIndex)
		// Convert HH:MM to minutes since midnight
		startMinutes, err := timeStringToMinutes(*filter.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %w", err)
		}
		args = append(args, startMinutes)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM blh.start_time) * 60 + EXTRACT(MINUTE FROM blh.start_time) <= $%d", argIndex)
		// Convert HH:MM to minutes since midnight
		endMinutes, err := timeStringToMinutes(*filter.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %w", err)
		}
		args = append(args, endMinutes)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY blh.start_time DESC"

	// Add limit and offset
	if filter.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *filter.Limit)
		argIndex++
	}

	if filter.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, *filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get filtered lap history: %w", err)
	}
	defer rows.Close()

	var laps []models.BusLapHistory
	// Initialize empty slice to avoid null response
	laps = make([]models.BusLapHistory, 0)

	for rows.Next() {
		var lap models.BusLapHistory
		var endTime sql.NullTime
		var halteVisitHistory, busNumber, plateNumber sql.NullString
		err := rows.Scan(
			&lap.ID,
			&lap.BusID,
			&lap.IMEI,
			&lap.LapNumber,
			&lap.StartTime,
			&endTime,
			&lap.RouteColor,
			&halteVisitHistory,
			&lap.CreatedAt,
			&lap.UpdatedAt,
			&lap.VehicleNo,
			&busNumber,
			&plateNumber,
			&lap.IsActive,
			&lap.Color,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan filtered lap history: %w", err)
		}

		if endTime.Valid {
			lap.EndTime = &endTime.Time
		}

		if halteVisitHistory.Valid {
			lap.HalteVisitHistory = halteVisitHistory.String
		}

		if busNumber.Valid {
			lap.BusNumber = busNumber.String
		}

		if plateNumber.Valid {
			lap.PlateNumber = plateNumber.String
		}

		laps = append(laps, lap)
	}

	return laps, nil
}

func (r *repository) GetFilteredLapHistoryCount(ctx context.Context, filter dto.LapHistoryFilter) (int, error) {
	query := `SELECT COUNT(*) 
			  FROM bus_lap_history blh 
			  JOIN bus b ON blh.bus_id = b.id 
			  WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Add the same filters as GetFilteredLapHistory but for count
	if filter.IMEI != nil {
		query += fmt.Sprintf(" AND blh.imei = $%d", argIndex)
		args = append(args, *filter.IMEI)
		argIndex++
	}

	if filter.BusID != nil {
		query += fmt.Sprintf(" AND blh.bus_id = $%d", argIndex)
		args = append(args, *filter.BusID)
		argIndex++
	}

	if filter.RouteColor != nil {
		query += fmt.Sprintf(" AND blh.route_color = $%d", argIndex)
		args = append(args, *filter.RouteColor)
		argIndex++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND blh.start_time >= $%d", argIndex)
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND blh.start_time <= $%d", argIndex)
		args = append(args, *filter.ToDate)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM blh.start_time) * 60 + EXTRACT(MINUTE FROM blh.start_time) >= $%d", argIndex)
		startMinutes, err := timeStringToMinutes(*filter.StartTime)
		if err != nil {
			return 0, fmt.Errorf("invalid start_time format: %w", err)
		}
		args = append(args, startMinutes)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM blh.start_time) * 60 + EXTRACT(MINUTE FROM blh.start_time) <= $%d", argIndex)
		endMinutes, err := timeStringToMinutes(*filter.EndTime)
		if err != nil {
			return 0, fmt.Errorf("invalid end_time format: %w", err)
		}
		args = append(args, endMinutes)
		argIndex++
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to get filtered lap history count: %w", err)
	}

	return count, nil
}

// Helper function to convert time string (HH:MM) to minutes since midnight
func timeStringToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("time must be in HH:MM format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour*60 + minute, nil
}

// Debug method to get lap history count
func (r *repository) GetLapHistoryCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM bus_lap_history").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to get lap history count: %w", err)
	}
	return count, nil
}
