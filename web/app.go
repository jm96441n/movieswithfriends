package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/exp/slog"
)

type Application struct {
	Logger        *slog.Logger
	TemplateCache map[string]*template.Template
}

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	a.Logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data TemplateData) {
	ts, ok := a.TemplateCache[page]

	if !ok {
		a.serverError(w, r, fmt.Errorf("template does not exist for page %q", page))
	}

	buf := bytes.NewBuffer([]byte{})

	// write to the buffer  to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}
