package partymgmt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type TMDBClient struct {
	client     *retryablehttp.Client
	tmdbKey    string
	baseURL    string
	genreCache genreCache
}

type genreCache struct {
	genres map[int]Genre
}

type trailerResults struct {
	Results []struct {
		Key  string `json:"key"`
		Type string `json:"type"`
	} `json:"results"`
}

type SearchResults struct {
	Movies []TMDBMovie `json:"results"`
	Page   int         `json:"page"`
}

type TMDBMovie struct {
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	Tagline     string `json:"tagline"`
	PosterURL   string `json:"poster_path"`
	TrailerURL  string `json:"trailer_url"`
	URL         string
	ID          int
	Runtime     int     `json:"runtime"`
	Rating      float64 `json:"vote_average"`
	Genres      []Genre `json:"genres"`
	GenreIDs    []int   `json:"genre_ids"`
	TMDBID      int     `json:"id"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewTMDBClient(baseURL, apiKey string, logger *slog.Logger) (*TMDBClient, error) {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = logger
	client := &TMDBClient{
		client:     httpClient,
		baseURL:    baseURL,
		tmdbKey:    apiKey,
		genreCache: genreCache{genres: make(map[int]Genre)},
	}
	err := client.fillCache()
	if err != nil {
		return nil, errors.New("failed to fill cache")
	}
	return client, nil
}

func (t *TMDBClient) GetGenre(genreID int) (Genre, error) {
	if genre, ok := t.genreCache.genres[genreID]; ok {
		return genre, nil
	}
	return Genre{}, fmt.Errorf("genre not found")
}

type GenreList struct {
	Genres []Genre `json:"genres"`
}

func (t *TMDBClient) fillCache() error {
	req, err := t.newRequest(context.Background(), http.MethodGet, fmt.Sprintf("%s/genre/movie/list?language=en", t.baseURL))
	if err != nil {
		return err
	}
	res, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	genres := GenreList{}
	err = json.Unmarshal(respBody, &genres)
	if err != nil {
		return err
	}
	for _, genre := range genres.Genres {
		t.genreCache.genres[genre.ID] = genre
	}
	return nil
}

func (t *TMDBClient) Search(ctx context.Context, term string, page int) (SearchResults, error) {
	term = url.QueryEscape(term)
	req, err := t.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/search/movie?query=%s&page=%d", t.baseURL, term, page))
	if err != nil {
		return SearchResults{}, err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return SearchResults{}, err
	}

	defer res.Body.Close()

	// TODO: use a limited reader here
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return SearchResults{}, err
	}

	result := SearchResults{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return SearchResults{}, err
	}
	return result, nil
}

func (t *TMDBClient) GetMovie(ctx context.Context, id int) (*TMDBMovie, error) {
	req, err := t.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/movie/%d", t.baseURL, id))
	if err != nil {
		return nil, err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	result := &TMDBMovie{}
	err = json.Unmarshal(respBody, result)
	if err != nil {
		return nil, err
	}

	if result.PosterURL != "" {
		result.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", result.PosterURL)
	} else {
		result.PosterURL = "https://placehold.co/270x400?text=No+Poster+Available"
	}

	req, err = t.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/movie/%d/videos", t.baseURL, id))
	if err != nil {
		return nil, err
	}

	res, err = t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	respBody, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	trailers := trailerResults{}
	err = json.Unmarshal(respBody, &trailers)
	if err != nil {
		return nil, err
	}

	for _, trailer := range trailers.Results {
		if trailer.Type == "Trailer" {
			result.TrailerURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", trailer.Key)
			break
		}
	}

	return result, nil
}

func (t *TMDBClient) newRequest(ctx context.Context, method, url string) (*retryablehttp.Request, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+t.tmdbKey)
	return req, nil
}

func (t *TMDBMovie) ToStoreMovie() store.CreateMovieParams {
	genres := make([]string, len(t.Genres))
	for i, genre := range t.Genres {
		genres[i] = genre.Name
	}
	return store.CreateMovieParams{
		Title:       t.Title,
		ReleaseDate: t.ReleaseDate,
		Overview:    t.Overview,
		Tagline:     t.Tagline,
		PosterURL:   t.PosterURL,
		TMDBID:      t.TMDBID,
		TrailerURL:  t.TrailerURL,
		Runtime:     t.Runtime,
		Rating:      t.Rating,
		Genres:      genres,
	}
}
