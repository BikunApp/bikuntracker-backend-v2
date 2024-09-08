package utils

import (
	"net/http"
)

type Middleware = func(http.Handler) http.Handler

func HandleRoute(path string, handler http.Handler, middlewares []Middleware) {
	// also apply default middlewares
	allMiddlewares := []Middleware{
		CorsMiddleware,
		LoggerMiddleware,
	}

	if len(middlewares) > 0 {
		allMiddlewares = append(allMiddlewares, middlewares...)
	}

	currentHandler := handler
	for _, middleware := range allMiddlewares {
		currentHandler = middleware(currentHandler)
	}

	http.HandleFunc(path, currentHandler.ServeHTTP)
}
