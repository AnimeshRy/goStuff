package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

type CalculationRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type CalculationResponse struct {
	Result int    `json:"result"`
	Error  string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// Initialize the structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Create a router and register handlers
	mux := http.NewServeMux()

	// Register calculation endpoints
	mux.HandleFunc("/add", loggingMiddleware(addHandler))
	mux.HandleFunc("/subtract", loggingMiddleware(subtractHandler))
	mux.HandleFunc("/multiply", loggingMiddleware(multiplyHandler))
	mux.HandleFunc("/divide", loggingMiddleware(divideHandler))

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	})

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next(rw, r)

		// Log the request details
		slog.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func writeResponse(w http.ResponseWriter, result int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CalculationResponse{Result: result})
}

func validateRequest(w http.ResponseWriter, r *http.Request) (*CalculationRequest, bool) {
	// Verify HTTP method
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, false
	}

	// Parse JSON request body
	var req CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return nil, false
	}

	return &req, true
}

// Handler functions
func addHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequest(w, r)
	if !ok {
		return
	}

	writeResponse(w, req.A+req.B)
}

func subtractHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequest(w, r)
	if !ok {
		return
	}

	writeResponse(w, req.A-req.B)
}

func multiplyHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequest(w, r)
	if !ok {
		return
	}

	writeResponse(w, req.A*req.B)
}

func divideHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequest(w, r)
	if !ok {
		return
	}

	// Check for division by zero
	if req.B == 0 {
		writeError(w, "Division by zero is not allowed", http.StatusBadRequest)
		return
	}

	writeResponse(w, req.A/req.B)
}
