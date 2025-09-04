package config

import (
	"fmt"
	"strings"
	"time"
)

// Application Constants
const (
	// Default Application Settings
	DefaultAppName        = "ticket-booking"
	DefaultHTTPAddr       = ":8080"
	DefaultMetricsAddr    = ":8081"
	DefaultLogDir         = "logs"
	DefaultConfigFile     = "configs/app.yaml"

	// Environment Names
	EnvLocal      = "local"
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"

	// Default Limits
	DefaultPageSize     = 20
	DefaultMaxPageSize  = 100
	DefaultMaxPageLimit = 1000
)

// Database Constants
const (
	// PostgreSQL Defaults
	DefaultPostgresHost     = "localhost"
	DefaultPostgresPort     = "5432"
	DefaultPostgresUser     = "postgres"
	DefaultPostgresPassword = "postgres"
	DefaultPostgresDBName   = "ticket_booking"
	DefaultPostgresSSLMode  = "disable"

	// PostgreSQL Connection Pool
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 10
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 2 * time.Minute
)

// Redis Constants
const (
	DefaultRedisHost = "localhost"
	DefaultRedisPort = "6379"
	DefaultRedisDB   = 0
	DefaultRedisTTL  = 24 * time.Hour
)

// RabbitMQ Constants
const (
	DefaultRabbitMQHost     = "localhost"
	DefaultRabbitMQPort     = "5672"
	DefaultRabbitMQUser     = "guest"
	DefaultRabbitMQPassword = "guest"
	DefaultRabbitMQVHost    = "/"

	// Queue Names
	DefaultPaymentQueue   = "payment_queue"
	DefaultCancelQueue    = "cancel_delay_queue"
	DefaultBookingExchange = "booking"
	DefaultExchangeType   = "topic"
	DefaultRoutingKey     = "booking.#"
)

// JWT Constants
const (
	DefaultAccessTTLMinutes  = 15
	DefaultRefreshTTLMinutes = 7 * 24 * 60 // 7 days
	DefaultJWTIssuer         = "ticket-booking-api"
)

// Security Constants
const (
	DefaultJWTSecretLength    = 32
	DefaultPasswordMinLength  = 8
	DefaultPasswordMaxLength  = 128
	DefaultRateLimitRequests  = 100
	DefaultRateLimitWindow    = time.Minute
	DefaultMaxLoginAttempts   = 5
	DefaultLockoutDuration    = 15 * time.Minute
)

// Booking Constants
const (
	DefaultAutoCancelMinutes    = 15
	DefaultPaymentTimeoutMinutes = 10
	DefaultMaxTicketsPerBooking  = 10
	DefaultMinTicketsPerBooking  = 1
)

// Worker Constants
const (
	DefaultPollerIntervalSeconds = 60
	DefaultPaymentSuccessRate    = 90 // percentage
	DefaultWorkerPoolSize        = 5
	DefaultWorkerTimeoutSeconds  = 30
)

// Observability Constants
const (
	DefaultMetricsUpdateSeconds = 15
	DefaultHealthCheckTimeout   = 5 * time.Second
	DefaultReadinessTimeout     = 10 * time.Second
)

// Logging Constants
const (
	DefaultLogRetentionDays    = 30
	DefaultLogRotationHours    = 24
	DefaultMetricsLogRetention = 7
	DefaultAccessLogRetention  = 30
	DefaultErrorLogRetention   = 90
)

// HTTP Constants
const (
	DefaultReadTimeout     = 15 * time.Second
	DefaultWriteTimeout    = 15 * time.Second
	DefaultIdleTimeout     = 60 * time.Second
	DefaultMaxHeaderBytes  = 1 << 20 // 1MB
	DefaultShutdownTimeout = 30 * time.Second
)

// Cache Constants
const (
	DefaultCacheTTLShort  = 5 * time.Minute
	DefaultCacheTTLMedium = 30 * time.Minute
	DefaultCacheTTLLong   = 2 * time.Hour
	DefaultCachePrefix    = "ticket:"
)

// File Constants
const (
	DefaultUploadMaxSize     = 10 << 20 // 10MB
	DefaultUploadTimeout     = 5 * time.Minute
	DefaultTempFilePrefix    = "upload_"
	DefaultAllowedFileTypes  = "jpg,jpeg,png,pdf"
)

// Email Constants
const (
	DefaultEmailTimeout      = 30 * time.Second
	DefaultEmailRetries      = 3
	DefaultEmailTemplateDir  = "templates/email"
	DefaultEmailFromAddress  = "noreply@ticket-booking.com"
	DefaultEmailFromName     = "Ticket Booking"
)

// Notification Constants
const (
	DefaultSMSRetries         = 2
	DefaultSMSTimeout         = 10 * time.Second
	DefaultPushNotificationTTL = 24 * time.Hour
	DefaultWebhookTimeout     = 30 * time.Second
	DefaultWebhookRetries     = 3
)

// Feature Flags
const (
	DefaultEnableMetrics      = true
	DefaultEnableTracing      = false
	DefaultEnableProfiling    = false
	DefaultEnableHealthChecks = true
	DefaultEnableRateLimiting = true
	DefaultEnableCaching      = true
	DefaultEnableQueue        = true
)

// Validation Constants
const (
	DefaultMaxEventNameLength    = 100
	DefaultMaxEventDescLength    = 1000
	DefaultMaxUserNameLength     = 50
	DefaultMaxEmailLength        = 254
	DefaultMaxPhoneLength        = 20
	DefaultMinEventPrice         = 0
	DefaultMaxEventPrice         = 1000000 // $10,000
	DefaultMaxEventCapacity      = 100000
	DefaultMinEventDuration      = time.Minute
	DefaultMaxEventDuration      = 24 * time.Hour * 365 // 1 year
)

// Environment represents different deployment environments
type Environment string

// IsValid checks if the environment string is valid
func (e Environment) IsValid() bool {
	switch e {
	case EnvLocal, EnvDevelopment, EnvStaging, EnvProduction:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (e Environment) String() string {
	return string(e)
}

// ParseEnvironment parses a string into Environment
func ParseEnvironment(s string) (Environment, error) {
	env := Environment(strings.ToLower(s))
	if !env.IsValid() {
		return "", fmt.Errorf("invalid environment: %s (must be one of: local, development, staging, production)", s)
	}
	return env, nil
}

// IsProduction returns true if this is a production environment
func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// IsDevelopment returns true if this is a development environment
func (e Environment) IsDevelopment() bool {
	return e == EnvDevelopment || e == EnvLocal
}

// GetConfigForEnv returns environment-specific configuration defaults
func GetConfigForEnv(env Environment) *Config {
	config := &Config{}
	config.SetDefaults()

	// Environment-specific overrides
	switch env {
	case EnvProduction:
		config.Logging.RetentionDays = 90
		config.Worker.PaymentSuccessRate = 95
	case EnvStaging:
		config.Logging.RetentionDays = 30
		config.Worker.PaymentSuccessRate = 90
	case EnvDevelopment:
		config.Logging.RetentionDays = 7
		config.Worker.PaymentSuccessRate = 85
	case EnvLocal:
		config.Logging.RetentionDays = 1
		config.Worker.PaymentSuccessRate = 80
	}

	return config
}

