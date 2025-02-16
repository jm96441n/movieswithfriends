package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/partymgmt"
)

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
}

type AddMovieToPartiesModalTemplateData struct {
	MovieID         int
	TMDBID          int
	AddedParties    []partymgmt.Party
	NotAddedParties []partymgmt.Party
}

type MoviesTemplateData struct {
	Movies                   []partymgmt.TMDBMovie
	Movie                    partymgmt.Movie
	MovieAddedToCurrentParty bool
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
	Party                 partymgmt.Party
	CurrentWatcherIsOwner bool
	Members               []partymgmt.PartyMember
	ModalData             InviteModalTemplateData
	BaseTemplateData
}

type PartiesIndexTemplateData struct {
	Parties        []partymgmt.Party
	InvitedParties []partymgmt.Party
	CurrentUserID  int
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

func (a *Application) NewPartiesIndexTemplateData(r *http.Request, w http.ResponseWriter, path string, parties, invitedParties []partymgmt.Party, currentUserID int) PartiesIndexTemplateData {
	return PartiesIndexTemplateData{
		Parties:          parties,
		InvitedParties:   invitedParties,
		CurrentUserID:    currentUserID,
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
		fullName string
		email    string
	)

	if authed {
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
		ErrorFlashes:    errorFlashes,
		InfoFlashes:     infoFlashes,
		WarningFlashes:  warningFlashes,
		CurrentPagePath: path,
		CurrentYear:     2025,
		IsAuthenticated: authed,
		FullName:        fullName,
		UserEmail:       email,
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
		"join": strings.Join,
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
		"sanitizeToID": func(input string) string {
			// If empty string, return a default
			if len(strings.TrimSpace(input)) == 0 {
				return "id"
			}

			// Convert to lowercase and trim spaces
			s := strings.ToLower(strings.TrimSpace(input))

			// Replace any whitespace with hyphens
			s = strings.Join(strings.Fields(s), "-")

			// Remove all characters except letters, numbers, hyphens, underscores, periods
			reg := regexp.MustCompile(`[^a-z0-9\-_.]+`)
			s = reg.ReplaceAllString(s, "")

			// Ensure it starts with a letter
			if len(s) > 0 && !unicode.IsLetter(rune(s[0])) {
				s = "id-" + s
			}

			// Handle empty string after sanitization
			if s == "" {
				return "id"
			}

			return s
		},
		"formatBudget": func(budget int) string {
			switch {
			case budget >= 1000000:
				return fmt.Sprintf("$%d million", budget/1000000)
			case budget >= 1000:
				return fmt.Sprintf("$%d thousand", budget/1000)
			default:
				return fmt.Sprintf("$%d", budget)
			}
		},
		"formatFullDate": func(date time.Time) string {
			est, err := time.LoadLocation("America/New_York")
			if err != nil {
				fmt.Printf("Error loading timezone: %v\n", err)
				return ""
			}

			// Convert to EST and format date only
			dateOnly := date.In(est).Format("Jan 02, 2006")
			return dateOnly
		},
		"formatStringDate": func(in string) string {
			utcTime, err := time.Parse("2006-01-02", in)
			if err != nil {
				fmt.Printf("Error parsing time: %v\n", err)
				return ""
			}

			est, err := time.LoadLocation("America/New_York")
			if err != nil {
				fmt.Printf("Error loading timezone: %v\n", err)
				return ""
			}

			// Convert to EST and format date only
			dateOnly := utcTime.In(est).Format("Jan 02, 2006")
			return dateOnly
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
		isPartial := strings.Contains(name, "partials")

		if isPartial {
			patterns = []string{path}
			paths := strings.Split(name, "/")

			if len(paths) > 1 && paths[0] != "partials" {
				pageGroup := paths[0]
				patterns = append(patterns, fmt.Sprintf("%s/%s/partials/*.gohtml", base, pageGroup))
			}
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
