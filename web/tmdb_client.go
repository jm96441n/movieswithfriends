package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type TMDBClient struct {
	client  *retryablehttp.Client
	tmdbKey string
	baseURL string
}

type SearchResults struct {
	Movies []Movie `json:"results"`
	Page   int     `json:"page"`
}

type Movie struct {
	Title       string `json:"original_title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	PosterURL   string `json:"poster_path"`
	URL         string
	TMDBID      int `json:"id"`
}

func NewTMDBClient(baseURL, apiKey string) *TMDBClient {
	return &TMDBClient{
		client:  retryablehttp.NewClient(),
		baseURL: baseURL,
		tmdbKey: apiKey,
	}
}

func (t *TMDBClient) Search(term string, page int) (SearchResults, error) {
	req, err := t.newRequest(http.MethodGet, fmt.Sprintf("%s/search/movie?query=%s&page=%d", t.baseURL, term, page))
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

func (t *TMDBClient) GetMovie(id int) (Movie, error) {
	req, err := t.newRequest(http.MethodGet, fmt.Sprintf("%s/movie/%d", t.baseURL, id))
	if err != nil {
		return Movie{}, err
	}
	res, err := t.client.Do(req)
	if err != nil {
		return Movie{}, err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return Movie{}, err
	}
	result := Movie{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return Movie{}, err
	}
	return result, nil
}

func (t *TMDBClient) newRequest(method, url string) (*retryablehttp.Request, error) {
	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+t.tmdbKey)
	return req, nil
}
