package auth

import (
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/FreeJ1nG/bikuntracker-backend/utils/middleware"
	"github.com/jackc/pgx/v5"
)

type handler struct {
	service interfaces.AuthService
	repo    interfaces.AuthRepository
}

func NewHandler(authService interfaces.AuthService, authRepo interfaces.AuthRepository) *handler {
	return &handler{
		service: authService,
		repo:    authRepo,
	}
}

func (h *handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	email := middleware.ExtractUserEmail(r)

	user, err := h.repo.GetUserByEmail(email)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.EncodeSuccessResponse(w, user)
}

func (h *handler) SsoLogin(w http.ResponseWriter, r *http.Request) {
	body, err := utils.ParseRequestBody[dto.SSOLoginRequestBody](r.Body)

	res, status, err := h.service.SsoLogin(body.Ticket, body.Service)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	utils.EncodeSuccessResponse(w, res)
}

func (h *handler) RefreshJwt(w http.ResponseWriter, r *http.Request) {
	body, err := utils.ParseRequestBody[dto.RefreshTokenRequestBody](r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, status, err := h.service.RefreshToken(body.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	utils.EncodeSuccessResponse(w, res)
}
