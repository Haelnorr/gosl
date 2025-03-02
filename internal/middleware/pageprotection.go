package middleware

import (
	"gosl/internal/handler"
	"gosl/pkg/contexts"
	"net/http"
)

// Checks if the user is set in the context and shows 401 page if not logged in
func LoginReq(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contexts.GetUser(r.Context())
		if user == nil {
			handler.ErrorPage(http.StatusUnauthorized, w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Checks if the user is set in the context and redirects them to profile if
// they are logged in
func LogoutReq(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contexts.GetUser(r.Context())
		if user != nil {
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
