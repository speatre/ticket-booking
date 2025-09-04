package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"ticket-booking/internal/event"
	"ticket-booking/internal/mocks"
)

func TestListEvents_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	logger := zap.NewNop()

	// Use nil for database since we're only testing cache operations
	svc := event.NewService(nil, repo, cache, logger)

	// Mock cache hit
	cache.EXPECT().Get(gomock.Any(), "events:list").Return(`[{"id":"e1","name":"Concert"}]`, nil)

	evts, err := svc.List(context.Background())

	require.NoError(t, err)
	require.Len(t, evts, 1)
	require.Equal(t, "Concert", evts[0].Name)
}

func TestListEvents_CacheMiss_FallbackToRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	logger := zap.NewNop()

	// Use nil for database since we're only testing repo operations
	svc := event.NewService(nil, repo, cache, logger)

	// Mock cache miss
	cache.EXPECT().Get(gomock.Any(), "events:list").Return("", assert.AnError)

	// Mock repo response
	events := []event.Event{
		{ID: "e1", Name: "Concert", Remaining: 100},
		{ID: "e2", Name: "Theater", Remaining: 50},
	}
	repo.EXPECT().List().Return(events, nil)

	// Mock cache set
	cache.EXPECT().Set(gomock.Any(), "events:list", gomock.Any(), gomock.Any()).Return(nil)

	evts, err := svc.List(context.Background())

	require.NoError(t, err)
	require.Len(t, evts, 2)
	require.Equal(t, "Concert", evts[0].Name)
	require.Equal(t, "Theater", evts[1].Name)
}

func TestReserve_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	logger := zap.NewNop()

	svc := event.NewService(nil, repo, cache, logger)

	cache.EXPECT().DecrementSeats(gomock.Any(), "e1", 2).Return(8, nil)

	ok, err := svc.Reserve(context.Background(), "e1", 2)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestReserve_NotEnoughSeats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	logger := zap.NewNop()

	svc := event.NewService(nil, repo, cache, logger)

	// Mock cache returning negative seats (not enough)
	cache.EXPECT().DecrementSeats(gomock.Any(), "e1", 10).Return(-2, nil)
	// Mock rollback call
	cache.EXPECT().DecrementSeats(gomock.Any(), "e1", -10).Return(8, nil)

	ok, err := svc.Reserve(context.Background(), "e1", 10)

	require.NoError(t, err)
	require.False(t, ok)
}

// Test interface compliance
func TestService_ImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	logger := zap.NewNop()

	svc := event.NewService(nil, repo, cache, logger)

	// This test ensures the service implements the ServiceInterface
	var _ event.ServiceInterface = svc
}

// Test repository interface methods
func TestEventRepository_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	// Test that the mock implements the interface
	var _ event.EventRepository = repo
}

// Test event model
func TestEvent_Model(t *testing.T) {
	e := &event.Event{
		ID:               "e1",
		Name:             "Concert",
		Remaining:        100,
		TicketPriceCents: 5000,
	}

	require.Equal(t, "e1", e.ID)
	require.Equal(t, "Concert", e.Name)
	require.Equal(t, 100, e.Remaining)
	require.Equal(t, int64(5000), e.TicketPriceCents)
}
