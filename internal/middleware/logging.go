package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

// LoggingMiddleware logs the incoming HTTP request and its duration
func LoggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			// Get request ID from context or generate new one
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = time.Now().Format("20060102150405.000000")
			}

			// Log incoming request
			logger.Printf(
				"REQUEST: [%s] %s %s %s",
				requestID,
				r.Method,
				r.RequestURI,
				r.RemoteAddr,
			)

			// Add request ID to response headers
			wrapped.Header().Set("X-Request-ID", requestID)

			defer func() {
				// Recover from panic
				if err := recover(); err != nil {
					// Log stack trace
					logger.Printf(
						"PANIC: [%s] %v\n%s",
						requestID,
						err,
						debug.Stack(),
					)
					wrapped.WriteHeader(http.StatusInternalServerError)
				}

				// Get status code if not set
				if !wrapped.wroteHeader {
					wrapped.WriteHeader(http.StatusOK)
				}

				// Log request completion
				logger.Printf(
					"RESPONSE: [%s] %s %s %d %s %v",
					requestID,
					r.Method,
					r.RequestURI,
					wrapped.status,
					r.RemoteAddr,
					time.Since(start),
				)
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

// RequestLogging adds structured request logging
func RequestLogging(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			// Log request details
			logRequest(logger, r)

			next.ServeHTTP(wrapped, r)

			// Log response details
			logResponse(logger, r, wrapped, time.Since(start))
		})
	}
}

// logRequest logs detailed request information
func logRequest(logger *log.Logger, r *http.Request) {
	logger.Printf("Request Details:")
	logger.Printf("  Method: %s", r.Method)
	logger.Printf("  URL: %s", r.URL.String())
	logger.Printf("  Protocol: %s", r.Proto)
	logger.Printf("  Remote Address: %s", r.RemoteAddr)

	// Log headers
	logger.Println("  Headers:")
	for name, headers := range r.Header {
		for _, h := range headers {
			logger.Printf("    %v: %v", name, h)
		}
	}
}

// logResponse logs response details
func logResponse(logger *log.Logger, r *http.Request, w *responseWriter, duration time.Duration) {
	logger.Printf("Response Details:")
	logger.Printf("  Request: %s %s", r.Method, r.URL.String())
	logger.Printf("  Status: %d", w.status)
	logger.Printf("  Duration: %v", duration)
	logger.Printf("  Size: %v", w.Header().Get("Content-Length"))
}

// ErrorLogging middleware for handling and logging errors
func ErrorLogging(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					logger.Printf("ERROR: Recovered from panic: %v", err)
					logger.Printf("Stack Trace:\n%s", debug.Stack())

					// Return Internal Server Error
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitLogging middleware to log rate limit information
func RateLimitLogging(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Example rate limit headers
			remaining := w.Header().Get("X-RateLimit-Remaining")
			limit := w.Header().Get("X-RateLimit-Limit")

			if remaining != "" && limit != "" {
				logger.Printf("Rate Limit: %s/%s remaining", remaining, limit)
			}

			next.ServeHTTP(w, r)
		})
	}
}
