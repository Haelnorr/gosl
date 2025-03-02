package contexts

import (
	"context"
	"gosl/internal/models"
)

type AuthenticatedUser struct {
	*models.User
	Fresh int64
}

// Return a new context with the user added in
func SetUser(ctx context.Context, u *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, contextKeyAuthorizedUser, u)
}

// Retrieve a user from the given context. Returns nil if not set
func GetUser(ctx context.Context) *AuthenticatedUser {
	user, ok := ctx.Value(contextKeyAuthorizedUser).(*AuthenticatedUser)
	if !ok {
		return nil
	}
	return user
}
