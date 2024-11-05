package hopehelps

import (
	"context"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
)

type service struct {
	repo interfaces.ReportRepository
}

func NewService(repo interfaces.ReportRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetReports(ctx context.Context) (res *dto.GetReportsResponse, err error) {
	resData, err := s.repo.GetReports(ctx)
	if err != nil {return}
	res = &dto.GetReportsResponse{
		Reports: resData,
	}
	return
}

func (s *service) GetReportById(ctx context.Context, id string) (res *dto.GetReportByIdResponse, err error) {
	resData, err := s.repo.GetReportById(ctx, id)
	if err != nil {return}
	res = &dto.GetReportByIdResponse{
		Report: *resData,
	}
	return
}

func (s *service) CreateReport(ctx context.Context, data *dto.CreateReportRequestBody) (res *dto.CreateReportResponse, err error) {
	resData, err := s.repo.CreateReport(ctx, data)
	if err != nil {return}
	res = &dto.CreateReportResponse{
		Report: *resData,
	}
	return
}

func (s *service) DeleteReport(ctx context.Context, id string) (err error) {
	err = s.repo.DeleteReport(ctx, id)
	return
}