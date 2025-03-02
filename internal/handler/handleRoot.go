package handler

import (
	"gosl/internal/view/page"
	"net/http"
)

func Root() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				ErrorPage(http.StatusNotFound, w, r)
				return
			}
			page.Index().Render(r.Context(), w)
		},
	)
}
