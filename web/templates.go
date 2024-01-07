package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jm96441n/movieswithfriends/store"
)

type BaseTemplateData struct {
	CurrentPagePath string
	Flash           string
	CurrentYear     int
}

type MoviesTemplateData struct {
	Movies      []store.Movie
	Movie       store.Movie
	Parties     []store.Party
	SearchValue string
	BaseTemplateData
}

type ProfilesTemplateData struct {
	Profile store.Profile
	BaseTemplateData
}

type PartiesTemplateData struct {
	Party         store.Party
	Movies        []store.Movie
	SelectedMovie store.Movie
	BaseTemplateData
}

func (a *Application) NewTemplateData(r *http.Request, path string) BaseTemplateData {
	return BaseTemplateData{
		//		Flash:       a.sessionManager.PopString(r.Context(), "flash"),
		CurrentPagePath: path,
		CurrentYear:     2023,
	}
}

func (a *Application) NewMoviesTemplateData(r *http.Request, path string) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: BaseTemplateData{
			//			Flash:       a.sessionManager.PopString(r.Context(), "flash"),
			CurrentPagePath: path,
			CurrentYear:     2023,
		},
	}
}

func (a *Application) NewProfilesTemplateData(r *http.Request, path string) ProfilesTemplateData {
	return ProfilesTemplateData{
		BaseTemplateData: BaseTemplateData{
			//			Flash:       a.sessionManager.PopString(r.Context(), "flash"),
			CurrentPagePath: path,
			CurrentYear:     2023,
		},
	}
}

func (a *Application) NewPartiesTemplateData(r *http.Request, path string) PartiesTemplateData {
	return PartiesTemplateData{
		BaseTemplateData: BaseTemplateData{
			//			Flash:       a.sessionManager.PopString(r.Context(), "flash"),
			CurrentPagePath: path,
			CurrentYear:     2023,
		},
	}
}

func navClasses(currentPath, targetPath string) string {
	if currentPath == targetPath {
		return "active"
	}
	return ""
}

var functions = template.FuncMap{
	"navClasses": navClasses,
	"hyphenate": func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), " ", "-")
	},
	"disableClassForTrailerButton": func(s string) string {
		if s == "" {
			return "disabled"
		}
		return ""
	},
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
			}

			paths := strings.Split(name, "/")
			if len(paths) > 1 && partialDirExist(filesystem, paths[0]) {
				pageGroup := paths[0]
				patterns = append(patterns, fmt.Sprintf("html/pages/%s/partials/*.gohtml", pageGroup))
			}

			patterns = append(patterns, path)
		}

		// parse the base template file into a template set
		ts, err := template.New(name).Funcs(functions).ParseFS(filesystem, patterns...)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cache[name] = ts
		return nil
	})

	fmt.Println(cache)
	return cache, nil
}

func partialDirExist(filesystem fs.FS, dir string) bool {
	_, err := fs.Stat(filesystem, fmt.Sprintf("html/pages/%s/partials", dir))
	return err == nil
}
