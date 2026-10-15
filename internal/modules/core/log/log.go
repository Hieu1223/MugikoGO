package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

// If Write is called without explicit WriteHeader, it defaults to 200 OK in net/http
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Log is a FastAPI-style logging middleware.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer
		wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Process request
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)

		// FastAPI style format: [timestamp] "METHOD path HTTP/version" status_code
		// Example output: INFO:     127.0.0.1:54321 - "GET /items/42 HTTP/1.1" 200 OK 1.23ms

		fmt.Printf("INFO:\t %s - \"%s %s %s\" %d %s - %.3fms\n",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			wrappedWriter.status,
			http.StatusText(wrappedWriter.status),
			float64(duration.Microseconds())/1000.0,
		)
	})
}
