package hopehelps

import (
	"context"
	"fmt"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
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

func (r *repository) GetReports(ctx context.Context) (res []models.Report, err error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM report;")
	if err != nil {
		err = fmt.Errorf("unable to do SQL Query to get reports: %s", err.Error())
		return
	}
	defer rows.Close()

	res = make([]models.Report, 0)
	for rows.Next() {
		var report models.Report
		if err = rows.Scan(
			&report.Id,
			&report.UserId,
			&report.Description,
			&report.Location,
			&report.OccuredAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		) ; err != nil {
			err = fmt.Errorf("error while parsing report object from SQL: %s", err.Error())
			return
		}
		res = append(res, report)
	}

	return
}

func (r *repository) GetReportById(ctx context.Context, id string) (res *models.Report, err error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT * FROM report
		WHERE id=$1`,
		id,
	)

	var report models.Report
	if err = row.Scan(
		&report.Id,
		&report.UserId,
		&report.Description,
		&report.Location,
		&report.OccuredAt,
		&report.CreatedAt,
		&report.UpdatedAt,
	); err != nil {
		err = fmt.Errorf("error while parsing created report object from SQL: %s", err.Error())
		return
	}

	res = &report
	return
}

func (r *repository) CreateReport(ctx context.Context, data *dto.CreateReportRequestBody) (res *models.Report, err error) {
	row := r.db.QueryRow(
		ctx, 
		`INSERT INTO report (user_id, description, location, occured_at)
		VALUES ($1, $2, $3, $4)
		RETURNING *`,
		data.UserId,
		data.Description,
		data.Location,
		data.OccuredAt,
	)

	var createdReport models.Report
	if err = row.Scan(
		&createdReport.Id,
		&createdReport.UserId,
		&createdReport.Description,
		&createdReport.Location,
		&createdReport.OccuredAt,
		&createdReport.CreatedAt,
		&createdReport.UpdatedAt,
	) ; err != nil {
		err = fmt.Errorf("error while parsing created report object from SQL: %s", err.Error())
		return
	}

	res = &createdReport
	return 
}

func (r *repository) DeleteReport(ctx context.Context, id string) (err error) {
	_, err = r.db.Exec(ctx, `DELETE FROM report WHERE id=$1`, id)
	return
}