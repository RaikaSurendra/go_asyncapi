package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LozClient struct {
	baseUrl    string
	httpClient HttpClient
}

func NewLozClient(httpClient HttpClient) *LozClient {
	return &LozClient{
		baseUrl:    "http://botw-compendium.herokuapp.com/api/v3/compendium",
		httpClient: httpClient,
	}
}

type Monster struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Location    []string `json:"common_locations"`
	Drops       []string `json:"drops"`
	Category    string   `json:"category"`
	Image       string   `json:"image"`
	Dlc         bool     `json:"dlc"`
}

type MonsterResponse struct {
	Monsters []Monster `json:"data"`
}

func (c *LozClient) GetMonsters() ([]Monster, error) {
	url := fmt.Sprintf("%s/category/monsters", c.baseUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monsters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result MonsterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Monsters, nil
}
