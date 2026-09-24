package config

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
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

type OLCRTCStore struct {
	Active  string                    `json:"active,omitempty"`
	Options map[string]*OLCRTCOptions `json:"options"`
}

var (
	ActiveOLCRTCOptions     *OLCRTCOptions
	ActiveOLCRTCStore       *OLCRTCStore
	RegisteredOLCRTCOptions = make(map[string]*OLCRTCOptions)
	registryMu              sync.RWMutex
)

func RegisterOLCRTCOption(tag string, opt *OLCRTCOptions) {
	registryMu.Lock()
	defer registryMu.Unlock()
	RegisteredOLCRTCOptions[tag] = opt
	RegisteredOLCRTCOptions[strings.TrimSpace(strings.Split(tag, "§")[0])] = opt
}

func GetOLCRTCOption(tag string) *OLCRTCOptions {
	registryMu.RLock()
	defer registryMu.RUnlock()
	cleanTag := strings.TrimSpace(strings.Split(tag, "§")[0])
	if opt, ok := RegisteredOLCRTCOptions[tag]; ok {
		return opt
	}
	if opt, ok := RegisteredOLCRTCOptions[cleanTag]; ok {
		return opt
	}
	if ActiveOLCRTCStore != nil {
		if opt, ok := ActiveOLCRTCStore.Options[tag]; ok {
			return opt
		}
		if opt, ok := ActiveOLCRTCStore.Options[cleanTag]; ok {
			return opt
		}
	}
	if ActiveOLCRTCOptions != nil {
		if ActiveOLCRTCOptions.Name == tag || ActiveOLCRTCOptions.Name == cleanTag {
			return ActiveOLCRTCOptions
		}
	}
	return nil
}

func IsOLCRTCTag(tag string) bool {
	cleanTag := strings.TrimSpace(strings.Split(tag, "§")[0])
	cleanLower := strings.ToLower(cleanTag)
	if strings.Contains(cleanLower, "olcrtc") || strings.Contains(cleanLower, "olconnect") {
		return true
	}
	return GetOLCRTCOption(tag) != nil
}

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
	trimmed := strings.TrimSpace(s)
	return strings.HasPrefix(trimmed, "olcrtc://") || strings.HasPrefix(trimmed, "olconnect://")
}

func TryDecodeSubscription(s string) string {
	trimmed := strings.TrimSpace(s)
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if dec, err := enc.DecodeString(trimmed); err == nil && len(dec) > 0 {
			decStr := string(dec)
			if strings.Contains(decStr, "://") || strings.Contains(decStr, "\n") {
				return decStr
			}
		}
	}
	return s
}

func DetectOLCRTCOptions(content string, path string) *OLCRTCOptions {
	// 1. Check sidecar file first (loaded from previously saved profile)
	if path != "" {
		sidecar := path + ".olcrtc"
		if data, err := os.ReadFile(sidecar); err == nil {
			var store OLCRTCStore
			if err := json.Unmarshal(data, &store); err == nil && len(store.Options) > 0 {
				for tag, opt := range store.Options {
					RegisterOLCRTCOption(tag, opt)
				}
				ActiveOLCRTCStore = &store
				if active, ok := store.Options[store.Active]; ok {
					ActiveOLCRTCOptions = active
					return active
				}
				for _, opt := range store.Options {
					ActiveOLCRTCOptions = opt
					return opt
				}
			}

			var singleOpt OLCRTCOptions
			if err := json.Unmarshal(data, &singleOpt); err == nil && singleOpt.RoomID != "" {
				RegisterOLCRTCOption(singleOpt.Name, &singleOpt)
				ActiveOLCRTCOptions = &singleOpt
				return &singleOpt
			}
		}
	}

	// 2. Check content (may be base64 or multi-line containing olcrtc:// links)
	sourceContent := content
	if sourceContent == "" && path != "" {
		if data, err := os.ReadFile(path); err == nil {
			sourceContent = string(data)
		}
	}

	if sourceContent != "" {
		decoded := TryDecodeSubscription(sourceContent)
		lines := strings.Split(strings.ReplaceAll(decoded, "\r\n", "\n"), "\n")
		store := &OLCRTCStore{Options: make(map[string]*OLCRTCOptions)}
		basePort := 10808
		for i, rawLine := range lines {
			line := strings.TrimSpace(rawLine)
			if isOLCRTCUri(line) {
				if opt, err := ParseOLCRTCURI(line); err == nil {
					if opt.SocksPort <= 0 {
						opt.SocksPort = basePort + i
					}
					tag := opt.Name
					if tag == "" {
						tag = fmt.Sprintf("olcRTC %s", opt.Provider)
					}
					if _, exists := store.Options[tag]; exists {
						tag = fmt.Sprintf("%s (%d)", tag, i+1)
						opt.Name = tag
					}
					store.Options[tag] = opt
					RegisterOLCRTCOption(tag, opt)
					if store.Active == "" {
						store.Active = tag
						ActiveOLCRTCOptions = opt
					}
				}
			}
		}

		if len(store.Options) > 0 {
			ActiveOLCRTCStore = store
			if path != "" {
				_ = SaveOLCRTCStore(path, store)
			}
			return ActiveOLCRTCOptions
		}

		if strings.Contains(sourceContent, "_olcrtc") {
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(sourceContent), &raw); err == nil {
				if olcData, ok := raw["_olcrtc"]; ok {
					if b, err := json.Marshal(olcData); err == nil {
						var opt OLCRTCOptions
						if err := json.Unmarshal(b, &opt); err == nil && opt.RoomID != "" {
							RegisterOLCRTCOption(opt.Name, &opt)
							ActiveOLCRTCOptions = &opt
							return &opt
						}
					}
				}
			}
		}
	}

	if ActiveOLCRTCOptions != nil {
		return ActiveOLCRTCOptions
	}
	return nil
}

func SaveOLCRTCOptions(path string, opts *OLCRTCOptions) error {
	if path == "" {
		return nil
	}
	if ActiveOLCRTCStore != nil && len(ActiveOLCRTCStore.Options) > 0 {
		return SaveOLCRTCStore(path, ActiveOLCRTCStore)
	}
	if opts == nil {
		return nil
	}
	b, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path+".olcrtc", b, 0o644)
}

func SaveOLCRTCStore(path string, store *OLCRTCStore) error {
	if path == "" || store == nil || len(store.Options) == 0 {
		return nil
	}
	b, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path+".olcrtc", b, 0o644)
}

