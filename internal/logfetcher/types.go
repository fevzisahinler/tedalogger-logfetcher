package logfetcher

import "time"

type LogMessage struct {
	Message       string `json:"message"`
	FromHost      string `json:"fromhost"`
	Facility      string `json:"facility"`
	Priority      string `json:"priority"`
	TimeReported  string `json:"timereported"`
	TimeGenerated string `json:"timegenerated"`
}

type ParsedLog struct {
	Brand      string    `json:"brand,omitempty"`
	SrcIP      string    `json:"src_ip,omitempty"`
	DstIP      string    `json:"dst_ip,omitempty"`
	SrcPort    string    `json:"src_port,omitempty"`
	DstPort    string    `json:"dst_port,omitempty"`
	URL        string    `json:"url,omitempty"`
	Action     string    `json:"action,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	RawMessage string    `json:"raw_message,omitempty"`
	FromHost   string    `json:"from_host,omitempty"`

	DeviceID    string `json:"device_id,omitempty"`
	URLCategory string `json:"url_category,omitempty"`
	SrcMac      string `json:"src_mac,omitempty"`
	PolicyName  string `json:"policy_name,omitempty"`

	User string `json:"user,omitempty"`

	DevID    string `json:"dev_id,omitempty"`
	DevName  string `json:"dev_name,omitempty"`
	SrcIntf  string `json:"src_intf,omitempty"`
	Hostname string `json:"hostname,omitempty"`

	NASName string `json:"nas_name,omitempty"`
}

type NAS struct {
	ID          int    `json:"id"`
	Nasname     string `json:"nasname"`
	Shortname   string `json:"shortname"`
	Type        string `json:"type"`
	Brand       string `json:"brand"`
	Port        int    `json:"port"`
	Secret      string `json:"secret"`
	Server      string `json:"server"`
	Community   string `json:"community"`
	Description string `json:"description"`
}

type APIAuthResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type NASListResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    []NAS  `json:"data"`
}
