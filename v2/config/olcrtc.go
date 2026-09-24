package config

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type OLCRTCOptions struct {
	Type                 string `json:"type"`
	Name                 string `json:"name,omitempty"`
	Provider             string `json:"provider"`
	Transport            string `json:"transport"`
	RoomID               string `json:"room_id"`
	RoomPassword         string `json:"room_password,omitempty"`
	ClientID             string `json:"client_id"`
	AuthToken            string `json:"auth_token,omitempty"`
	KeyHex               string `json:"key_hex"`
	DNSServer            string `json:"dns_server,omitempty"`
	VP8FPS               int    `json:"vp8_fps,omitempty"`
	VP8BatchSize         int    `json:"vp8_batch,omitempty"`
	KeepaliveIntervalSec int    `json:"keepalive_interval_sec,omitempty"`
	SocksPort            int    `json:"socks_port,omitempty"`
}

var ActiveOLCRTCOptions *OLCRTCOptions

func ParseOLCRTCURI(uriStr string) (*OLCRTCOptions, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return nil, fmt.Errorf("invalid olcrtc uri: %w", err)
	}
	if u.Scheme != "olcrtc" && u.Scheme != "olconnect" {
		return nil, fmt.Errorf("not an olcrtc/olconnect uri")
	}

	q := u.Query()
	opts := &OLCRTCOptions{
		Type:                 "olcrtc",
		DNSServer:            "77.88.8.8:53",
		VP8FPS:               60,
		VP8BatchSize:         8,
		KeepaliveIntervalSec: 15,
		SocksPort:            10808,
	}

	roomID := strings.Trim(u.Path, "/")
	if roomID == "" || roomID == "room" {
		roomID = u.Host
	}
	if roomID == "room" {
		roomID = strings.Trim(u.Path, "/")
	}
	opts.RoomID = roomID

	if u.User != nil && u.User.Username() != "" {
		opts.Provider = u.User.Username()
	} else if p := q.Get("provider"); p != "" {
		opts.Provider = p
	} else {
		opts.Provider = "telemost"
	}

	if opts.Provider == "wb_stream" {
		opts.Provider = "wbstream"
	}

	opts.KeyHex = q.Get("key")
	if opts.KeyHex == "" {
		opts.KeyHex = q.Get("k")
	}

	opts.ClientID = q.Get("client_id")
	if opts.ClientID == "" {
		opts.ClientID = q.Get("c")
	}

	opts.Transport = q.Get("transport")
	if opts.Transport == "" {
		opts.Transport = q.Get("t")
	}
	if opts.Transport == "" {
		opts.Transport = "datachannel"
	}

	opts.RoomPassword = q.Get("room_password")
	if opts.RoomPassword == "" {
		opts.RoomPassword = q.Get("rp")
	}

	opts.AuthToken = q.Get("auth_token")
	if opts.AuthToken == "" {
		opts.AuthToken = q.Get("auth.token")
	}
	if opts.AuthToken == "" {
		opts.AuthToken = q.Get("a")
	}

	if dns := q.Get("dns"); dns != "" {
		opts.DNSServer = dns
	} else if dns := q.Get("d"); dns != "" {
		opts.DNSServer = dns
	}

	if fps := q.Get("vp8_fps"); fps != "" {
		if v, err := strconv.Atoi(fps); err == nil && v > 0 {
			opts.VP8FPS = v
		}
	}

	if batch := q.Get("vp8_batch"); batch != "" {
		if v, err := strconv.Atoi(batch); err == nil && v > 0 {
			opts.VP8BatchSize = v
		}
	}

	if ka := q.Get("keepalive"); ka != "" {
		if v, err := strconv.Atoi(ka); err == nil && v > 0 {
			opts.KeepaliveIntervalSec = v
		}
	}

	if u.Fragment != "" {
		opts.Name = u.Fragment
	} else {
		opts.Name = fmt.Sprintf("olcRTC %s", opts.Provider)
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return opts, nil
}

func (opts *OLCRTCOptions) Validate() error {
	if opts.RoomID == "" {
		return fmt.Errorf("olcRTC: room_id is required")
	}
	if opts.KeyHex == "" {
		return fmt.Errorf("olcRTC: key_hex is required")
	}
	if len(opts.KeyHex) != 64 {
		return fmt.Errorf("olcRTC: key_hex must be 64 hex characters")
	}
	if _, err := hex.DecodeString(opts.KeyHex); err != nil {
		return fmt.Errorf("olcRTC: key_hex is not valid hex: %w", err)
	}
	return nil
}

func (opts *OLCRTCOptions) ToSingboxJSON() []byte {
	outboundTag := opts.Name
	if outboundTag == "" {
		outboundTag = "olcRTC"
	}
	port := opts.SocksPort
	if port <= 0 {
		port = 10808
	}

	socksOutbound := map[string]interface{}{
		"type":        "socks",
		"tag":         outboundTag,
		"server":      "127.0.0.1",
		"server_port": port,
		"version":     "5",
	}

	res := map[string]interface{}{
		"outbounds": []interface{}{socksOutbound},
	}
	b, _ := json.Marshal(res)
	return b
}

func isOLCRTCUri(s string) bool {
	return strings.HasPrefix(s, "olcrtc://") || strings.HasPrefix(s, "olconnect://")
}

func DetectOLCRTCOptions(content string, path string) *OLCRTCOptions {
	trimmed := strings.TrimSpace(content)
	if isOLCRTCUri(trimmed) {
		if opts, err := ParseOLCRTCURI(trimmed); err == nil {
			ActiveOLCRTCOptions = opts
			return opts
		}
	}

	if strings.Contains(content, "_olcrtc") {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(content), &raw); err == nil {
			if olcData, ok := raw["_olcrtc"]; ok {
				if b, err := json.Marshal(olcData); err == nil {
					var opts OLCRTCOptions
					if err := json.Unmarshal(b, &opts); err == nil && opts.RoomID != "" {
						ActiveOLCRTCOptions = &opts
						return &opts
					}
				}
			}
		}
	}

	if path != "" {
		sidecar := path + ".olcrtc"
		if data, err := os.ReadFile(sidecar); err == nil {
			var opts OLCRTCOptions
			if err := json.Unmarshal(data, &opts); err == nil && opts.RoomID != "" {
				ActiveOLCRTCOptions = &opts
				return &opts
			}
		}

		if data, err := os.ReadFile(path); err == nil {
			trimmedFile := strings.TrimSpace(string(data))
			if isOLCRTCUri(trimmedFile) {
				if opts, err := ParseOLCRTCURI(trimmedFile); err == nil {
					ActiveOLCRTCOptions = opts
					return opts
				}
			}
			if strings.Contains(string(data), "_olcrtc") {
				var raw map[string]interface{}
				if err := json.Unmarshal(data, &raw); err == nil {
					if olcData, ok := raw["_olcrtc"]; ok {
						if b, err := json.Marshal(olcData); err == nil {
							var opts OLCRTCOptions
							if err := json.Unmarshal(b, &opts); err == nil && opts.RoomID != "" {
								ActiveOLCRTCOptions = &opts
								return &opts
							}
						}
					}
				}
			}
		}
	}

	ActiveOLCRTCOptions = nil
	return nil
}

func SaveOLCRTCOptions(path string, opts *OLCRTCOptions) error {
	if path == "" || opts == nil {
		return nil
	}
	b, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path+".olcrtc", b, 0o644)
}

