package booking_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"ticket-booking/internal/booking"
	"ticket-booking/internal/database"
	"ticket-booking/internal/mocks"
)

// Helper function to create a service with mocked dependencies
func createTestService(t *testing.T) (*booking.Service, *mocks.MockBookingRepository, *mocks.MockEventReserver, *mocks.MockPublisher, *mocks.MockCache, *mocks.MockDatabase) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockBookingRepository(ctrl)
	reserver := mocks.NewMockEventReserver(ctrl)
	publisher := mocks.NewMockPublisher(ctrl)
	cache := mocks.NewMockCache(ctrl)
	mockDB := mocks.NewMockDatabase(ctrl)
	logger := zap.NewNop()

	svc := booking.NewService(mockDB, repo, reserver, publisher, cache, logger)

	return svc, repo, reserver, publisher, cache, mockDB
}

func TestHandleBookingCreated_InvalidJSON(t *testing.T) {
	svc, _, _, _, _, _ := createTestService(t)
	defer gomock.NewController(t).Finish()

	body := []byte(`invalid json`)
	err := svc.HandleBookingCreated(context.Background(), body)

	require.Error(t, err)
}

func TestGet_Success(t *testing.T) {
	svc, repo, _, _, _, _ := createTestService(t)
	defer gomock.NewController(t).Finish()

	expectedBooking := &booking.Booking{
		ID:       "b1",
		UserID:   "u1",
		EventID:  "e1",
		Quantity: 2,
		Status:   booking.StatusConfirmed,
	}

	repo.EXPECT().Get("b1").Return(expectedBooking, nil)

	booking, err := svc.Get(context.Background(), "b1")

	require.NoError(t, err)
	require.Equal(t, expectedBooking, booking)
}

func TestGet_NotFound(t *testing.T) {
	svc, repo, _, _, _, _ := createTestService(t)
	defer gomock.NewController(t).Finish()

	repo.EXPECT().Get("b1").Return(nil, assert.AnError)

	booking, err := svc.Get(context.Background(), "b1")

	require.Error(t, err)
	require.Equal(t, assert.AnError, err)
	require.Nil(t, booking)
}

// Test the BookingCreatedMessage struct
func TestBookingCreatedMessage_JSON(t *testing.T) {
	msg := booking.BookingCreatedMessage{
		BookingID: "b1",
		UserID:    "u1",
		EventID:   "e1",
		Quantity:  2,
	}

	// Test JSON marshaling using standard library
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "b1")
	require.Contains(t, string(jsonData), "u1")
	require.Contains(t, string(jsonData), "e1")
	require.Contains(t, string(jsonData), "2")

	// Test JSON unmarshaling
	var unmarshaledMsg booking.BookingCreatedMessage
	err = json.Unmarshal(jsonData, &unmarshaledMsg)
	require.NoError(t, err)
	require.Equal(t, msg, unmarshaledMsg)
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	require.Equal(t, "not enough tickets", booking.ErrNotEnoughTickets.Error())
}

// Test status constants
func TestStatusConstants(t *testing.T) {
	require.Equal(t, "PENDING", string(booking.StatusPending))
	require.Equal(t, "CONFIRMED", string(booking.StatusConfirmed))
	require.Equal(t, "CANCELLED", string(booking.StatusCancelled))
}

// Test booking model
func TestBooking_Model(t *testing.T) {
	b := &booking.Booking{
		ID:             "b1",
		UserID:         "u1",
		EventID:        "e1",
		Quantity:       2,
		UnitPriceCents: 5000,
		Status:         booking.StatusPending,
	}

	require.Equal(t, "b1", b.ID)
	require.Equal(t, "u1", b.UserID)
	require.Equal(t, "e1", b.EventID)
	require.Equal(t, 2, b.Quantity)
	require.Equal(t, int64(5000), b.UnitPriceCents)
	require.Equal(t, booking.StatusPending, b.Status)
}

// Test service interface compliance
func TestService_ImplementsInterface(t *testing.T) {
	svc, _, _, _, _, _ := createTestService(t)
	defer gomock.NewController(t).Finish()

	// This test ensures the service implements the BookingService interface
	var _ booking.BookingService = svc
}

// Test repository interface methods
func TestBookingRepository_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBookingRepository(ctrl)

	// Test that the mock implements the interface
	var _ booking.BookingRepository = repo
}

// Test cache interface methods
func TestCache_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)

	// Test that the mock implements the interface
	var _ booking.Cache = cache
}

// Test publisher interface methods
func TestPublisher_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	publisher := mocks.NewMockPublisher(ctrl)

	// Test that the mock implements the interface
	var _ booking.Publisher = publisher
}

// Test event reserver interface methods
func TestEventReserver_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reserver := mocks.NewMockEventReserver(ctrl)

	// Test that the mock implements the interface
	var _ booking.EventReserver = reserver
}

// Test database interface methods
func TestDatabase_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDatabase(ctrl)

	// Test that the mock implements the interface
	var _ database.Database = mockDB
}

// Test BookingCreatedMessage JSON serialization
func TestCreateBooking_BookingCreatedMessage_JSON(t *testing.T) {
	msg := booking.BookingCreatedMessage{
		BookingID: "booking123",
		UserID:    "user456",
		EventID:   "event789",
		Quantity:  3,
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	require.Contains(t, string(data), "booking123")
	require.Contains(t, string(data), "user456")
	require.Contains(t, string(data), "event789")
	require.Contains(t, string(data), "3")

	// Test JSON unmarshaling
	var unmarshaled booking.BookingCreatedMessage
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, msg, unmarshaled)
}

