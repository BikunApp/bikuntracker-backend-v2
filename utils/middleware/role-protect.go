package middleware

import (
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
)

// This middleware should only be run AFTER the JWT middleware is run
func RoleProtectMiddlewareFactory(authRepo interfaces.AuthRepository, role string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := ExtractUserEmail(r)

			user, err := authRepo.GetUserByEmail(email)
			if err != nil {
				// We want to hide the existance of services that are protected
				// So throw a 404 route not found error
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			if user.Role != role {
				// Same thing here
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
