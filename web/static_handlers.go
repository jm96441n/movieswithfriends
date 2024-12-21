package web

import (
	"net/http"
)

func (a *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/")
	a.render(w, r, http.StatusOK, "home.gohtml", data)
}
