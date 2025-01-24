package web

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
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
	CurrentPagePath string
	ErrorFlashes    []interface{}
	InfoFlashes     []interface{}
	WarningFlashes  []interface{}
	IsAuthenticated bool
	CurrentYear     int
	FullName        string
	UserEmail       string
	Parties         []partymgmt.Party
	CurrentParty    partymgmt.Party
}

type MoviesTemplateData struct {
	Movies                   []partymgmt.TMDBMovie
	CurrentPartyMovieTMDBIDs map[int]struct{}
	Movie                    partymgmt.Movie
	SearchValue              string
	BaseTemplateData
}

type ProfilesTemplateData struct {
	Profile           *identityaccess.Profile
	WatchedMovies     []partymgmt.PartyMovie
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
	Party partymgmt.Party
	BaseTemplateData
}

type SignupTemplateData struct {
	HasEmailError     *bool
	HasPasswordError  *bool
	HasFirstNameError *bool
	HasLastNameError  *bool
	BaseTemplateData
}

type SidebarTemplateData struct {
	Parties      []partymgmt.Party
	CurrentParty partymgmt.Party
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

func (a *Application) NewSidebarTemplateData(r *http.Request, w http.ResponseWriter, currentPartyID int) SidebarTemplateData {
	watcher, err := a.getWatcherFromSession(r)

	if errors.Is(err, ErrFailedToGetProfileIDFromSession) {
		a.Logger.DebugContext(r.Context(), "profileID is not in session")
		return SidebarTemplateData{}
	} else if err != nil {
		a.Logger.Error("failed to get watcher from session", slog.Any("error", err))
	}

	parties, err := watcher.GetParties(r.Context())
	if err != nil {
		// handle later
		a.Logger.Error("failed to get watcher from session", slog.Any("error", err))
	}

	currentParty := parties[0]

	if currentPartyID > 0 {
		res, err := a.PartiesRepository.GetPartyByID(r.Context(), currentPartyID)
		if errors.Is(err, partymgmtstore.ErrNoRecord) {
			a.Logger.Error("party not found", slog.Any("error", err))
		}
		currentParty = partymgmt.Party{
			ID:   res.ID,
			Name: res.Name,
		}
	}
	return SidebarTemplateData{
		Parties:      parties,
		CurrentParty: currentParty,
	}
}

func (a *Application) newBaseTemplateData(r *http.Request, w http.ResponseWriter, path string) BaseTemplateData {
	authed := isAuthenticated(r.Context())

	var (
		fullName       string
		email          string
		currentPartyID int
	)

	if authed {
		fullName = r.Context().Value(fullNameContextKey).(string)
		email = r.Context().Value(emailContextKey).(string)
		if id, ok := r.Context().Value(currentPartyIDContextKey).(int); ok {
			currentPartyID = id
		}
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

	watcher, err := a.getWatcherFromSession(r)

	if errors.Is(err, ErrFailedToGetProfileIDFromSession) {
		return BaseTemplateData{
			ErrorFlashes:    errorFlashes,
			InfoFlashes:     infoFlashes,
			WarningFlashes:  warningFlashes,
			CurrentPagePath: path,
			CurrentYear:     2024,
			IsAuthenticated: authed,
			FullName:        fullName,
			UserEmail:       email,
		}
	} else if err != nil {
		a.Logger.Error("failed to get watcher from session", slog.Any("error", err))
	}

	var currentParty partymgmt.Party

	if currentPartyID > 0 {
		res, err := a.PartiesRepository.GetPartyByID(r.Context(), currentPartyID)
		if errors.Is(err, partymgmtstore.ErrNoRecord) {
			a.Logger.Error("party not found", slog.Any("error", err))
		}
		currentParty.ID = res.ID
		currentParty.Name = res.Name
	}

	parties, err := watcher.GetParties(r.Context())
	if err != nil {
		// handle later
		a.Logger.Error("failed to get watcher from session", slog.Any("error", err))
	}

	return BaseTemplateData{
		ErrorFlashes:    errorFlashes,
		InfoFlashes:     infoFlashes,
		WarningFlashes:  warningFlashes,
		CurrentPagePath: path,
		CurrentYear:     2024,
		IsAuthenticated: authed,
		FullName:        fullName,
		UserEmail:       email,
		CurrentParty:    currentParty,
		Parties:         parties,
	}
}

func (a *Application) templFunctions() template.FuncMap {
	return template.FuncMap{
		"navClasses": func(currentPath, targetPath string) string {
			if currentPath == targetPath {
				return "active"
			}
			return ""
		},
		"hyphenate": func(s string) string {
			s = strings.ReplaceAll(strings.ToLower(s), ":", "")
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
		"showSidebar": func(path string) bool {
			_, ok := nonsidebarPaths[path]
			return !ok
		},
		"assetPath": a.assetPath,
		"movieWatched": func(id int, tmdbIDS map[int]struct{}) bool {
			_, ok := tmdbIDS[id]
			return ok
		},
	}
}

func (a *Application) assetPath(path string) string {
	return a.AssetLoader.Path(path)
}

var nonsidebarPaths = map[string]struct{}{
	"/login":  {},
	"/signup": {},
	"/":       {},
}

func (a *Application) initTemplateCache(filesystem embed.FS) error {
	cache := make(map[string]*template.Template)

	fs.WalkDir(filesystem, "html", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		name, _ := strings.CutPrefix(path, "html/pages/")
		base := "html/pages"
		if strings.Contains(path, "html/partials/") {
			name, _ = strings.CutPrefix(path, "html/")
			base = "html"
		}

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
				patterns = append(patterns, fmt.Sprintf("%s/%s/partials/*.gohtml", base, pageGroup))
			}

			patterns = append(patterns, path)
		}

		// parse the base template file into a template set
		ts, err := template.New(name).Funcs(a.templFunctions()).ParseFS(filesystem, patterns...)
		if err != nil {
			return err
		}

		cache[name] = ts
		return nil
	})

	a.templateCache = cache
	return nil
}

func partialDirExist(filesystem fs.FS, dir string) bool {
	_, err := fs.Stat(filesystem, fmt.Sprintf("html/pages/%s/partials", dir))
	return err == nil
}
