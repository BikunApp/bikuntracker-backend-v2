package utils

import (
	"net/http"
)

func AllowedMethodMiddlewareFactory(allowedMethods []string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			found := false

			for _, allowedMethod := range allowedMethods {
				if allowedMethod == r.Method {
					found = true
				}
			}

			if !found {
				http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
