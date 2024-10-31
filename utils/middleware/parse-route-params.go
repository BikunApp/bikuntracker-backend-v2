package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const routeParamsKey = "route-params"

func ExtractRouteParams(r *http.Request) map[string]string {
	res := make(map[string]string)
	if res, ok := r.Context().Value(routeParamsKey).(map[string]string); ok {
		return res
	}
	return res
}

func GetRouteParam(r *http.Request, key string) (res string, status int, err error) {
	status = http.StatusOK

	routeParams := ExtractRouteParams(r)
	res, ok := routeParams[key]
	if !ok {
		err = fmt.Errorf("Unable to extract field %s from request url", key)
		status = http.StatusBadRequest
		return
	}

	return
}

func ParseRouteParamsMiddlewareFactory(path string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestUrl := r.URL.Path
			requestUrlSplit := strings.Split(requestUrl, "/")
			pathSplit := strings.Split(path, "/")

			log.Println(" >> request url", requestUrlSplit)
			log.Println(" >> path split", pathSplit)

			if len(requestUrlSplit) != len(pathSplit) {
				http.Error(w, "Invalid URL format, not able to supply route params", http.StatusBadRequest)
				return
			}

			routeParams := make(map[string]string)
			for i, p := range pathSplit {
				if strings.HasPrefix(p, ":") {
					// This means that this path must have a variable associated with it
					variableName := p[1:]
					routeParams[variableName] = requestUrlSplit[i]
				}
			}

			ctx := context.WithValue(r.Context(), routeParamsKey, routeParams)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
	}
}
