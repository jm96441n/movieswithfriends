package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/ui"
)

// Renderer provides template rendering capabilities for HTTP handlers.
// This interface allows domain handlers to render templates without
// depending on the full Application struct.
type Renderer struct {
	templateCache map[string]*template.Template
	sessionStore  *SessionAccessor
	assetLoader   *Loader
	logger        *slog.Logger
}

func NewRenderer(logger *slog.Logger, sessionStore *SessionAccessor, loader *Loader) *Renderer {
	r := Renderer{
		sessionStore: sessionStore,
		assetLoader:  loader,
		logger:       logger,
	}

	r.initTemplateCache(ui.TemplateFS)
	return &r
}

// Render renders a full page template with the base layout.
func (rr *Renderer) Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := rr.templateCache[page]

	if !ok {
		rr.ServerError(w, r, fmt.Errorf("template does not exist for page %q", page))
		return
	}

	buf := bytes.NewBuffer([]byte{})

	// write to the buffer to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		rr.ServerError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// RenderPartial renders a partial template without the base layout.
// This is typically used for HTMX responses.
func (rr *Renderer) RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := rr.templateCache[page]

	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.RenderPartial")
	defer span.End()

	if !ok {
		rr.logger.ErrorContext(ctx, "template does not exist for page", slog.Any("page", page))
		rr.ServerError(w, r, fmt.Errorf("template does not exist for page %q", page))
		return
	}

	tmplName := filepath.Base(page)

	buf := bytes.NewBuffer([]byte{})
	// write to the buffer to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		rr.ServerError(w, r, err)
		return
	}

	rr.logger.InfoContext(ctx, "rendering template", slog.Any("template", page))

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// ServerError logs the error and returns a 500 Internal Server Error response.
func (rr *Renderer) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	ctx, span, _ := metrics.SpanFromContext(r.Context(), "serverError")
	defer span.End()

	rr.logger.ErrorContext(ctx, err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ClientError sets an error flash message and returns the specified HTTP status code.
// TODO: fix flash message and have it re-render the page that we're on
func (rr *Renderer) ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	http.Error(w, http.StatusText(status), status)
}

func (rr *Renderer) initTemplateCache(filesystem embed.FS) error {
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
		ts, err := template.New(name).Funcs(rr.templFunctions()).ParseFS(filesystem, patterns...)
		if err != nil {
			return err
		}

		cache[name] = ts
		return nil
	})

	rr.templateCache = cache
	return nil
}

func partialDirExist(filesystem fs.FS, dir string) bool {
	_, err := fs.Stat(filesystem, fmt.Sprintf("html/pages/%s/partials", dir))
	return err == nil
}

var nonsidebarPaths = map[string]struct{}{
	"/login":  {},
	"/signup": {},
	"/":       {},
}

func (rr *Renderer) templFunctions() template.FuncMap {
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
		"assetPath": rr.assetPath,
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

func (rr *Renderer) assetPath(path string) string {
	return rr.assetLoader.Path(path)
}
