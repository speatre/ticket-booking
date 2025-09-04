package user

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(r Repository, logger *zap.Logger) *Service {
	return &Service{repo: r, logger: logger}
}

func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	hash, err := hashPassword(password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.String("email", email), zap.Error(err))
		return "", err
	}

	u := &User{Email: email, PasswordHash: hash}
	if err := s.repo.Create(u); err != nil {
		s.logger.Error("Failed to create user", zap.String("email", email), zap.Error(err))
		return "", err
	}

	s.logger.Info("User registered successfully", zap.String("user_id", u.ID), zap.String("email", email))
	return u.ID, nil
}

func (s *Service) VerifyLogin(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.ByEmail(email)
	if err != nil {
		s.logger.Error("Failed to find user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Invalid login attempt", zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	s.logger.Info("User logged in successfully", zap.String("user_id", u.ID), zap.String("email", email))
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, callerID, targetID string, fullName *string) error {
	// Ownership check
	if callerID != targetID {
		s.logger.Warn("Unauthorized profile update attempt", zap.String("caller_id", callerID), zap.String("target_id", targetID))
		return ErrForbidden
	}

	u, err := s.repo.ByID(targetID)
	if err != nil {
		s.logger.Error("Failed to find user by ID", zap.String("user_id", targetID), zap.Error(err))
		return err
	}

	u.FullName = fullName
	if err := s.repo.Update(u); err != nil {
		s.logger.Error("Failed to update user profile", zap.String("user_id", targetID), zap.Error(err))
		return err
	}

	s.logger.Info("User profile updated successfully", zap.String("user_id", targetID))
	return nil
}

var (
	ErrInvalidCredentials = Err("invalid credentials")
	ErrForbidden          = Err("forbidden")
)

type Err string

func (e Err) Error() string { return string(e) }
