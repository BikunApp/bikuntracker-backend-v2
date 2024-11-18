package interfaces

import (
	"context"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type ReportRepository interface {
	GetReports(ctx context.Context) (reports []models.Report, err error)
	GetReportById(ctx context.Context, id string) (report *models.Report, err error)
	CreateReport(ctx context.Context, data *dto.CreateReportRequestBody) (report *models.Report, err error)
	DeleteReport(ctx context.Context, id string) (err error)
}

type ReportService interface {
	GetReports(ctx context.Context) (res *dto.GetReportsResponse, err error)
	GetReportById(ctx context.Context, id string) (res *dto.GetReportByIdResponse, err error)
	CreateReport(ctx context.Context, data *dto.CreateReportRequestBody) (res *dto.CreateReportResponse, err error)
	DeleteReport(ctx context.Context, id string) (err error)
}