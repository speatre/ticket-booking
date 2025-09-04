package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"ticket-booking/internal/user"
)

// MockRepository is a mock for the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockRepository) ByEmail(email string) (*user.User, error) {
	args := m.Called(email)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockRepository) ByID(id string) (*user.User, error) {
	args := m.Called(id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockRepository) Update(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func TestService_Register(t *testing.T) {
	logger := zap.NewNop() // No-op logger for tests
	mockRepo := new(MockRepository)
	svc := user.NewService(mockRepo, logger)

	tests := []struct {
		name        string
		email       string
		password    string
		mockSetup   func()
		expectedID  string
		expectedErr error
	}{
		{
			name:     "Successful registration",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				matcher := mock.MatchedBy(func(u *user.User) bool {
					if u.Email != "test@example.com" || u.PasswordHash == "" {
						return false
					}
					return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")) == nil
				})
				mockRepo.On("Create", matcher).Return(nil).Run(func(args mock.Arguments) {
					u := args.Get(0).(*user.User)
					u.ID = "user-1"
				}).Once()
			},
			expectedID:  "",
			expectedErr: nil,
		},
		{
			name:     "Repository error",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("Create", mock.Anything).Return(assert.AnError).Once()
			},
			expectedID:  "",
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			id, err := svc.Register(context.Background(), tt.email, tt.password)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				assert.NotEmpty(t, id)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_VerifyLogin(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRepository)
	svc := user.NewService(mockRepo, logger)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		email        string
		password     string
		mockSetup    func()
		expectedUser *user.User
		expectedErr  error
	}{
		{
			name:     "Successful login",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("ByEmail", "test@example.com").Return(&user.User{PasswordHash: string(hashed)}, nil).Once()
			},
			expectedUser: &user.User{PasswordHash: string(hashed)},
			expectedErr:  nil,
		},
		{
			name:     "Invalid credentials",
			email:    "test@example.com",
			password: "wrongpass",
			mockSetup: func() {
				mockRepo.On("ByEmail", "test@example.com").Return(&user.User{PasswordHash: string(hashed)}, nil).Once()
			},
			expectedUser: nil,
			expectedErr:  user.ErrInvalidCredentials,
		},
		{
			name:     "User not found",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("ByEmail", "test@example.com").Return((*user.User)(nil), assert.AnError).Once()
			},
			expectedUser: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			u, err := svc.VerifyLogin(context.Background(), tt.email, tt.password)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedUser, u)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateProfile(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRepository)
	svc := user.NewService(mockRepo, logger)

	fullName := "New Name"

	tests := []struct {
		name        string
		callerID    string
		targetID    string
		fullName    *string
		mockSetup   func()
		expectedErr error
	}{
		{
			name:     "Successful update",
			callerID: "123",
			targetID: "123",
			fullName: &fullName,
			mockSetup: func() {
				mockRepo.On("ByID", "123").Return(&user.User{ID: "123"}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(u *user.User) bool {
					return u.ID == "123" && *u.FullName == "New Name"
				})).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			name:        "Forbidden",
			callerID:    "123",
			targetID:    "456",
			fullName:    &fullName,
			mockSetup:   func() {},
			expectedErr: user.ErrForbidden,
		},
		{
			name:     "User not found",
			callerID: "123",
			targetID: "123",
			fullName: &fullName,
			mockSetup: func() {
				mockRepo.On("ByID", "123").Return((*user.User)(nil), assert.AnError).Once()
			},
			expectedErr: assert.AnError,
		},
		{
			name:     "Update error",
			callerID: "123",
			targetID: "123",
			fullName: &fullName,
			mockSetup: func() {
				mockRepo.On("ByID", "123").Return(&user.User{ID: "123"}, nil).Once()
				mockRepo.On("Update", mock.Anything).Return(assert.AnError).Once()
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := svc.UpdateProfile(context.Background(), tt.callerID, tt.targetID, tt.fullName)
			assert.Equal(t, tt.expectedErr, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
