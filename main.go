package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokenURL     string `json:"token_url"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ProgramInfo struct {
	AvailableMarkets []string `json:"available_markets"`
	Copyrights       []any    `json:"copyrights"`
	Description      string   `json:"description"`
	Episodes         struct {
		Href     string `json:"href"`
		Items    []Item `json:"items"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous any    `json:"previous"`
		Total    int    `json:"total"`
	} `json:"episodes"`
	Explicit     bool `json:"explicit"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href            string `json:"href"`
	HTMLDescription string `json:"html_description"`
	ID              string `json:"id"`
	Images          []struct {
		Height int    `json:"height"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
	} `json:"images"`
	IsExternallyHosted bool     `json:"is_externally_hosted"`
	Languages          []string `json:"languages"`
	MediaType          string   `json:"media_type"`
	Name               string   `json:"name"`
	Publisher          string   `json:"publisher"`
	TotalEpisodes      int      `json:"total_episodes"`
	Type               string   `json:"type"`
	URI                string   `json:"uri"`
}

type ProgramInfoNext struct {
	Href     string `json:"href"`
	Items    []Item `json:"items"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
}

type Item struct {
	AudioPreviewURL string `json:"audio_preview_url"`
	Description     string `json:"description"`
	DurationMs      int    `json:"duration_ms"`
	Explicit        bool   `json:"explicit"`
	ExternalUrls    struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href            string `json:"href"`
	HTMLDescription string `json:"html_description"`
	ID              string `json:"id"`
	Images          []struct {
		Height int    `json:"height"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
	} `json:"images"`
	IsExternallyHosted   bool     `json:"is_externally_hosted"`
	IsPlayable           bool     `json:"is_playable"`
	Language             string   `json:"language"`
	Languages            []string `json:"languages"`
	Name                 string   `json:"name"`
	ReleaseDate          string   `json:"release_date"`
	ReleaseDatePrecision string   `json:"release_date_precision"`
	Type                 string   `json:"type"`
	URI                  string   `json:"uri"`
}

func GetAccessToken(config Config) (TokenResponse, error) {
	var tokenResponse TokenResponse

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequest("POST", config.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return tokenResponse, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return tokenResponse, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse, err
	}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return tokenResponse, err
	}

	return tokenResponse, nil
}

func GetProgramData(tokenResponse TokenResponse, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func main() {
	// config読込
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}

	// アクセストークン取得
	tokenResponse, err := GetAccessToken(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// データ取得
	url := "https://api.spotify.com/v1/shows/"
	program := "4zqDMbg9WSpC5l81gJCfEc"
	url += program

	var pi ProgramInfo
	var pin ProgramInfoNext
	var items []Item

	var totalItem int
	var readItem int

	for i := 0; ; i++ {
		body, err := GetProgramData(tokenResponse, url)
		if err != nil {
			log.Fatal(err.Error())
		}

		if i == 0 {
			err = json.Unmarshal(body, &pi)
			if err != nil {
				log.Fatal(err.Error())
			}

			totalItem = pi.TotalEpisodes
			readItem += len(pi.Episodes.Items)

			items = append(items, pi.Episodes.Items...)

			if pi.Episodes.Next != "" {
				url = pi.Episodes.Next
			} else {
				break
			}
		} else {
			err = json.Unmarshal(body, &pin)
			if err != nil {
				log.Fatal(err.Error())
			}

			readItem += len(pin.Items)

			items = append(items, pin.Items...)

			if pin.Next != "" {
				url = pin.Next
			} else {
				break
			}
		}

		if totalItem == readItem {
			break
		}
	}
}
