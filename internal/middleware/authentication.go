package middleware

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"gosl/internal/handler"
	"gosl/internal/models"
	"gosl/pkg/config"
	"gosl/pkg/contexts"
	"gosl/pkg/cookies"
	"gosl/pkg/db"
	"gosl/pkg/jwt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Attempt to use a valid refresh token to generate a new token pair
func refreshAuthTokens(
	config *config.Config,
	ctx context.Context,
	tx *db.SafeWTX,
	w http.ResponseWriter,
	req *http.Request,
	ref *jwt.RefreshToken,
) (*models.User, error) {
	user, err := ref.GetUser(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "ref.GetUser")
	}

	rememberMe := map[string]bool{
		"session": false,
		"exp":     true,
	}[ref.TTL]

	// Set fresh to true because new tokens coming from refresh request
	err = cookies.SetTokenCookies(w, req, config, user, false, rememberMe)
	if err != nil {
		return nil, errors.Wrap(err, "cookies.SetTokenCookies")
	}
	// New tokens sent, revoke the used refresh token
	err = jwt.RevokeToken(ctx, tx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "jwt.RevokeToken")
	}
	// Return the authorized user
	return user, nil
}

// Check the cookies for token strings and attempt to authenticate them
func getAuthenticatedUser(
	config *config.Config,
	ctx context.Context,
	tx *db.SafeWTX,
	w http.ResponseWriter,
	r *http.Request,
) (*contexts.AuthenticatedUser, error) {
	// Get token strings from cookies
	atStr, rtStr := cookies.GetTokenStrings(r)
	// Attempt to parse the access token
	aT, err := jwt.ParseAccessToken(config, ctx, tx, atStr)
	if err != nil {
		// Access token invalid, attempt to parse refresh token
		rT, err := jwt.ParseRefreshToken(config, ctx, tx, rtStr)
		if err != nil {
			return nil, errors.Wrap(err, "jwt.ParseRefreshToken")
		}
		// Refresh token valid, attempt to get a new token pair
		user, err := refreshAuthTokens(config, ctx, tx, w, r, rT)
		if err != nil {
			return nil, errors.Wrap(err, "refreshAuthTokens")
		}
		// New token pair sent, return the authorized user
		authUser := contexts.AuthenticatedUser{
			User:  user,
			Fresh: time.Now().Unix(),
		}
		return &authUser, nil
	}
	// Access token valid
	user, err := aT.GetUser(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "aT.GetUser")
	}
	authUser := contexts.AuthenticatedUser{
		User:  user,
		Fresh: aT.Fresh,
	}
	return &authUser, nil
}

// Attempt to authenticate the user and add their account details
// to the request context
func Authentication(
	logger *zerolog.Logger,
	config *config.Config,
	conn *db.SafeConn,
	next http.Handler,
	maint *uint32,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/output.css" ||
			r.URL.Path == "/static/favicon.ico" {
			next.ServeHTTP(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if atomic.LoadUint32(maint) == 1 {
			cancel()
		}

		// Start the transaction
		tx, err := conn.Begin(ctx, "Authentication middleware")
		if err != nil {
			// Failed to start transaction, skip auth
			logger.Warn().Err(err).
				Msg("Skipping Auth - unable to start a transaction")
			handler.ErrorPage(http.StatusServiceUnavailable, w, r)
			return
		}
		user, err := getAuthenticatedUser(config, r.Context(), tx, w, r)
		if err != nil {
			tx.Rollback()
			// User auth failed, delete the cookies to avoid repeat requests
			cookies.DeleteCookie(w, "access", "/")
			cookies.DeleteCookie(w, "refresh", "/")
			logger.Debug().
				Str("remote_addr", r.RemoteAddr).
				Err(err).
				Msg("Failed to authenticate user")
			next.ServeHTTP(w, r)
			return
		}
		tx.Commit()
		uctx := contexts.SetUser(r.Context(), user)
		newReq := r.WithContext(uctx)
		next.ServeHTTP(w, newReq)
	})
}
