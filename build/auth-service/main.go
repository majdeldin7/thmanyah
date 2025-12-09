package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// -------- Response Struct --------
type Response struct {
	Auth    bool   `json:"auth"`
	Message string `json:"message"`
}

// -------- Prometheus Metrics (use a custom registry) --------
var (
	registry = prometheus.NewRegistry()

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// -------- DB Connection (global) --------
var db *pgxpool.Pool

// -------- Middleware for metrics --------
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.statusCode)).Observe(duration)
	})
}

// -------- Helper to record status codes --------
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// -------- Handlers --------
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	resp := Response{Auth: true, Message: "auth-service working"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("live"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

// NEW: /db-check
func dbCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var now time.Time
	err := db.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	if err != nil {
		http.Error(w, "DB NOT OK: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := Response{
		Auth:    true,
		Message: "DB connected at " + now.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// -------- main --------
func main() {
	ctx := context.Background()

	// ---- Read env vars (same as Node.js) ----
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	// ---- Build connection string ----
	connStr := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName

	// ---- Connect to DB ----
	var err error
	db, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("FAILED TO INIT DB POOL: " + err.Error())
	}

	// ---- Test DB connection on startup ----
	if err := db.Ping(ctx); err != nil {
		panic("FAILED TO CONNECT TO DB: " + err.Error())
	}

	// ---- Routes ----
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("/live", liveHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/db-check", dbCheckHandler)

	// Prometheus metrics
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// ---- Start server ----
	http.ListenAndServe(":4000", metricsMiddleware(mux))
}