package config

import (
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type App struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
	Env  string `yaml:"env"`
}

type Server struct {
	HTTPAddr    string `yaml:"http_addr"`
	MetricsAddr string `yaml:"metrics_addr"`
}

type Security struct {
	JWTAccessSecret  string `yaml:"jwt_access_secret"`
	JWTRefreshSecret string `yaml:"jwt_refresh_secret"`
	AccessTTLMinute  int    `yaml:"access_ttl_minutes"`
	RefreshTTLMinute int    `yaml:"refresh_ttl_minutes"`
}

type Postgres struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Addr string `yaml:"addr"`
	DB   int    `yaml:"db"`
}

type RabbitMQ struct {
	URL         string `yaml:"url"`
	PaymentQueue string `yaml:"payment_queue"`
	CancelQueue  string `yaml:"cancel_queue"`
}

type Elasticsearch struct {
	URL string `yaml:"url"`
}

type Logging struct {
	Dir           string `yaml:"dir"`
	RetentionDays int    `yaml:"retention_days"`
}

type Booking struct {
	AutoCancelMinutes int `yaml:"auto_cancel_minutes"`
	PageDefaultLimit  int `yaml:"page_default_limit"`
	PageMaxLimit      int `yaml:"page_max_limit"`
}

type Worker struct {
	AutoCancelMinutes      int `yaml:"auto_cancel_minutes"`
	PollerIntervalSeconds  int `yaml:"poller_interval_seconds"`
	PaymentSuccessRate     int `yaml:"payment_success_rate"`
}

type Observability struct {
	MetricsUpdateSeconds int `yaml:"metrics_update_seconds"`
}

type Config struct {
	App           App           `yaml:"app"`
	Server        Server        `yaml:"server"`
	Security      Security      `yaml:"security"`
	Postgres      Postgres      `yaml:"postgres"`
	Redis         Redis         `yaml:"redis"`
	RabbitMQ      RabbitMQ      `yaml:"rabbitmq"`
	Elasticsearch Elasticsearch `yaml:"elasticsearch"`
	Logging       Logging       `yaml:"logging"`
	Booking       Booking       `yaml:"booking"`
	Worker        Worker        `yaml:"worker"`
	Observability Observability `yaml:"observability"`
}

// Load reads config file and sets defaults
func Load(path string) *Config {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(raw))

	var c Config
	if err := yaml.Unmarshal([]byte(expanded), &c); err != nil {
		log.Fatalf("yaml unmarshal: %v", err)
	}

	// Apply default values for any missing configuration
	c.SetDefaults()

	// Validate configuration
	if err := c.Validate(); err != nil {
		log.Fatalf("configuration validation: %v", err)
	}

	return &c
}

// SetDefaults applies default values to configuration
func (c *Config) SetDefaults() {
	// Server defaults
	if c.Server.HTTPAddr == "" {
		c.Server.HTTPAddr = DefaultHTTPAddr
	}
	if c.Server.MetricsAddr == "" {
		c.Server.MetricsAddr = DefaultMetricsAddr
	}

	// Security defaults
	if c.Security.AccessTTLMinute == 0 {
		c.Security.AccessTTLMinute = DefaultAccessTTLMinutes
	}
	if c.Security.RefreshTTLMinute == 0 {
		c.Security.RefreshTTLMinute = DefaultRefreshTTLMinutes
	}

	// Logging defaults
	if c.Logging.Dir == "" {
		c.Logging.Dir = DefaultLogDir
	}
	if c.Logging.RetentionDays == 0 {
		c.Logging.RetentionDays = DefaultLogRetentionDays
	}

	// Booking defaults
	if c.Booking.AutoCancelMinutes == 0 {
		c.Booking.AutoCancelMinutes = DefaultAutoCancelMinutes
	}
	if c.Booking.PageDefaultLimit == 0 {
		c.Booking.PageDefaultLimit = DefaultPageSize
	}
	if c.Booking.PageMaxLimit == 0 {
		c.Booking.PageMaxLimit = DefaultMaxPageSize
	}

	// Worker defaults
	if c.Worker.AutoCancelMinutes == 0 {
		c.Worker.AutoCancelMinutes = c.Booking.AutoCancelMinutes
	}
	if c.Worker.PollerIntervalSeconds == 0 {
		c.Worker.PollerIntervalSeconds = DefaultPollerIntervalSeconds
	}
	if c.Worker.PaymentSuccessRate == 0 {
		c.Worker.PaymentSuccessRate = DefaultPaymentSuccessRate
	}

	// Observability defaults
	if c.Observability.MetricsUpdateSeconds == 0 {
		c.Observability.MetricsUpdateSeconds = DefaultMetricsUpdateSeconds
	}

	// RabbitMQ defaults
	if c.RabbitMQ.PaymentQueue == "" {
		c.RabbitMQ.PaymentQueue = DefaultPaymentQueue
	}
	if c.RabbitMQ.CancelQueue == "" {
		c.RabbitMQ.CancelQueue = DefaultCancelQueue
	}
}

// expandEnvVars expands environment variables in the format ${VAR} or ${VAR:-default}
func expandEnvVars(text string) string {
	// Pattern to match ${VAR} or ${VAR:-default}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Remove ${ and }
		varExpr := match[2 : len(match)-1]
		
		// Check if it has a default value (VAR:-default)
		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			varName := parts[0]
			defaultValue := parts[1]
			
			if value := os.Getenv(varName); value != "" {
				return value
			}
			return defaultValue
		}
		
		// No default value, just return env var or empty string
		return os.Getenv(varExpr)
	})
}
