// internal/logfetcher/api_client.go

package logfetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"tedalogger-logfetcher/config"
)

var (
	tokenMutex   sync.RWMutex
	currentToken string
)

func loginOnce() error {
	cfg := config.GetConfig()
	baseURL := cfg.APIBaseURL
	if baseURL == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}

	endpoint := fmt.Sprintf("%s/auth/login", baseURL)
	reqBody := map[string]string{
		"username": cfg.APIUsername,
		"password": cfg.APIPassword,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("login request creation error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed, status: %d, body: %s", resp.StatusCode, string(b))
	}

	var parsed APIAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if parsed.Error {
		return fmt.Errorf("login error from API: %s", parsed.Message)
	}

	tokenMutex.Lock()
	currentToken = parsed.Data.Token
	tokenMutex.Unlock()

	log.Printf("API login successful. Token acquired.")
	return nil
}

func GetToken() (string, error) {
	tokenMutex.RLock()
	tkn := currentToken
	tokenMutex.RUnlock()

	if tkn == "" {
		if err := loginOnce(); err != nil {
			return "", err
		}
		tokenMutex.RLock()
		tkn = currentToken
		tokenMutex.RUnlock()
	}
	return tkn, nil
}

func fetchNASList() ([]NAS, error) {
	tkn, err := GetToken()
	if err != nil {
		return nil, fmt.Errorf("cannot get token: %w", err)
	}

	cfg := config.GetConfig()
	baseURL := cfg.APIBaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("API_BASE_URL not set")
	}

	endpoint := fmt.Sprintf("%s/nas/get_all", baseURL)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tkn)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("Got 401 Unauthorized. Attempting to re-login...")

		tokenMutex.Lock()
		currentToken = ""
		tokenMutex.Unlock()

		if err := loginOnce(); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		return fetchNASList()
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetchNASList failed, status: %d, body: %s", resp.StatusCode, string(b))
	}

	var parsed NASListResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if parsed.Error {
		return nil, fmt.Errorf("API error: %s", parsed.Message)
	}

	return parsed.Data, nil
}
