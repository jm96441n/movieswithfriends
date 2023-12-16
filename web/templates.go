package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jm96441n/movieswithfriends/store"
)

type BaseTemplateData struct {
	CurrentYear int
	Flash       string
}

type MoviesTemplateData struct {
	BaseTemplateData
	Movies []store.Movie
	Movie  store.Movie
}

func (a *Application) NewTemplateData(r *http.Request) BaseTemplateData {
	return BaseTemplateData{
		//		Flash:       a.sessionManager.PopString(r.Context(), "flash"),
		CurrentYear: 2023,
	}
}

func (a *Application) NewMoviesTemplateData(r *http.Request) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: BaseTemplateData{
			//			Flash:       a.sessionManager.PopString(r.Context(), "flash"),
			CurrentYear: 2023,
		},
	}
}

func NewTemplateCache(filesystem embed.FS) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	fs.WalkDir(filesystem, "html/pages", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		name, _ := strings.CutPrefix(path, "html/pages/")

		var patterns []string
		if strings.Contains(name, "partials") {
			patterns = []string{path}
		} else {
			patterns = []string{
				"html/base.gohtml",
				"html/partials/*.gohtml",
				path,
			}
		}

		// parse the base template file into a template set
		ts, err := template.New(name).ParseFS(filesystem, patterns...)
		if err != nil {
			return err
		}

		cache[name] = ts
		return nil
	})

	return cache, nil
}
