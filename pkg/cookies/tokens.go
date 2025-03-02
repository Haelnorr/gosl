package cookies

import (
	"net/http"
	"time"

	"gosl/internal/models"
	"gosl/pkg/config"
	"gosl/pkg/jwt"

	"github.com/pkg/errors"
)

// Get the value of the access and refresh tokens
func GetTokenStrings(
	r *http.Request,
) (acc string, ref string) {
	accCookie, accErr := r.Cookie("access")
	refCookie, refErr := r.Cookie("refresh")
	var (
		accStr string = ""
		refStr string = ""
	)
	if accErr == nil {
		accStr = accCookie.Value
	}
	if refErr == nil {
		refStr = refCookie.Value
	}
	return accStr, refStr
}

// Set a token with the provided details
func setToken(
	w http.ResponseWriter,
	config *config.Config,
	token string,
	scope string,
	exp int64,
	rememberme bool,
) {
	tokenCookie := &http.Cookie{
		Name:     scope,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   config.SSL,
	}
	if rememberme {
		tokenCookie.Expires = time.Unix(exp, 0)
	}
	http.SetCookie(w, tokenCookie)
}

// Generate new tokens for the user and set them as cookies
func SetTokenCookies(
	w http.ResponseWriter,
	r *http.Request,
	config *config.Config,
	user *models.User,
	fresh bool,
	rememberMe bool,
) error {
	at, atexp, err := jwt.GenerateAccessToken(config, user, fresh, rememberMe)
	if err != nil {
		return errors.Wrap(err, "jwt.GenerateAccessToken")
	}
	rt, rtexp, err := jwt.GenerateRefreshToken(config, user, rememberMe)
	if err != nil {
		return errors.Wrap(err, "jwt.GenerateRefreshToken")
	}
	// Don't set the cookies until we know no errors occured
	setToken(w, config, at, "access", atexp, rememberMe)
	setToken(w, config, rt, "refresh", rtexp, rememberMe)
	return nil
}
