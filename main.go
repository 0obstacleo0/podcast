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

func main() {
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

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequest("POST", config.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	fmt.Printf("Access Token: %s\n", tokenResponse.AccessToken)
	fmt.Printf("Token Type: %s\n", tokenResponse.TokenType)
	fmt.Printf("Expires In: %d seconds\n", tokenResponse.ExpiresIn)
}
