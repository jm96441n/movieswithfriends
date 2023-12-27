package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jm96441n/movieswithfriends/store"
)

type TMDBClient struct {
	client  *retryablehttp.Client
	tmdbKey string
	baseURL string
}

type trailerResults struct {
	Results []struct {
		Key  string `json:"key"`
		Type string `json:"type"`
	} `json:"results"`
}

type SearchResults struct {
	Movies []store.Movie `json:"results"`
	Page   int           `json:"page"`
}

func NewTMDBClient(baseURL, apiKey string) *TMDBClient {
	return &TMDBClient{
		client:  retryablehttp.NewClient(),
		baseURL: baseURL,
		tmdbKey: apiKey,
	}
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

func (t *TMDBClient) GetMovie(ctx context.Context, id int) (store.Movie, error) {
	req, err := t.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/movie/%d", t.baseURL, id))
	if err != nil {
		return store.Movie{}, err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return store.Movie{}, err
	}

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return store.Movie{}, err
	}

	result := store.Movie{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return store.Movie{}, err
	}

	result.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", result.PosterURL)

	req, err = t.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/movie/%d/videos", t.baseURL, id))
	if err != nil {
		return store.Movie{}, err
	}

	res, err = t.client.Do(req)
	if err != nil {
		return store.Movie{}, err
	}

	defer res.Body.Close()
	respBody, err = io.ReadAll(res.Body)
	if err != nil {
		return store.Movie{}, err
	}

	trailers := trailerResults{}
	err = json.Unmarshal(respBody, &trailers)
	if err != nil {
		return store.Movie{}, err
	}

	fmt.Println(trailers.Results)

	for _, trailer := range trailers.Results {
		if trailer.Type == "Trailer" {
			result.TrailerURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", trailer.Key)
			break
		}
	}
	fmt.Println(result)

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
