package handler

import (
	"gosl/internal/view/page"
	"net/http"
)

func ErrorPage(
	errorCode int,
	w http.ResponseWriter,
	r *http.Request,
) {
	message := map[int]string{
		401: "You need to login to view this page.",
		403: "You do not have permission to view this page.",
		404: "The page or resource you have requested does not exist.",
		500: `An error occured on the server. Please try again, and if this
        continues to happen contact an administrator.`,
		503: "The server is currently down for maintenance and should be back soon. =)",
	}
	w.WriteHeader(errorCode)
	page.Error(errorCode, http.StatusText(errorCode), message[errorCode]).
		Render(r.Context(), w)
}
