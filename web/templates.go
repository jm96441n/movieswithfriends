package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

type TemplateData struct {
	CurrentYear int
	Flash       string
}

func (a *Application) NewTemplateData(r *http.Request) TemplateData {
	return TemplateData{
		CurrentYear: 2023,
	}
}

func NewTemplateCache(filesystem embed.FS) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(filesystem, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		// parse the base template file into a template set
		ts, err := template.New(name).ParseFS(filesystem, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
