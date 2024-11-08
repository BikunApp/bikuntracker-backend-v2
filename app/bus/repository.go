package bus

import (
	"context"
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
	rows, err := r.db.Query(ctx, `SELECT * FROM bus;`)
	if err != nil {
		err = fmt.Errorf("Unable to execute SQL query to get buses: %w", err)
		return
	}
	res = make([]models.Bus, 0)
	for rows.Next() {
		var bus models.Bus
		if err := rows.Scan(
			&bus.Id,
			&bus.VehicleNo,
			&bus.Imei,
			&bus.IsActive,
			&bus.Color,
			&bus.CreatedAt,
			&bus.UpdatedAt,
		); err != nil {
			err = fmt.Errorf("Unable to scan SQL result: %w", err)
			return res, err
		}
		res = append(res, bus)
	}

	return
}

func (r *repository) CreateBus(ctx context.Context, data dto.CreateBusRequestBody) (res *models.Bus, err error) {
	row := r.db.QueryRow(
		ctx,
		`INSERT INTO bus (imei, vehicle_no, is_active, color)
    VALUES ($1, $2, $3, $4)
    RETURNING *;`,
		data.Imei,
		data.VehicleNo,
		data.IsActive,
		data.Color,
	)

	var createdBus models.Bus
	if err = row.Scan(
		&createdBus.Id,
		&createdBus.VehicleNo,
		&createdBus.Imei,
		&createdBus.IsActive,
		&createdBus.Color,
		&createdBus.CreatedAt,
		&createdBus.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("unable to execute create bus SQL: %w", err)
		return
	}

	res = &createdBus
	return
}

func (r *repository) UpdateBus(ctx context.Context, whereData *models.WhereData, data dto.UpdateBusRequestBody) (res *models.Bus, err error) {
	sql, params, err := utils.GetPartialUpdateSQL("bus", data, whereData)
	if err != nil {
		return
	}

	row := r.db.QueryRow(
		ctx,
		sql+" RETURNING *",
		params...,
	)

	var updatedBus models.Bus
	if err = row.Scan(
		&updatedBus.Id,
		&updatedBus.VehicleNo,
		&updatedBus.Imei,
		&updatedBus.IsActive,
		&updatedBus.Color,
		&updatedBus.CreatedAt,
		&updatedBus.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("unable to execute create bus SQL: %w", err)
		return
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
		err = fmt.Errorf("Unable to batch insert buses: %w", err)
		return
	}
	return
}
