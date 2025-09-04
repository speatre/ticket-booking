package booking

// CreateBookingRequest input for creating a booking
type CreateBookingRequest struct {
	EventID  string `json:"event_id" binding:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity int    `json:"quantity" binding:"required,min=1,max=10" example:"2"`
}

// CreateBookingResponse output after creating a booking
type CreateBookingResponse struct {
	BookingID string `json:"booking_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status    Status `json:"status" example:"PENDING"`
}

// BookingResponse represents a booking record
type BookingResponse struct {
	ID       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	EventID  string `json:"event_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID   string `json:"user_id" example:"42e1d21e-1111-2222-3333-444455556666"`
	Quantity int    `json:"quantity" example:"2"`
	Status   Status `json:"status" example:"CONFIRMED"`
}

// ErrorResponse standard error model
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
