// internal/logexporter/daily_auth.go

package logexporter

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
	dailyJobToken   string
	dailyJobTokenMu sync.RWMutex
)

func dailyJobLogin() error {
	cfg := config.GetConfig()
	body := map[string]string{
		"username": cfg.APIUsername,
		"password": cfg.APIPassword,
	}

	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, cfg.APIBaseURL+"/auth/login", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("login request hata: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login isteği hata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bs, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login HTTP %d - %s", resp.StatusCode, bs)
	}

	type loginResp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	var lr loginResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return fmt.Errorf("login parse hata: %w", err)
	}
	if lr.Error {
		return fmt.Errorf("login api error: %s", lr.Message)
	}

	dailyJobTokenMu.Lock()
	dailyJobToken = lr.Data.Token
	dailyJobTokenMu.Unlock()

	log.Println("[DailyJob] Login başarılı, token alındı.")
	return nil
}

func getDailyJobToken() (string, error) {
	dailyJobTokenMu.RLock()
	tk := dailyJobToken
	dailyJobTokenMu.RUnlock()

	if tk == "" {
		if err := dailyJobLogin(); err != nil {
			return "", err
		}
		dailyJobTokenMu.RLock()
		tk = dailyJobToken
		dailyJobTokenMu.RUnlock()
	}
	return tk, nil
}

func resetDailyJobToken() {
	dailyJobTokenMu.Lock()
	dailyJobToken = ""
	dailyJobTokenMu.Unlock()
}
