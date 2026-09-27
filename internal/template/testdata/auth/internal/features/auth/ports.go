package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"__MODULE__/internal/features/auth/domain"
	"__MODULE__/internal/features/user"
)

type UserDirectory interface {
	GetByEmail(ctx context.Context, email string) (user.Credentials, error)
	Create(ctx context.Context, p user.NewUser) (user.Profile, error)
	GetProfile(ctx context.Context, id uuid.UUID) (user.Profile, error)
}

type Tokens interface {
	Issue(claims domain.Claims, kind domain.Kind, ttl time.Duration) (string, error)
	Verify(token string, want domain.Kind) (domain.Claims, error)
}
