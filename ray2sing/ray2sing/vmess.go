package ray2sing

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	T "github.com/sagernet/sing-box/option"
)

func decodeVmess(vmessConfig string) (map[string]string, error) {
	vmessConfig = strings.TrimSpace(vmessConfig)
	idx := strings.Index(vmessConfig, "://")
	if idx == -1 {
		return nil, fmt.Errorf("invalid vmess config: missing scheme delimiter")
	}
	vmessData := strings.TrimSpace(vmessConfig[idx+3:])

	var fragment string
	if hashIdx := strings.Index(vmessData, "#"); hashIdx != -1 {
		rawFrag := vmessData[hashIdx+1:]
		if unescaped, err := url.QueryUnescape(rawFrag); err == nil && unescaped != "" {
			fragment = strings.TrimSpace(unescaped)
		} else {
			fragment = strings.TrimSpace(rawFrag)
		}
		vmessData = vmessData[:hashIdx]
	}

	if qIdx := strings.Index(vmessData, "?"); qIdx != -1 {
		vmessData = vmessData[:qIdx]
	}

	vmessData = strings.TrimRight(strings.TrimSpace(vmessData), "/")

	decodedData, err := decodeBase64FaultTolerant(vmessData)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(decodedData), &data)
	if err != nil {
		return nil, err
	}
	strdata := convertToStrings(data)
	if fragment != "" {
		strdata["ps"] = fragment
	} else if ps, ok := strdata["ps"]; ok {
		strdata["ps"] = strings.TrimSpace(ps)
	}
	return strdata, nil
}

func convertToStrings(data map[string]interface{}) map[string]string {
	stringMap := make(map[string]string)
	for key, value := range data {
		switch v := value.(type) {
		case string:
			stringMap[key] = v
		case float64:
			stringMap[key] = strconv.FormatInt(int64(v), 10)
		case bool:
			stringMap[key] = strconv.FormatBool(v)
		case nil:
			stringMap[key] = ""
		case map[string]interface{}, []interface{}:
			b, err := json.Marshal(v)
			if err == nil {
				stringMap[key] = string(b)
			} else {
				stringMap[key] = fmt.Sprintf("%v", v)
			}
		default:
			stringMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return stringMap
}

func VmessSingbox(vmessURL string) (*T.Outbound, error) {
	decoded, err := decodeVmess(vmessURL)
	if err != nil {
		return nil, err
	}

	port := toUInt16(decoded["port"], 443)
	transportOptions, err := getTransportOptions(decoded)
	if err != nil {
		return nil, err
	}
	security := "auto"
	if decoded["scy"] != "" {
		security = decoded["scy"]
	}
	packetEncoding := decoded["packetEncoding"]
	if packetEncoding == "" {
		packetEncoding = "xudp"
	}
	return &T.Outbound{
		Tag:  decoded["ps"],
		Type: "vmess",
		Options: &T.VMessOutboundOptions{
			DialerOptions: getDialerOptions(decoded),
			ServerOptions: T.ServerOptions{
				Server:     decoded["add"],
				ServerPort: port,
			},
			UUID:                        decoded["id"],
			Security:                    security,
			AlterId:                     toInt(decoded["aid"]),
			GlobalPadding:               false,
			AuthenticatedLength:         true,
			PacketEncoding:              packetEncoding,
			OutboundTLSOptionsContainer: getTLSOptions(decoded),
			Transport:                   transportOptions,
			Multiplex:                   getMuxOptions(decoded),
		},
	}, nil
}
