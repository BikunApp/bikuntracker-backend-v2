package utils

import (
	"net/http"
	"strings"

	"github.com/FreeJ1nG/bikuntracker-backend/utils/middleware"
)

type MethodHandler = map[string]http.HandlerFunc

type Handler interface {
	MethodHandler | http.HandlerFunc
}

type MethodSpecificMiddlewares = map[string][]middleware.Middleware

type Options struct {
	Middlewares               []middleware.Middleware
	MethodSpecificMiddlewares MethodSpecificMiddlewares
}

func initialHandlerFactory[H Handler](handler H) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch handler := any(handler).(type) {
		case http.HandlerFunc:
			// Case where the provided handler is used right away, this will only work on GET requests
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler.ServeHTTP(w, r)
		case MethodHandler:
			// Create the base handler, which serves handlers conditionally based on the request method
			if handler, ok := handler[r.Method]; ok {
				handler.ServeHTTP(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			panic("Unsupported handler type")
		}
	})
}

func HandleRoute[H Handler](path string, handler H, options *Options) {
	// Also apply default middlewares
	allMiddlewares := []middleware.Middleware{
		middleware.LoggerMiddleware, // This will be executed first
		middleware.CorsMiddleware,
		middleware.ParseRouteParamsMiddlewareFactory(path), // This will be executed last
	}

	if options != nil && options.Middlewares != nil {
		// Also apply any other middleware
		for _, middleware := range options.Middlewares {
			allMiddlewares = append(allMiddlewares, middleware)
		}
	}

	if options != nil && options.MethodSpecificMiddlewares != nil && len(options.MethodSpecificMiddlewares) > 0 {
		// So this complicated logic ensures that each method specific middleware is applied (and combined into one middleware)
		allMiddlewares = append(allMiddlewares, func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if middlewares, ok := options.MethodSpecificMiddlewares[r.Method]; ok {
					// next is the next handler in the middleware chain (most likely going to be the base handler)
					currentHandler := next
					for i := len(middlewares) - 1; i >= 0; i-- {
						// Don't forget to apply in reverse just like below
						currentHandler = middlewares[i](currentHandler)
					}
					// After getting the final handler wrapped by each middleware, serve it
					currentHandler.ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			}
		})
	}

	currentHandler := initialHandlerFactory[H](handler)
	for i := len(allMiddlewares) - 1; i >= 0; i-- {
		currentHandler = allMiddlewares[i](currentHandler)
	}

	var resolvedPath string
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) > 0 && strings.Contains(pathSplit[len(pathSplit)-1], ":") {
		// If the path contains a dynamic segment *at the end* (this is important)
		// Then delete the dynamic segment and allow http.HandleFunc to match the path with a trailling slash
		// For example /bus/:id will be transformed to /bus/, meaning any calls to /bus/5, /bus/10, etc will be matched
		resolvedPath = strings.Join(pathSplit[:len(pathSplit)-1], "/") + "/"
	} else {
		resolvedPath = path
	}

	http.HandleFunc(resolvedPath, currentHandler.ServeHTTP)
}