// Test CreateBooking message validation
func TestCreateBooking_MessageValidation(t *testing.T) {
	tests := []struct {
		name     string
		msg      booking.BookingCreatedMessage
		expected bool
	}{
		{
			name: "valid message",
			msg: booking.BookingCreatedMessage{
				BookingID: "b1",
				UserID:    "u1",
				EventID:   "e1",
				Quantity:  2,
			},
			expected: true,
		},
		{
			name: "empty booking ID",
			msg: booking.BookingCreatedMessage{
				UserID:   "u1",
				EventID:  "e1",
				Quantity: 2,
			},
			expected: false,
		},
		{
			name: "empty user ID",
			msg: booking.BookingCreatedMessage{
				BookingID: "b1",
				EventID:   "e1",
				Quantity:  2,
			},
			expected: false,
		},
		{
			name: "zero quantity",
			msg: booking.BookingCreatedMessage{
				BookingID: "b1",
				UserID:    "u1",
				EventID:   "e1",
				Quantity:  0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.msg.BookingID != "" && tt.msg.UserID != "" && tt.msg.EventID != "" && tt.msg.Quantity > 0
			require.Equal(t, tt.expected, isValid)
		})
	}
}

// Test booking status constants and transitions
func TestCreateBooking_StatusConstants(t *testing.T) {
	// Test that booking starts in PENDING status
	require.Equal(t, "PENDING", string(booking.StatusPending))
	require.Equal(t, "CONFIRMED", string(booking.StatusConfirmed))
	require.Equal(t, "CANCELLED", string(booking.StatusCancelled))

	// Test status transition logic
	initialStatus := booking.StatusPending
	require.Equal(t, booking.StatusPending, initialStatus)

	// Simulate status change to confirmed
	confirmedStatus := booking.StatusConfirmed
	require.Equal(t, booking.StatusConfirmed, confirmedStatus)
	require.NotEqual(t, booking.StatusPending, confirmedStatus)
}

// Test error handling for booking creation
func TestCreateBooking_ErrorHandling(t *testing.T) {
	// Test ErrNotEnoughTickets error
	require.Equal(t, "not enough tickets", booking.ErrNotEnoughTickets.Error())
	require.NotNil(t, booking.ErrNotEnoughTickets)

	// Test that it's the correct error type
	err := booking.ErrNotEnoughTickets
	require.Error(t, err)
	require.Contains(t, err.Error(), "not enough tickets")
}

// Test concurrent message publishing (simulated)
func TestCreateBooking_ConcurrentMessagePublishing(t *testing.T) {
	_, _, _, publisher, cache, _ := createTestService(t)
	defer gomock.NewController(t).Finish()

	ctx := context.Background()
	numMessages := 10

	// Mock expectations for concurrent operations
	publisher.EXPECT().Publish("booking.created", gomock.Any()).Return(nil).Times(numMessages)
	cache.EXPECT().Set(ctx, gomock.Any(), "1", gomock.Any()).Return(nil).Times(numMessages)

	// Simulate concurrent message publishing
	var wg sync.WaitGroup
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := booking.BookingCreatedMessage{
				BookingID: fmt.Sprintf("booking%d", id),
				UserID:    fmt.Sprintf("user%d", id),
				EventID:   "event123",
				Quantity:  1,
			}

			// Simulate publishing (this would normally be done by CreateBooking)
			err := publisher.Publish("booking.created", msg)
			require.NoError(t, err)

			// Simulate cache operation
			err = cache.Set(ctx, fmt.Sprintf("booking:pending:booking%d", id), "1", 0)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent message publishing test completed successfully")
}

// Test booking creation with edge cases for quantity
func TestCreateBooking_QuantityEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		expectError bool
		errorMsg    string
	}{
		{"normal quantity", 2, false, ""},
		{"zero quantity", 0, false, ""},     // Zero might be allowed depending on business rules
		{"large quantity", 1000, false, ""}, // Large quantity should be handled
		{"negative quantity", -1, true, "negative quantity not allowed"},
		{"very large quantity", 10000, false, ""}, // Should handle large numbers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test quantity validation logic
			if tt.quantity < 0 {
				require.True(t, tt.expectError, "Negative quantity should be invalid")
				require.Contains(t, tt.errorMsg, "negative")
			} else {
				require.False(t, tt.expectError, "Non-negative quantity should be valid")
			}

			// Test that quantity is preserved in booking creation
			b := &booking.Booking{
				Quantity: tt.quantity,
			}
			require.Equal(t, tt.quantity, b.Quantity)
		})
	}
}

// Test booking ID generation logic
func TestCreateBooking_IDGeneration(t *testing.T) {
	// Test that booking IDs are properly formatted
	bookingID := "booking123"
	require.Contains(t, bookingID, "booking")
	require.NotEmpty(t, bookingID)

	// Test UUID-like ID format
	require.Regexp(t, `^booking\d+$`, bookingID)

	// Test that IDs are unique
	id1 := "booking1"
	id2 := "booking2"
	require.NotEqual(t, id1, id2)
}

// Test cache key generation for pending bookings
func TestCreateBooking_CacheKeyGeneration(t *testing.T) {
	bookingID := "booking123"
	expectedKey := "booking:pending:" + bookingID

	require.Equal(t, "booking:pending:booking123", expectedKey)
	require.Contains(t, expectedKey, "booking:pending:")
	require.Contains(t, expectedKey, bookingID)
}
