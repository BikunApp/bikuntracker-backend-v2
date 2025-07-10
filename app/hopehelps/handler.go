package hopehelps

import (
	"context"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

type handler struct {
	service interfaces.ReportService
}

func NewHandler(service interfaces.ReportService) *handler {
	return &handler{
		service: service,
	}
}

var tes interfaces.ReportService

func (h *handler) GetReports(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	res, err := h.service.GetReports(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse(w, res)
	return
}

func (h *handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := utils.ParseRequestBody[dto.CreateReportRequestBody](r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.service.CreateReport(ctx, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse(w, res)
	return
}