package user

// RegisterRequest represents input for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"secret123"`
}

// RegisterResponse represents successful registration response
type RegisterResponse struct {
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// LoginRequest represents input for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"secret123"`
}

// LoginResponse represents login success with token
type LoginResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"dGhpc19pc19hX3NhbXBsZV9yZWZyZXNoX3Rva2Vu"`
}

// UpdateProfileRequest represents input for updating profile
type UpdateProfileRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,min=2,max=100" example:"John Doe"`
}

// UpdateProfileResponse represents success response for profile update
type UpdateProfileResponse struct {
	OK bool `json:"ok" example:"true"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

// RefreshTokenRequest represents input for refreshing JWT
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"dGhpc19pc19hX3NhbXBsZV9yZWZyZXNoX3Rva2Vu"`
}
