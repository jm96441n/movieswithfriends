package http

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

type profileFinder interface {
	GetProfile(context.Context, int) store.Profile
}

func ProfileShowHandler(logger *slog.Logger, tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request for profile show")
		pageData := map[string]string{"Name": "John"}
		err := tmpl.ExecuteTemplate(w, "show.gohtml", pageData)
		if err != nil {
			logger.Error(err.Error())
			fmt.Printf("%#v\n", *tmpl)
			w.WriteHeader(500)
		}
		return
	})
}
