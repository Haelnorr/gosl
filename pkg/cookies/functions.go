package cookies

import (
	"net/http"
	"time"
)

// Tell the browser to delete the cookie matching the name provided
// Path must match the original set cookie for it to delete
func DeleteCookie(w http.ResponseWriter, name string, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Expires:  time.Unix(0, 0), // Expire in the past
		MaxAge:   -1,              // Immediately expire
		HttpOnly: true,
	})
}

// Set a cookie with the given name, path and value. maxAge directly relates
// to cookie MaxAge (0 for no max age, >0 for TTL in seconds)
func SetCookie(
	w http.ResponseWriter,
	name string,
	path string,
	value string,
	maxAge int,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		MaxAge:   maxAge,
	})
}
