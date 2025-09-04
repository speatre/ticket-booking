package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Validate checks the configuration for validity and reasonable values
func (c *Config) Validate() error {
	var errors []string

	// Server validation
	if err := c.validateServer(); err != nil {
		errors = append(errors, fmt.Sprintf("server: %v", err))
	}

	// Security validation
	if err := c.validateSecurity(); err != nil {
		errors = append(errors, fmt.Sprintf("security: %v", err))
	}

	// Database validation
	if err := c.validatePostgres(); err != nil {
		errors = append(errors, fmt.Sprintf("postgres: %v", err))
	}

	// Redis validation
	if err := c.validateRedis(); err != nil {
		errors = append(errors, fmt.Sprintf("redis: %v", err))
	}

	// RabbitMQ validation
	if err := c.validateRabbitMQ(); err != nil {
		errors = append(errors, fmt.Sprintf("rabbitmq: %v", err))
	}

	// Logging validation
	if err := c.validateLogging(); err != nil {
		errors = append(errors, fmt.Sprintf("logging: %v", err))
	}

	// Booking validation
	if err := c.validateBooking(); err != nil {
		errors = append(errors, fmt.Sprintf("booking: %v", err))
	}

	// Worker validation
	if err := c.validateWorker(); err != nil {
		errors = append(errors, fmt.Sprintf("worker: %v", err))
	}

	// Observability validation
	if err := c.validateObservability(); err != nil {
		errors = append(errors, fmt.Sprintf("observability: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

func (c *Config) validateServer() error {
	var errors []string

	// Validate HTTP address
	if c.Server.HTTPAddr == "" {
		errors = append(errors, "http_addr is required")
	} else if _, err := net.ResolveTCPAddr("tcp", c.Server.HTTPAddr); err != nil {
		errors = append(errors, fmt.Sprintf("invalid http_addr format: %v", err))
	}

	// Validate metrics address
	if c.Server.MetricsAddr == "" {
		errors = append(errors, "metrics_addr is required")
	} else if _, err := net.ResolveTCPAddr("tcp", c.Server.MetricsAddr); err != nil {
		errors = append(errors, fmt.Sprintf("invalid metrics_addr format: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateSecurity() error {
	var errors []string

	if c.Security.AccessTTLMinute <= 0 {
		errors = append(errors, "access_ttl_minutes must be positive")
	}
	if c.Security.AccessTTLMinute > 60*24 { // Max 24 hours
		errors = append(errors, "access_ttl_minutes too large (>1440)")
	}

	if c.Security.RefreshTTLMinute <= 0 {
		errors = append(errors, "refresh_ttl_minutes must be positive")
	}
	if c.Security.RefreshTTLMinute < c.Security.AccessTTLMinute {
		errors = append(errors, "refresh_ttl_minutes must be >= access_ttl_minutes")
	}

	if len(c.Security.JWTAccessSecret) < 16 {
		errors = append(errors, "jwt_access_secret too short (<16 chars)")
	}
	if len(c.Security.JWTRefreshSecret) < 16 {
		errors = append(errors, "jwt_refresh_secret too short (<16 chars)")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validatePostgres() error {
	var errors []string

	if c.Postgres.DSN == "" {
		errors = append(errors, "dsn is required")
	} else {
		// Parse DSN to validate format
		if !strings.Contains(c.Postgres.DSN, "host=") ||
			!strings.Contains(c.Postgres.DSN, "user=") ||
			!strings.Contains(c.Postgres.DSN, "dbname=") {
			errors = append(errors, "dsn must contain host, user, and dbname parameters")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateRedis() error {
	var errors []string

	if c.Redis.Addr == "" {
		errors = append(errors, "addr is required")
	} else if _, err := net.ResolveTCPAddr("tcp", c.Redis.Addr); err != nil {
		errors = append(errors, fmt.Sprintf("invalid addr format: %v", err))
	}

	if c.Redis.DB < 0 {
		errors = append(errors, "db must be >= 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateRabbitMQ() error {
	var errors []string

	if c.RabbitMQ.URL == "" {
		errors = append(errors, "url is required")
	} else {
		if parsed, err := url.Parse(c.RabbitMQ.URL); err != nil {
			errors = append(errors, fmt.Sprintf("invalid url format: %v", err))
		} else if parsed.Scheme != "amqp" && parsed.Scheme != "amqps" {
			errors = append(errors, "url must use amqp or amqps scheme")
		}
	}

	if c.RabbitMQ.PaymentQueue == "" {
		errors = append(errors, "payment_queue is required")
	}
	if c.RabbitMQ.CancelQueue == "" {
		errors = append(errors, "cancel_queue is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateLogging() error {
	var errors []string

	if c.Logging.Dir == "" {
		errors = append(errors, "dir is required")
	}

	if c.Logging.RetentionDays <= 0 {
		errors = append(errors, "retention_days must be positive")
	}
	if c.Logging.RetentionDays > 365 {
		errors = append(errors, "retention_days too large (>365)")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateBooking() error {
	var errors []string

	if c.Booking.AutoCancelMinutes <= 0 {
		errors = append(errors, "auto_cancel_minutes must be positive")
	}
	if c.Booking.AutoCancelMinutes > 60*24 { // Max 24 hours
		errors = append(errors, "auto_cancel_minutes too large (>1440)")
	}

	if c.Booking.PageDefaultLimit <= 0 {
		errors = append(errors, "page_default_limit must be positive")
	}
	if c.Booking.PageDefaultLimit > c.Booking.PageMaxLimit {
		errors = append(errors, "page_default_limit must be <= page_max_limit")
	}

	if c.Booking.PageMaxLimit <= 0 {
		errors = append(errors, "page_max_limit must be positive")
	}
	if c.Booking.PageMaxLimit > DefaultMaxPageLimit {
		errors = append(errors, fmt.Sprintf("page_max_limit too large (>%d)", DefaultMaxPageLimit))
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateWorker() error {
	var errors []string

	if c.Worker.PollerIntervalSeconds <= 0 {
		errors = append(errors, "poller_interval_seconds must be positive")
	}
	if c.Worker.PollerIntervalSeconds > 3600 { // Max 1 hour
		errors = append(errors, "poller_interval_seconds too large (>3600)")
	}

	if c.Worker.PaymentSuccessRate < 0 || c.Worker.PaymentSuccessRate > 100 {
		errors = append(errors, "payment_success_rate must be between 0 and 100")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateObservability() error {
	var errors []string

	if c.Observability.MetricsUpdateSeconds <= 0 {
		errors = append(errors, "metrics_update_seconds must be positive")
	}
	if c.Observability.MetricsUpdateSeconds > 3600 { // Max 1 hour
		errors = append(errors, "metrics_update_seconds too large (>3600)")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}
