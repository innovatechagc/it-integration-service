package domain

import "context"

// UserRepository define las operaciones de persistencia para usuarios
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// AuditRepository define las operaciones de persistencia para auditoría
type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
}

// HealthRepository define las operaciones para health checks
type HealthRepository interface {
	CheckDatabase(ctx context.Context) error
	CheckExternalServices(ctx context.Context) map[string]error
}