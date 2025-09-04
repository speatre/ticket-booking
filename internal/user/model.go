// Package user provides user management functionality including
// authentication, authorization, and profile management.
package user

import "time"

// User represents a system user with authentication and role-based access control.
type User struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"` // Unique user identifier
	Email        string    `gorm:"uniqueIndex;not null"`                            // Unique email for authentication
	PasswordHash string    `gorm:"not null"`                                        // Bcrypt hashed password
	Role         string    `gorm:"type:text;not null;default:'USER'"`               // User role: 'USER' or 'ADMIN'
	FullName     *string   // Optional display name
	CreatedAt    time.Time // Account creation timestamp
	UpdatedAt    time.Time // Last profile update timestamp
}
