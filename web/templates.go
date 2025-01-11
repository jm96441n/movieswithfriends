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

const (
	FlashErrorKey   = "error"
	FlashInfoKey    = "info"
	FlashWarningKey = "warn"
)

type BaseTemplateData struct {
	CurrentPagePath    string
	ErrorFlashes       []interface{}
	InfoFlashes        []interface{}
	WarningFlashes     []interface{}
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
	Profile           *identityaccess.Profile
	Parties           []store.PartiesForMemberResult
	WatchedMovies     []store.WatchedMoviesForMemberResult
	NumPages          int
	CurPage           int
	HasEmailError     *bool
	HasPasswordError  *bool
	HasFirstNameError *bool
	HasLastNameError  *bool
	BaseTemplateData
}

func (s *ProfilesTemplateData) InitHasErrorFields() {
	s.HasEmailError = new(bool)
	s.HasPasswordError = new(bool)
	s.HasFirstNameError = new(bool)
	s.HasLastNameError = new(bool)
}

type PartiesTemplateData struct {
	Party   partymgmt.Party
	Parties []store.PartiesForMemberResult
	BaseTemplateData
}

type SignupTemplateData struct {
	HasEmailError     *bool
	HasPasswordError  *bool
	HasFirstNameError *bool
	HasLastNameError  *bool
	BaseTemplateData
}

func (s *SignupTemplateData) InitHasErrorFields() {
	s.HasEmailError = new(bool)
	s.HasPasswordError = new(bool)
	s.HasFirstNameError = new(bool)
	s.HasLastNameError = new(bool)
}

func (a *Application) NewTemplateData(r *http.Request, w http.ResponseWriter, path string) BaseTemplateData {
	return a.newBaseTemplateData(r, w, path)
}

func (a *Application) NewMoviesTemplateData(r *http.Request, w http.ResponseWriter, path string) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: a.newBaseTemplateData(r, w, path),
	}
}

func (a *Application) NewProfilesTemplateData(r *http.Request, w http.ResponseWriter, path string) ProfilesTemplateData {
	return ProfilesTemplateData{
		BaseTemplateData: a.newBaseTemplateData(r, w, path),
	}
}

func (a *Application) NewPartiesTemplateData(r *http.Request, w http.ResponseWriter, path string) PartiesTemplateData {
	return PartiesTemplateData{
		BaseTemplateData: a.newBaseTemplateData(r, w, path),
	}
}

func (a *Application) NewSignupTemplateData(r *http.Request, w http.ResponseWriter, path string) *SignupTemplateData {
	return &SignupTemplateData{
		BaseTemplateData: a.newBaseTemplateData(r, w, path),
	}
}

func (a *Application) newBaseTemplateData(r *http.Request, w http.ResponseWriter, path string) BaseTemplateData {
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

	var (
		errorFlashes   []interface{}
		warningFlashes []interface{}
		infoFlashes    []interface{}
	)
	session, err := a.SessionStore.Get(r, sessionName)
	if err == nil {
		errorFlashes = session.Flashes(FlashErrorKey)
		warningFlashes = session.Flashes(FlashWarningKey)
		infoFlashes = session.Flashes(FlashInfoKey)
		session.Save(r, w)
	}

	return BaseTemplateData{
		ErrorFlashes:       errorFlashes,
		InfoFlashes:        infoFlashes,
		WarningFlashes:     warningFlashes,
		CurrentPagePath:    path,
		CurrentYear:        2024,
		IsAuthenticated:    authed,
		CurrentUserParties: partiesForNav,
		FullName:           fullName,
		UserEmail:          email,
	}
}

var functions = template.FuncMap{
	"navClasses": func(currentPath, targetPath string) string {
		if currentPath == targetPath {
			return "active"
		}
		return ""
	},
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
	"disableIfEmpty": func(s string) string {
		if s == "" {
			return "disabled"
		}
		return ""
	},
	"isInvalidClass": func(invalid *bool) string {
		if invalid == nil {
			return ""
		}
		val := *invalid
		if val {
			return "is-invalid"
		}

		return "is-valid"
	},
	"activeIfCurrentPageForPagination": func(currentPage, targetPage int) string {
		if currentPage == targetPage {
			return "active"
		}
		return ""
	},
	"pageNums": func(curPage, numPages int) []int {
		pages := make([]int, 0, 3)
		// when on the first page show the following 2 pages
		// when on the last page show the previous 2 pages
		// when on a page in the middle show the previous page, current page, and next page
		switch curPage {
		case 1:
			for i := curPage; i <= numPages && i < curPage+3; i++ {
				pages = append(pages, i)
			}
		case numPages:
			for i := max(1, curPage-2); i <= numPages; i++ {
				pages = append(pages, i)
			}
		default:
			for i := curPage - 1; i <= curPage+1; i++ {
				pages = append(pages, i)
			}
		}

		return pages
	},
}

func NewTemplateCache(filesystem embed.FS) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	fs.WalkDir(filesystem, "html/pages", func(path string, d fs.DirEntry, _ error) error {
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
