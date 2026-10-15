package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code and response body.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        *bytes.Buffer
	stackTrace  []byte // captured when 5xx is written
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	// Capture stack trace when 5xx error is written
	if code >= 500 {
		rw.stackTrace = debug.Stack()
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	// Capture response body for error responses
	if rw.status >= 500 && rw.body != nil {
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// Log is a FastAPI-style logging middleware with panic recovery and stack trace logging.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			body:           &bytes.Buffer{},
		}

		// Panic recovery
		defer func() {
			if rec := recover(); rec != nil {
				wrappedWriter.status = http.StatusInternalServerError
				if wrappedWriter.stackTrace == nil {
					wrappedWriter.stackTrace = debug.Stack()
				}
				logLine := fmt.Sprintf("PANIC:	 %s - \"%s %s %s\" %d %s - %.3fms\n",
					r.RemoteAddr,
					r.Method,
					r.URL.Path,
					r.Proto,
					wrappedWriter.status,
					http.StatusText(wrappedWriter.status),
					float64(time.Since(start).Microseconds())/1000.0,
				)
				fmt.Print(logLine)
				errMsg := wrappedWriter.body.String()
				logErrorToFile(logLine, rec, wrappedWriter.stackTrace, errMsg)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Process request
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)

		// FastAPI style format
		logLine := fmt.Sprintf("INFO:	 %s - \"%s %s %s\" %d %s - %.3fms\n",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			wrappedWriter.status,
			http.StatusText(wrappedWriter.status),
			float64(duration.Microseconds())/1000.0,
		)

		// Always print to stdout
		fmt.Print(logLine)

		// Log 500 errors to file with stack trace and error message
		if wrappedWriter.status >= 500 {
			errMsg := wrappedWriter.body.String()
			logErrorToFile(logLine, nil, wrappedWriter.stackTrace, errMsg)
		}
	})
}

func logErrorToFile(logLine string, rec interface{}, stackTrace []byte, errMsg string) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("ERROR: failed to create log directory: %v\n", err)
		return
	}

	logFile := filepath.Join(logDir, "error.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("ERROR: failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(fmt.Sprintf("[%s] %s", timestamp, logLine))

	if errMsg != "" {
		f.WriteString(fmt.Sprintf("Error response: %s\n", errMsg))
	}

	if rec != nil {
		f.WriteString(fmt.Sprintf("Panic: %v\n", rec))
	}
	if stackTrace != nil && len(stackTrace) > 0 {
		f.WriteString(fmt.Sprintf("Stack trace:\n%s\n", stackTrace))
	}
}
