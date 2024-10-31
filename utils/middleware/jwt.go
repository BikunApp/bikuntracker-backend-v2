package middleware

import (
	"context"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

var userContextKey = ContextKey("user-email")

func ExtractUserEmail(r *http.Request) string {
	if email, ok := r.Context().Value(userContextKey).(string); ok {
		return email
	}
	return ""
}

func JwtMiddlewareFactory(authUtil interfaces.AuthUtil) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := authUtil.ExtractJwtToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			token, err := authUtil.ToJwtToken(tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "unable to get token claims", http.StatusUnauthorized)
				return
			}

			tokenType := claims["typ"].(string)
			if tokenType != "access" {
				http.Error(w, "invalid token type, must be access token", http.StatusBadGateway)
				return
			}

			email := claims["sub"].(string)
			ctx := context.WithValue(r.Context(), userContextKey, email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
