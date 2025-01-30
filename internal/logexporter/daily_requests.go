// internal/logexporter/daily_requests.go

package logexporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tedalogger-logfetcher/config"
)

func getDailyJobNASList() ([]DailyNAS, error) {
	cfg := config.GetConfig()
	tk, err := getDailyJobToken()
	if err != nil {
		return nil, err
	}

	url := cfg.APIBaseURL + "/nas/get_all"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("NAS request oluşturulamadı: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tk)

	cl := &http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NAS isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("[DailyJob] NAS list 401 geldi, token yenileniyor...")
		resetDailyJobToken()
		if err := dailyJobLogin(); err != nil {
			return nil, err
		}
		return getDailyJobNASList() // Tekrar dene
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NAS isteği HTTP %d - %s", resp.StatusCode, bs)
	}

	type responseType struct {
		Error   bool       `json:"error"`
		Message string     `json:"message"`
		Data    []DailyNAS `json:"data"`
	}
	var r responseType
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("NAS parse hatası: %w", err)
	}
	if r.Error {
		return nil, fmt.Errorf("NAS error: %s", r.Message)
	}
	return r.Data, nil
}

func getDailyJobDestinations() ([]DailyDestination, error) {
	cfg := config.GetConfig()
	tk, err := getDailyJobToken()
	if err != nil {
		return nil, err
	}

	url := cfg.APIBaseURL + "/log-destinations/get-all"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Dest request oluşturulamadı: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tk)

	cl := &http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Dest isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("[DailyJob] Dest list 401, token yenileniyor...")
		resetDailyJobToken()
		if err := dailyJobLogin(); err != nil {
			return nil, err
		}
		return getDailyJobDestinations()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Dest isteği HTTP %d - %s", resp.StatusCode, bs)
	}

	type respType struct {
		Error   bool               `json:"error"`
		Message string             `json:"message"`
		Data    []DailyDestination `json:"data"`
	}
	var r respType
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Dest parse hatası: %w", err)
	}
	if r.Error {
		return nil, fmt.Errorf("Dest error: %s", r.Message)
	}
	return r.Data, nil
}

// doExport gelen NAS ve Destination bilgisine göre uygun export isteğini yapar.
func doExport(nas DailyNAS, dest DailyDestination, dateStr string) error {
	switch strings.ToUpper(dest.Type) {
	case "FTP":
		return doJSONExportRequest(ExportRequestBody{
			StartDate:          dateStr,
			EndDate:            dateStr,
			DeviceID:           nas.ID,
			Destination:        "FTP",
			FTPServerAddress:   dest.IpAddress,
			FTPPort:            dest.Port,
			FTPUsername:        dest.Username,
			FTPPassword:        dest.Password,
			FTPDestinationPath: dest.FilePath,
		})

	case "SFTP":
		if dest.SSHKeyPath != "" && dest.Password == "" {
			return doSFTPKeyExportRequest(nas.ID, dateStr, dateStr, dest)
		} else {
			return doJSONExportRequest(ExportRequestBody{
				StartDate:           dateStr,
				EndDate:             dateStr,
				DeviceID:            nas.ID,
				Destination:         "SFTP",
				SFTPServerAddress:   dest.IpAddress,
				SFTPPort:            dest.Port,
				SFTPUseKey:          false,
				SFTPUsername:        dest.Username,
				SFTPPassword:        dest.Password,
				SFTPDestinationPath: dest.FilePath,
			})
		}

	case "LOCAL":
		return doJSONExportRequest(ExportRequestBody{
			StartDate:     dateStr,
			EndDate:       dateStr,
			DeviceID:      nas.ID,
			Destination:   "Local",
			LocalFilePath: dest.FilePath,
		})

	default:
		return fmt.Errorf("Bilinmeyen tip: %s", dest.Type)
	}
}

func doJSONExportRequest(body ExportRequestBody) error {
	cfg := config.GetConfig()
	tk, err := getDailyJobToken()
	if err != nil {
		return err
	}

	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, cfg.APIBaseURL+"/logs/export", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Request oluşturulamadı: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tk)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("JSON export isteği hata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("[DailyJob] 401 -> token yenileniyor (JSON export)")
		resetDailyJobToken()
		if err := dailyJobLogin(); err != nil {
			return err
		}
		return doJSONExportRequest(body)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bs, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Export HTTP %d - %s", resp.StatusCode, bs)
	}

	log.Println("[DailyJob] Export (JSON) başarıyla tamam.")
	return nil
}

func doSFTPKeyExportRequest(deviceID int, startDate, endDate string, dest DailyDestination) error {
	cfg := config.GetConfig()
	tk, err := getDailyJobToken()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("startDate", startDate)
	_ = writer.WriteField("endDate", endDate)
	_ = writer.WriteField("deviceId", strconv.Itoa(deviceID))
	_ = writer.WriteField("destination", "SFTP")
	_ = writer.WriteField("sftpUseKey", "true")
	_ = writer.WriteField("sftpUsername", dest.Username)
	_ = writer.WriteField("sftpPort", strconv.Itoa(dest.Port))
	_ = writer.WriteField("sftpServerAddress", dest.IpAddress)
	_ = writer.WriteField("sftpDestinationPath", dest.FilePath)

	fw, err := writer.CreateFormFile("sftpKeyPath", filepath.Base(dest.SSHKeyPath))
	if err != nil {
		return fmt.Errorf("FormFile hata: %w", err)
	}
	keyBytes, err := os.ReadFile(dest.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("Anahtar okuma hata: %w", err)
	}
	_, _ = fw.Write(keyBytes)

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, cfg.APIBaseURL+"/logs/export", &buf)
	if err != nil {
		return fmt.Errorf("SFTP key request hata: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tk)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SFTP key export hata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("[DailyJob] 401 -> token yenileniyor (SFTP key)")
		resetDailyJobToken()
		if err := dailyJobLogin(); err != nil {
			return err
		}
		return doSFTPKeyExportRequest(deviceID, startDate, endDate, dest)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bs, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SFTP key export HTTP %d - %s", resp.StatusCode, bs)
	}

	log.Println("[DailyJob] Export (SFTP-Key) başarıyla tamam.")
	return nil
}
