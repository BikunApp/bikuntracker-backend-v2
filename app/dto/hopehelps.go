package dto

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

type GetReportByIdResponse struct {
	Report models.Report `json:"report"`
}

type GetReportsResponse struct {
	Reports []models.Report `json:"reports"`
}

type CreateReportRequestBody struct {
	UserId      uint   `json:"user_id"`
	Description string `json:"description"`
	Location    string `json:"location"`
	OccuredAt   uint   `json:"occured_at"`
}