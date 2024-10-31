package bus

import (
	"context"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/FreeJ1nG/bikuntracker-backend/utils/middleware"
)

type handler struct {
	repo interfaces.BusRepository
}

func NewHandler(repo interfaces.BusRepository) *handler {
	return &handler{
		repo: repo,
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

	res, err := h.repo.UpdateBus(ctx, id, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse[models.Bus](w, *res)
}
