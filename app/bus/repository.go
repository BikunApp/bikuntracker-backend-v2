package bus

import (
	"context"
	"database/sql"
	"fmt"

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
	rows, err := r.db.Query(ctx, `SELECT id, vehicle_no, imei, is_active, color, current_halte, next_halte, created_at, updated_at FROM bus;`)
	if err != nil {
		err = fmt.Errorf("unable to execute SQL query to get buses: %w", err)
		return
	}
	res = make([]models.Bus, 0)
	for rows.Next() {
		var bus models.Bus
		var currentHalte, nextHalte sql.NullString
		if err := rows.Scan(
			&bus.Id,
			&bus.VehicleNo,
			&bus.Imei,
			&bus.IsActive,
			&bus.Color,
			&currentHalte,
			&nextHalte,
			&bus.CreatedAt,
			&bus.UpdatedAt,
		); err != nil {
			err = fmt.Errorf("unable to scan SQL result: %w", err)
			return res, err
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
		`INSERT INTO bus (imei, vehicle_no, is_active, color) VALUES ($1, $2, $3, $4) RETURNING id, vehicle_no, imei, is_active, color, current_halte, next_halte, created_at, updated_at;`,
		data.Imei,
		data.VehicleNo,
		data.IsActive,
		data.Color,
	)

	var createdBus models.Bus
	var currentHalte, nextHalte sql.NullString
	if err = row.Scan(
		&createdBus.Id,
		&createdBus.VehicleNo,
		&createdBus.Imei,
		&createdBus.IsActive,
		&createdBus.Color,
		&currentHalte,
		&nextHalte,
		&createdBus.CreatedAt,
		&createdBus.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("unable to execute create bus SQL: %w", err)
		return
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
		sqlStr+" RETURNING id, vehicle_no, imei, is_active, color, current_halte, next_halte, created_at, updated_at",
		params...,
	)
	var updatedBus models.Bus
	var currentHalte, nextHalte sql.NullString
	if err = row.Scan(
		&updatedBus.Id,
		&updatedBus.VehicleNo,
		&updatedBus.Imei,
		&updatedBus.IsActive,
		&updatedBus.Color,
		&currentHalte,
		&nextHalte,
		&updatedBus.CreatedAt,
		&updatedBus.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("unable to execute create bus SQL: %w", err)
		return
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
		batch.Queue("INSERT INTO bus (vehicle_no, imei, is_active, color) VALUES ($1, $2, $3, $4)", bus.VehicleNo, bus.Imei, bus.IsActive, bus.Color)
	}
	err = r.db.SendBatch(ctx, batch).Close()
	if err != nil {
		err = fmt.Errorf("unable to batch insert buses: %w", err)
		return
	}
	return
}
