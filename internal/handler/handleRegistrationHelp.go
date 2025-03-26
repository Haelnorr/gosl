package handler

import (
	"gosl/internal/view/page"
	"net/http"
)

func RegistrationHelp() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			page.RegistrationHelp().Render(r.Context(), w)
		},
	)
}
