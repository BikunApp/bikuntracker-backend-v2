package utils

import (
	"net/http"
)

type Middleware = func(http.HandlerFunc) http.HandlerFunc

type Options struct {
	Middlewares    []Middleware
	AllowedMethods []string
}

func HandleRoute(path string, handler http.HandlerFunc, options *Options) {
	// also apply default middlewares
	allMiddlewares := []Middleware{
		LoggerMiddleware,
		CorsMiddleware,
	}

	if options != nil && options.AllowedMethods != nil && len(options.AllowedMethods) > 0 {
		allMiddlewares = append(allMiddlewares, AllowedMethodMiddlewareFactory(options.AllowedMethods))
	}

	if options != nil && options.Middlewares != nil {
		for _, middleware := range options.Middlewares {
			allMiddlewares = append(allMiddlewares, middleware)
		}
	}

	currentHandler := handler
	for i := len(allMiddlewares) - 1; i >= 0; i-- {
		currentHandler = allMiddlewares[i](currentHandler)
	}

	http.HandleFunc(path, currentHandler.ServeHTTP)
}
