package middleware

import (
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type roleProtectMiddlewareFactory struct {
	config   *models.Config
	authRepo interfaces.AuthRepository
}

func NewRoleProtectMiddlewareFactory(config *models.Config, authRepo interfaces.AuthRepository) *roleProtectMiddlewareFactory {
	return &roleProtectMiddlewareFactory{
		config:   config,
		authRepo: authRepo,
	}
}

// This middleware should only be run AFTER the JWT middleware is run
func (rpmf *roleProtectMiddlewareFactory) MakeUserRoleProtector(role string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := ExtractUserEmail(r)

			user, err := rpmf.authRepo.GetUserByEmail(email)
			if err != nil {
				// We want to hide the existance of services that are protected
				// So throw a 404 route not found error
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			// We don't need to check other roles, if the user data does not exist
			// That case is handled by the if block above this
			if role == "admin" && user.Role != "admin" {
				// Same thing here
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rpmf *roleProtectMiddlewareFactory) MakeAdminApiKeyProtector() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("API_KEY") != rpmf.config.AdminApiKey {
				http.Error(w, "Invalid API Key", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
