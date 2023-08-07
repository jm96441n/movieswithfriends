package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

type profileFinder interface {
	GetProfile(context.Context, int) (store.Profile, error)
}

func ProfileShowHandler(logger *slog.Logger, db profileFinder, tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)
		logger.Info("Received request for profile show")
		profile, err := db.GetProfile(ctx, 1)
		pageData := map[string]string{"Name": profile.Name}
		err = tmpl.ExecuteTemplate(w, "show.gohtml", pageData)
		if err != nil {
			logger.Error(err.Error())
			fmt.Printf("%#v\n", *tmpl)
			w.WriteHeader(500)
		}
		return
	})
}
