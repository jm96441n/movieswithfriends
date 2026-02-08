# Web Rendering Utilities

This document describes the rendering utilities available for HTTP handlers in domain packages.

## Overview

The `web.Renderer` interface provides template rendering and flash message capabilities that can be used by handlers in domain packages (`identityaccess/http/` and `partymgmt/http/`).

## Renderer Interface

```go
type Renderer interface {
    Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
    RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
    SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
    SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
    ServerError(w http.ResponseWriter, r *http.Request, err error)
    ClientError(w http.ResponseWriter, r *http.Request, status int, message string)
}
```

## Usage in Domain Handlers

Domain handlers should receive a `web.Renderer` reference to access rendering capabilities:

```go
package http

import (
    "net/http"
    "github.com/jm96441n/movieswithfriends/web"
    "github.com/jm96441n/movieswithfriends/partymgmt"
)

type PartyHandler struct {
    renderer       web.Renderer
    partyService   partymgmt.PartyService
}

func NewPartyHandler(renderer web.Renderer, partyService partymgmt.PartyService) *PartyHandler {
    return &PartyHandler{
        renderer:     renderer,
        partyService: partyService,
    }
}

func (h *PartyHandler) Show(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    party, err := h.partyService.GetParty(ctx, partyID)
    if err != nil {
        h.renderer.ServerError(w, r, err)
        return
    }
    
    // Render full page template
    data := PartyTemplateData{Party: party}
    h.renderer.Render(w, r, http.StatusOK, "parties/show.gohtml", data)
}

func (h *PartyHandler) AddMovie(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    err := h.partyService.AddMovie(ctx, partyID, movieID)
    if err != nil {
        h.renderer.SetErrorFlashMessage(w, r, "Failed to add movie")
        http.Redirect(w, r, "/parties", http.StatusSeeOther)
        return
    }
    
    h.renderer.SetInfoFlashMessage(w, r, "Movie added successfully!")
    http.Redirect(w, r, fmt.Sprintf("/parties/%d", partyID), http.StatusSeeOther)
}

func (h *PartyHandler) MovieListPartial(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    movies, err := h.partyService.GetMovies(ctx, partyID)
    if err != nil {
        h.renderer.ServerError(w, r, err)
        return
    }
    
    // Render partial template (for HTMX responses)
    data := MoviesPartialData{Movies: movies}
    h.renderer.RenderPartial(w, r, http.StatusOK, "partials/movies/list.gohtml", data)
}
```

## Method Descriptions

### Render
Renders a full page template with the base layout. Use this for complete page responses.

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `status`: HTTP status code (e.g., `http.StatusOK`)
- `page`: Template path relative to `ui/html/pages/` (e.g., `"parties/show.gohtml"`)
- `data`: Template data struct

### RenderPartial
Renders a partial template without the base layout. Use this for HTMX responses or modal content.

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `status`: HTTP status code
- `page`: Template path (e.g., `"partials/movies/list.gohtml"`)
- `data`: Template data struct

### SetInfoFlashMessage
Sets an informational flash message that will be displayed on the next page load.

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `msg`: Message to display

### SetErrorFlashMessage
Sets an error flash message that will be displayed on the next page load.

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `msg`: Error message to display

### ServerError
Logs the error and returns a 500 Internal Server Error response. Use this for unexpected server errors.

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `err`: Error to log

### ClientError
Sets an error flash message and returns the specified HTTP status code. Use this for client errors (400, 404, etc.).

**Parameters:**
- `w`: HTTP response writer
- `r`: HTTP request
- `status`: HTTP status code
- `message`: Error message to display

## Integration with web.Application

The `web.Application` struct implements the `Renderer` interface, so it can be passed directly to domain handlers:

```go
// In cmd/movieswithfriends/main.go or web/routes.go
app := web.NewApplication(cfg)

// Pass app as Renderer to domain handlers
partyHandler := partymgmthttp.NewPartyHandler(app, app.PartyService)
```

## Template Data

Domain handlers should work with template data structures defined in `web/templates.go`. Domain-specific template data can extend the base template data:

```go
type PartyTemplateData struct {
    Party partymgmt.Party
    web.BaseTemplateData
}
```

## Flash Message Keys

Flash messages are categorized by key:
- `FlashInfoKey` - Informational messages (success, confirmation)
- `FlashErrorKey` - Error messages
- `FlashWarningKey` - Warning messages

These are automatically extracted and passed to templates via `BaseTemplateData`.
