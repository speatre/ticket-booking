package metrics

import (
	"context"
	"encoding/json"
	"net/http"

	"ticket-booking/internal/booking"
	"ticket-booking/pkg/cache"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics holds dependencies for metrics collection
type Metrics struct {
	bookingRepo booking.BookingRepository
	cache       *cache.Redis
	logger      *zap.Logger
}

// Stats holds tickets sold and revenue for an event
type Stats struct {
	Tickets int     `json:"tickets"`
	Revenue float64 `json:"revenue"`
}

// Prometheus metrics
var (
	TicketsSold = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tickets_sold_total",
			Help: "Total tickets sold per event",
		},
		[]string{"event_id"},
	)
	Revenue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "revenue_total",
			Help: "Total revenue from confirmed bookings per event",
		},
		[]string{"event_id"},
	)
)

// NewMetrics initializes metrics with repo, cache, and logger
func NewMetrics(repo booking.BookingRepository, cacheClient *cache.Redis, logger *zap.Logger) *Metrics {
	// Register metrics
	prometheus.MustRegister(TicketsSold, Revenue)
	return &Metrics{
		bookingRepo: repo,
		cache:       cacheClient,
		logger:      logger,
	}
}

// UpdateMetrics fetches stats from Redis or DB and updates Prometheus metrics
func (m *Metrics) UpdateMetrics(ctx context.Context) {
	eventIDs, err := m.cache.GetEventIDs(ctx)
	if err != nil {
		m.logger.Error("Failed to get event IDs from cache",
			zap.Error(err))
		return
	}

	// Silent operation - metrics are for monitoring, not verbose logging

	for _, eventID := range eventIDs {
		tickets, revenue, err := m.getStats(ctx, eventID)
		if err != nil {
			m.logger.Error("Failed to get stats for event",
				zap.String("event_id", eventID),
				zap.Error(err))
			continue
		}
		// Update Prometheus metrics silently - no logging for successful operations
		TicketsSold.WithLabelValues(eventID).Set(float64(tickets))
		Revenue.WithLabelValues(eventID).Set(revenue)
	}
}

// getStats fetches tickets sold & revenue for an event
func (m *Metrics) getStats(ctx context.Context, eventID string) (int, float64, error) {
	cacheKey := "booking:stats:" + eventID

	// Try Redis first
	val, err := m.cache.Get(ctx, cacheKey)
	if err == nil && val != "" {
		var stats Stats
		if err := json.Unmarshal([]byte(val), &stats); err == nil {
			return stats.Tickets, stats.Revenue, nil
		}
		m.logger.Warn("Failed to unmarshal cached stats",
			zap.String("event_id", eventID),
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	}

	// Fallback to DB
	bookings, err := m.bookingRepo.ListConfirmedByEvent(ctx, eventID)
	if err != nil {
		m.logger.Error("Failed to fetch confirmed bookings from DB",
			zap.String("event_id", eventID),
			zap.Error(err))
		return 0, 0, err
	}

	var totalTickets int
	var totalRevenue float64
	for _, b := range bookings {
		totalTickets += b.Quantity
		// convert cents to dollars for the metric
		const centsToDollars = 100.0
		totalRevenue += float64(b.Quantity) * float64(b.UnitPriceCents) / centsToDollars
	}

	// Update Redis for future
	stats := Stats{Tickets: totalTickets, Revenue: totalRevenue}
	data, err := json.Marshal(stats)
	if err != nil {
		m.logger.Warn("Failed to marshal stats for cache",
			zap.String("event_id", eventID),
			zap.Error(err))
	} else {
		if err := m.cache.Set(ctx, cacheKey, string(data), 0); err != nil {
			m.logger.Warn("Failed to set stats in cache",
				zap.String("event_id", eventID),
				zap.String("cache_key", cacheKey),
				zap.Error(err))
		}
		// No logging for successful cache operations - this is operational noise
	}

	return totalTickets, totalRevenue, nil
}

// MetricsServer holds the HTTP server for metrics
type MetricsServer struct {
	server *http.Server
}

// StartHTTPServer exposes /metrics endpoint with graceful shutdown support
func StartHTTPServer(addr string) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	metricsServer := &MetricsServer{server: server}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Note: Cannot use zap here as logger is not passed; consider logging in caller
			panic(err)
		}
	}()

	return metricsServer
}

// Shutdown gracefully shuts down the metrics server
func (m *MetricsServer) Shutdown(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}
