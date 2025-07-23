package usecase

import (
	"context"
	"fmt"

	"github.com/company/microservice-template/internal/domain"
	"github.com/company/microservice-template/pkg/logger"
)

type UserUseCase interface {
	GetUser(ctx context.Context, id string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error)
}

type userUseCase struct {
	userRepo  domain.UserRepository
	auditRepo domain.AuditRepository
	logger    logger.Logger
}

func NewUserUseCase(
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
	logger logger.Logger,
) UserUseCase {
	return &userUseCase{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

func (u *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user", "user_id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Log audit event
	u.logAuditEvent(ctx, "", "USER_READ", "user", map[string]interface{}{
		"target_user_id": id,
	})

	return user, nil
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	// Validaciones de negocio
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Verificar si el usuario ya existe
	existingUser, _ := u.userRepo.GetByEmail(ctx, user.Email)
	if existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	err := u.userRepo.Create(ctx, user)
	if err != nil {
		u.logger.Error("Failed to create user", "email", user.Email, "error", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Log audit event
	u.logAuditEvent(ctx, "", "USER_CREATE", "user", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})

	u.logger.Info("User created successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	// Verificar que el usuario existe
	existingUser, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		u.logger.Error("Failed to update user", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log audit event
	u.logAuditEvent(ctx, "", "USER_UPDATE", "user", map[string]interface{}{
		"user_id":    user.ID,
		"old_email":  existingUser.Email,
		"new_email":  user.Email,
	})

	u.logger.Info("User updated successfully", "user_id", user.ID)
	return nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, id string) error {
	// Verificar que el usuario existe
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = u.userRepo.Delete(ctx, id)
	if err != nil {
		u.logger.Error("Failed to delete user", "user_id", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log audit event
	u.logAuditEvent(ctx, "", "USER_DELETE", "user", map[string]interface{}{
		"user_id": id,
		"email":   user.Email,
	})

	u.logger.Info("User deleted successfully", "user_id", id)
	return nil
}

func (u *userUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	users, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		u.logger.Error("Failed to list users", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Log audit event
	u.logAuditEvent(ctx, "", "USER_LIST", "user", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})

	return users, nil
}

func (u *userUseCase) logAuditEvent(ctx context.Context, userID, action, resource string, details map[string]interface{}) {
	auditLog := &domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Resource: resource,
		Details:  details,
	}

	// No bloquear la operación principal si falla el audit log
	go func() {
		if err := u.auditRepo.Create(context.Background(), auditLog); err != nil {
			u.logger.Error("Failed to create audit log", "error", err)
		}
	}()
}