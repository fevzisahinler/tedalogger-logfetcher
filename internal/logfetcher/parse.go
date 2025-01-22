package logfetcher

import (
	"regexp"
	"strings"
	"time"
)

var (
	ruijieKV = regexp.MustCompile(`(\w+)=([^,]+)`)
	fortiKV  = regexp.MustCompile(`(\w+)=(?:"([^"]+)"|([^",\s]+))`)
)

func parseAndDetermineBrand(lm LogMessage) ParsedLog {
	lowerMsg := strings.ToLower(lm.Message)
	switch {
	case strings.Contains(lowerMsg, "urlfilterlog"):
		return parseRuijie(lm)
	case strings.Contains(lowerMsg, "devname=") || strings.Contains(lowerMsg, "devid="):
		return parseForti(lm)
	default:
		pl := ParsedLog{
			Brand:      "unknown",
			RawMessage: lm.Message,
			FromHost:   lm.FromHost,
		}
		if lm.TimeReported != "" {
			t, err := time.Parse(time.RFC3339, lm.TimeReported)
			if err == nil {
				pl.Timestamp = t
			}
		}
		return pl
	}
}

func parseRuijie(msg LogMessage) ParsedLog {
	pl := ParsedLog{
		Brand:      "ruijie",
		RawMessage: msg.Message,
		FromHost:   msg.FromHost,
	}
	matches := ruijieKV.FindAllStringSubmatch(msg.Message, -1)
	for _, m := range matches {
		key := m[1]
		val := strings.TrimSpace(m[2])
		switch key {
		case "deviceId":
			pl.DeviceID = val
		case "urlCategory":
			pl.URLCategory = val
		case "srcIpv4":
			pl.SrcIP = val
		case "dstIpv4":
			pl.DstIP = val
		case "srcPort":
			pl.SrcPort = val
		case "dstPort":
			pl.DstPort = val
		case "srcUser":
			pl.User = val
		case "srcMac":
			pl.SrcMac = val
		case "policyName":
			pl.PolicyName = val
		case "url":
			pl.URL = val
		case "action":
			if val == "1" {
				pl.Action = "allowed"
			} else {
				pl.Action = "blocked"
			}
		case "timestamp":
			t, err := time.Parse("2006-01-02 15:04:05", val)
			if err == nil {
				pl.Timestamp = t
			}
		}
	}
	if pl.Timestamp.IsZero() && msg.TimeReported != "" {
		t, err := time.Parse(time.RFC3339, msg.TimeReported)
		if err == nil {
			pl.Timestamp = t
		}
	}
	return pl
}

func parseForti(msg LogMessage) ParsedLog {
	pl := ParsedLog{
		Brand:      "forti",
		RawMessage: msg.Message,
		FromHost:   msg.FromHost,
	}
	matches := fortiKV.FindAllStringSubmatch(msg.Message, -1)
	for _, m := range matches {
		key := m[1]
		val := m[2]
		if val == "" {
			val = m[3]
		}
		val = strings.TrimSpace(val)
		switch key {
		case "devname":
			pl.DevName = val
		case "devid":
			pl.DevID = val
		case "srcip":
			pl.SrcIP = val
		case "dstip":
			pl.DstIP = val
		case "srcport":
			pl.SrcPort = val
		case "dstport":
			pl.DstPort = val
		case "srcintf":
			pl.SrcIntf = val
		case "hostname":
			pl.Hostname = val
		case "url":
			pl.URL = val
		case "user":
			pl.User = val
		case "action":
			pl.Action = val
		case "time":
		}
	}
	if pl.Timestamp.IsZero() && msg.TimeReported != "" {
		t, err := time.Parse(time.RFC3339, msg.TimeReported)
		if err == nil {
			pl.Timestamp = t
		}
	}
	return pl
}
