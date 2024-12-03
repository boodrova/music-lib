package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SwaggerClient struct {
	BaseURL string
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func NewSwaggerClient(baseURL string) *SwaggerClient {
	return &SwaggerClient{BaseURL: baseURL}
}

func (c *SwaggerClient) FetchSongDetails(group, song string) (*SongDetail, error) {
	url := fmt.Sprintf("%s/info?group=%s&song=%s", c.BaseURL, group, song)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch song details: %s", resp.Status)
	}

	var details SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}
