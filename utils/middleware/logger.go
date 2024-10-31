package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijacker interface not supported")
	}
	return hj.Hijack()
}

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &responseWriter{w, http.StatusOK}
		now := time.Now()
		defer func() {
			fmt.Println("================== START OF REQUEST ==================")
			fmt.Printf(
				"method=%s, url=%s, host=%s, path=%s, duration=%s, status=%d\n",
				r.Method,
				r.RequestURI,
				r.Host,
				r.URL.Path,
				time.Since(now).String(),
				writer.status,
			)
			fmt.Println("================== END OF REQUEST ==================")
			fmt.Println()
		}()
		next.ServeHTTP(writer, r)
	})
}
