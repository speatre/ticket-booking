package database

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// Database defines the database operations needed by the booking service
type Database interface {
	WithContext(ctx context.Context) *gorm.DB
	Transaction(fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

// gormDBAdapter adapts *gorm.DB to the Database interface
type gormDBAdapter struct {
	db *gorm.DB
}

// NewDatabaseAdapter creates a new database adapter
func NewDatabaseAdapter(db *gorm.DB) Database {
	return &gormDBAdapter{db: db}
}

func (g *gormDBAdapter) WithContext(ctx context.Context) *gorm.DB {
	return g.db.WithContext(ctx)
}

func (g *gormDBAdapter) Transaction(fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return g.db.Transaction(fn, opts...)
}
