package web

import (
	"net/http"
)

func (a *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		data := a.NewTemplateData(r, w, "/")
		a.render(w, r, http.StatusNotFound, "404.gohtml", data)
		return
	}
	data := a.NewTemplateData(r, w, "/")
	a.render(w, r, http.StatusOK, "home.gohtml", data)
}
