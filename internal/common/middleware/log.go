package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r) // Call the next handler
		fmt.Printf("%s %f\n", r.URL.String(), time.Since(start).Seconds())
	})
}
