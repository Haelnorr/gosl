package jwt

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Token interface {
	GetJTI() uuid.UUID
	GetEXP() int64
	GetScope() string
	GetUser(ctx context.Context, tx *db.SafeTX) (*models.User, error)
}

// Access token
type AccessToken struct {
	ISS   string    // Issuer, generally TrustedHost
	IAT   int64     // Time issued at
	EXP   int64     // Time expiring at
	TTL   string    // Time-to-live: "session" or "exp". Used with 'remember me'
	SUB   int       // Subject (user) ID
	JTI   uuid.UUID // UUID-4 used for identifying blacklisted tokens
	Fresh int64     // Time freshness expiring at
	Scope string    // Should be "access"
}

// Refresh token
type RefreshToken struct {
	ISS   string    // Issuer, generally TrustedHost
	IAT   int64     // Time issued at
	EXP   int64     // Time expiring at
	TTL   string    // Time-to-live: "session" or "exp". Used with 'remember me'
	SUB   int       // Subject (user) ID
	JTI   uuid.UUID // UUID-4 used for identifying blacklisted tokens
	Scope string    // Should be "refresh"
}

func (a AccessToken) GetUser(ctx context.Context, tx *db.SafeTX) (*models.User, error) {
	user, err := models.GetUserFromID(ctx, tx, a.SUB)
	if err != nil {
		return nil, errors.Wrap(err, "db.GetUserFromID")
	}
	return user, nil
}
func (r RefreshToken) GetUser(ctx context.Context, tx *db.SafeTX) (*models.User, error) {
	user, err := models.GetUserFromID(ctx, tx, r.SUB)
	if err != nil {
		return nil, errors.Wrap(err, "db.GetUserFromID")
	}
	return user, nil
}

func (a AccessToken) GetJTI() uuid.UUID {
	return a.JTI
}
func (r RefreshToken) GetJTI() uuid.UUID {
	return r.JTI
}
func (a AccessToken) GetEXP() int64 {
	return a.EXP
}
func (r RefreshToken) GetEXP() int64 {
	return r.EXP
}
func (a AccessToken) GetScope() string {
	return a.Scope
}
func (r RefreshToken) GetScope() string {
	return r.Scope
}
