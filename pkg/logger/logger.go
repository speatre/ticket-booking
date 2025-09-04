package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a zap logger that writes JSON logs to stdout and to daily rotated files under dir.
// Files are named <app>-YYYYMMDD.log and a symlink <dir>/app.log points to the latest.
// Old files older than 30 days are removed automatically. Compression is enabled by rotatelogs.
func New(appName string, env string, dir string) *zap.Logger {
	if dir == "" {
		dir = "logs"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("create log dir: %v", err)
	}

	pattern := filepath.Join(dir, appName+"-%Y%m%d.log")
	linkName := filepath.Join(dir, "app.log")

	rotator, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(linkName),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(30*24*time.Hour),
	)
	if err != nil {
		log.Fatalf("create rotatelogs: %v", err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	level := zap.InfoLevel
	if env == "dev" || env == "development" {
		level = zap.DebugLevel
	}

	fileWS := zapcore.AddSync(rotator)
	stdoutWS := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileWS, level),
		zapcore.NewCore(jsonEncoder, stdoutWS, level),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger
}

// NewMetricsLogger creates a logger for metrics operations that writes to a separate metrics.log file
// This keeps metrics logs completely separate from application business logic logs
func NewMetricsLogger(dir string) *zap.Logger {
	if dir == "" {
		dir = "logs"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("create metrics log dir: %v", err)
	}

	pattern := filepath.Join(dir, "metrics-%Y%m%d.log")
	linkName := filepath.Join(dir, "metrics.log")

	rotator, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(linkName),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(7*24*time.Hour), // Keep metrics logs shorter
	)
	if err != nil {
		log.Fatalf("create metrics rotatelogs: %v", err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	// Create a custom core that adds log type
	core := zapcore.NewCore(jsonEncoder, zapcore.AddSync(rotator), zap.WarnLevel)

	// Add log type field to identify metrics logs
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).
		With(zap.String("log_type", "metrics"))

	return logger
}

// NewAccessLogger creates a logger for HTTP access logs (separate from business logic)
func NewAccessLogger(dir string) *zap.Logger {
	if dir == "" {
		dir = "logs"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("create access log dir: %v", err)
	}

	pattern := filepath.Join(dir, "access-%Y%m%d.log")
	linkName := filepath.Join(dir, "access.log")

	rotator, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(linkName),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(30*24*time.Hour),
	)
	if err != nil {
		log.Fatalf("create access rotatelogs: %v", err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)
	core := zapcore.NewCore(jsonEncoder, zapcore.AddSync(rotator), zap.InfoLevel)

	return zap.New(core, zap.AddCaller()).
		With(zap.String("log_type", "access"))
}


