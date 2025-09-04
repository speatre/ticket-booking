package user

import "time"

type User struct {
	ID           string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"type:text;not null;default:'USER'"`
	FullName     *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
