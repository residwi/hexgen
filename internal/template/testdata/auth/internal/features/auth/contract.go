package auth

import (
	"time"

	"__MODULE__/internal/features/user"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	User         user.Profile
}
