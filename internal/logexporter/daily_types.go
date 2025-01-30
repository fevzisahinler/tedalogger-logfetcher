// internal/logexporter/daily_types.go

package logexporter

import "time"

type DailyNAS struct {
	ID                int       `json:"id"`
	Nasname           string    `json:"nasname"`
	Brand             string    `json:"brand"`
	Syslog5651Enabled bool      `json:"syslog_5651_enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type DailyDestination struct {
	ID         int       `json:"id"`
	Type       string    `json:"type"`
	Username   string    `json:"username,omitempty"`
	Password   string    `json:"password,omitempty"`
	Port       int       `json:"port,omitempty"`
	SSHKeyPath string    `json:"sshKeyPath,omitempty"`
	IpAddress  string    `json:"ipAddress,omitempty"`
	FilePath   string    `json:"filePath,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ExportRequestBody struct {
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	DeviceID    int    `json:"deviceId,omitempty"`
	Destination string `json:"destination,omitempty"`

	FTPServerAddress   string `json:"ftpServerAddress,omitempty"`
	FTPPort            int    `json:"ftpPort,omitempty"`
	FTPUsername        string `json:"ftpUsername,omitempty"`
	FTPPassword        string `json:"ftpPassword,omitempty"`
	FTPDestinationPath string `json:"ftpDestinationPath,omitempty"`

	SFTPServerAddress   string `json:"sftpServerAddress,omitempty"`
	SFTPPort            int    `json:"sftpPort,omitempty"`
	SFTPUseKey          bool   `json:"sftpUseKey,omitempty"`
	SFTPKeyPath         string `json:"sftpKeyPath,omitempty"`
	SFTPUsername        string `json:"sftpUsername,omitempty"`
	SFTPPassword        string `json:"sftpPassword,omitempty"`
	SFTPDestinationPath string `json:"sftpDestinationPath,omitempty"`

	LocalFilePath string `json:"localFilePath,omitempty"`
}
