package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/store"
)

type partyNav struct {
	ID   int
	Name string
}

type BaseTemplateData struct {
	CurrentPagePath    string
	Flash              string
	IsAuthenticated    bool
	CurrentYear        int
	CurrentUserParties []partyNav
	FullName           string
	UserEmail          string
}

// TODO: refactor out references to store from here
type MoviesTemplateData struct {
	Movies      []partymgmt.TMDBMovie
	Movie       store.Movie
	Parties     []store.Party
	SearchValue string
	BaseTemplateData
}

type ProfilesTemplateData struct {
	Profile       identityaccess.Profile
	Parties       []store.PartiesForMemberResult
	WatchedMovies []store.WatchedMoviesForMemberResult
	BaseTemplateData
}

type PartiesTemplateData struct {
	Party partymgmt.Party
	BaseTemplateData
}

func (a *Application) NewTemplateData(r *http.Request, path string) BaseTemplateData {
	return newBaseTemplateData(r, path)
}

func (a *Application) NewMoviesTemplateData(r *http.Request, path string) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: newBaseTemplateData(r, path),
	}
}

func (a *Application) NewProfilesTemplateData(r *http.Request, path string) ProfilesTemplateData {
	return ProfilesTemplateData{
		BaseTemplateData: newBaseTemplateData(r, path),
	}
}

func (a *Application) NewPartiesTemplateData(r *http.Request, path string) PartiesTemplateData {
	return PartiesTemplateData{
		BaseTemplateData: newBaseTemplateData(r, path),
	}
}

func newBaseTemplateData(r *http.Request, path string) BaseTemplateData {
	authed := isAuthenticated(r.Context())

	var (
		partiesForNav []partyNav
		fullName      string
		email         string
	)
	if authed {
		partiesForNav = r.Context().Value(partiesForNavContextKey).([]partyNav)
		fullName = r.Context().Value(fullNameContextKey).(string)
		email = r.Context().Value(emailContextKey).(string)
	}

	return BaseTemplateData{
		//			Flash:       a.sessionManager.PopString(r.Context(), "flash"),
		CurrentPagePath:    path,
		CurrentYear:        2024,
		IsAuthenticated:    authed,
		CurrentUserParties: partiesForNav,
		FullName:           fullName,
		UserEmail:          email,
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
	"joinGenres": func(genres []partymgmt.Genre) string {
		res := ""
		for i, g := range genres {
			if i == 0 {
				res = g.Name
			} else {
				res = fmt.Sprintf("%s, %s", res, g.Name)
			}
		}
		return res
	},
	"timeToDuration": func(minutes int) string {
		hours := minutes / 60
		mins := minutes % 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	},
	"formatDate": func(date time.Time) string {
		return date.Format("January 2006")
	},
	"formatWatchDate": func(date time.Time) string {
		return date.Format("January 03, 2006")
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

	return cache, nil
}

func partialDirExist(filesystem fs.FS, dir string) bool {
	_, err := fs.Stat(filesystem, fmt.Sprintf("html/pages/%s/partials", dir))
	return err == nil
}
