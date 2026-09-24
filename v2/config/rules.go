package config

import (
	"encoding/json"
	"strconv"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"google.golang.org/protobuf/encoding/protojson"
)

func (r *Rule) MakeRule() (option.Rule, bool) {
	if r == nil || !r.Enabled {
		return option.Rule{}, false
	}
	raw := option.RawDefaultRule{
		Domain:        r.Domains,
		DomainSuffix:  r.DomainSuffixes,
		DomainKeyword: r.DomainKeywords,
		DomainRegex:   r.DomainRegexes,
		IPCIDR:        r.IpCidrs,
		SourceIPCIDR:  r.SourceIpCidrs,
		PackageName:   r.PackageNames,
		ProcessName:   r.ProcessNames,
		ProcessPath:   r.ProcessPaths,
	}

	for _, item := range r.RuleSets {
		itemLower := strings.ToLower(item)
		if strings.HasPrefix(itemLower, "geosite:spotify") || itemLower == "spotify" {
			raw.DomainSuffix = append(raw.DomainSuffix, "spotify.com", "scdn.co", "spoti.fi", "spotifycdn.com", "spotifycdn.net")
		} else if strings.HasPrefix(itemLower, "geosite:category-ru") || itemLower == "category-ru" {
			raw.DomainRegex = append(raw.DomainRegex, `.*\.ru$`, `.*\.xn--p1ai$`, `.*\.su$`)
			raw.DomainSuffix = append(raw.DomainSuffix, "ru", ".ru", "su", ".su", "xn--p1ai", ".xn--p1ai", "yandex.ru", "ya.ru", "vk.com", "vk.me", "ok.ru", "mail.ru", "gosuslugi.ru", "avito.ru", "avito.st", "tinkoff.ru", "t-bank.ru", "sberbank.ru", "ozon.ru", "wildberries.ru", "rutube.ru", "dzen.ru")
		} else if strings.HasPrefix(item, "geosite:") {
			tag := "geosite-" + strings.TrimPrefix(item, "geosite:")
			raw.RuleSet = append(raw.RuleSet, tag)
		} else if strings.HasPrefix(item, "geoip:") {
			tag := "geoip-" + strings.TrimPrefix(item, "geoip:")
			raw.RuleSet = append(raw.RuleSet, tag)
		} else {
			raw.RuleSet = append(raw.RuleSet, item)
		}
	}

	for _, p := range r.PortRanges {
		if strings.Contains(p, ":") {
			raw.PortRange = append(raw.PortRange, p)
		} else if port, err := strconv.Atoi(p); err == nil {
			raw.Port = append(raw.Port, uint16(port))
		}
	}
	for _, p := range r.SourcePortRanges {
		if strings.Contains(p, ":") {
			raw.SourcePortRange = append(raw.SourcePortRange, p)
		} else if port, err := strconv.Atoi(p); err == nil {
			raw.SourcePort = append(raw.SourcePort, uint16(port))
		}
	}

	for _, proto := range r.Protocols {
		switch proto {
		case Protocol_tls:
			raw.Protocol = append(raw.Protocol, C.ProtocolTLS)
		case Protocol_http:
			raw.Protocol = append(raw.Protocol, C.ProtocolHTTP)
		case Protocol_quic:
			raw.Protocol = append(raw.Protocol, C.ProtocolQUIC)
		case Protocol_stun:
			raw.Protocol = append(raw.Protocol, C.ProtocolSTUN)
		case Protocol_dns:
			raw.Protocol = append(raw.Protocol, C.ProtocolDNS)
		case Protocol_bittorrent:
			raw.Protocol = append(raw.Protocol, C.ProtocolBitTorrent)
		}
	}

	if r.Network == Network_tcp {
		raw.Network = append(raw.Network, "tcp")
	} else if r.Network == Network_udp {
		raw.Network = append(raw.Network, "udp")
	}

	var action option.RuleAction
	switch r.Outbound {
	case Outbound_direct:
		action = option.RuleAction{
			Action: C.RuleActionTypeRoute,
			RouteOptions: option.RouteActionOptions{
				Outbound: OutboundDirectTag,
			},
		}
	case Outbound_block:
		action = option.RuleAction{
			Action: C.RuleActionTypeReject,
			RejectOptions: option.RejectActionOptions{
				Method: C.RuleActionRejectMethodDefault,
			},
		}
	case Outbound_direct_with_fragment:
		action = option.RuleAction{
			Action: C.RuleActionTypeRoute,
			RouteOptions: option.RouteActionOptions{
				Outbound: OutboundDirectFragmentTag,
			},
		}
	default: // Outbound_proxy
		action = option.RuleAction{
			Action: C.RuleActionTypeRoute,
			RouteOptions: option.RouteActionOptions{
				Outbound: OutboundSelectTag,
			},
		}
	}

	return option.Rule{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultRule{
			RawDefaultRule: raw,
			RuleAction:     action,
		},
	}, true
}

func (r *Rule) MakeDNSRule() (option.DefaultDNSRule, bool) {
	if r == nil || !r.Enabled {
		return option.DefaultDNSRule{}, false
	}
	raw := option.RawDefaultDNSRule{
		Domain:        r.Domains,
		DomainSuffix:  r.DomainSuffixes,
		DomainKeyword: r.DomainKeywords,
		DomainRegex:   r.DomainRegexes,
		PackageName:   r.PackageNames,
	}

	for _, item := range r.RuleSets {
		itemLower := strings.ToLower(item)
		if strings.HasPrefix(itemLower, "geosite:spotify") || itemLower == "spotify" {
			raw.DomainSuffix = append(raw.DomainSuffix, "spotify.com", "scdn.co", "spoti.fi", "spotifycdn.com", "spotifycdn.net")
		} else if strings.HasPrefix(itemLower, "geosite:category-ru") || itemLower == "category-ru" {
			raw.DomainRegex = append(raw.DomainRegex, `.*\.ru$`, `.*\.xn--p1ai$`, `.*\.su$`)
			raw.DomainSuffix = append(raw.DomainSuffix, "ru", ".ru", "su", ".su", "xn--p1ai", ".xn--p1ai", "yandex.ru", "ya.ru", "vk.com", "vk.me", "ok.ru", "mail.ru", "gosuslugi.ru", "avito.ru", "avito.st", "tinkoff.ru", "t-bank.ru", "sberbank.ru", "ozon.ru", "wildberries.ru", "rutube.ru", "dzen.ru")
		} else if strings.HasPrefix(item, "geosite:") {
			tag := "geosite-" + strings.TrimPrefix(item, "geosite:")
			raw.RuleSet = append(raw.RuleSet, tag)
		} else {
			raw.RuleSet = append(raw.RuleSet, item)
		}
	}

	if len(raw.Domain) == 0 && len(raw.DomainSuffix) == 0 && len(raw.DomainKeyword) == 0 &&
		len(raw.DomainRegex) == 0 && len(raw.RuleSet) == 0 && len(raw.PackageName) == 0 {
		return option.DefaultDNSRule{}, false
	}

	var action option.DNSRuleAction
	switch r.Outbound {
	case Outbound_direct, Outbound_direct_with_fragment:
		action = option.DNSRuleAction{
			Action: C.RuleActionTypeRoute,
			RouteOptions: option.DNSRouteActionOptions{
				Server: DNSMultiDirectTag,
			},
		}
	case Outbound_block:
		action = option.DNSRuleAction{
			Action: C.RuleActionTypeReject,
		}
	default:
		action = option.DNSRuleAction{
			Action: C.RuleActionTypeRoute,
			RouteOptions: option.DNSRouteActionOptions{
				Server: DNSMultiRemoteTag,
			},
		}
	}

	return option.DefaultDNSRule{
		RawDefaultDNSRule: raw,
		DNSRuleAction:     action,
	}, true
}

func (x *RouteRule) UnmarshalJSON(b []byte) error {
	if x == nil {
		return nil
	}
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*x = RouteRule{}
		return nil
	}
	s = strings.ReplaceAll(s, `"bypass"`, `"direct"`)
	s = strings.ReplaceAll(s, `"reject"`, `"block"`)
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(s), x)
}

func (x *RouteRule) MarshalJSON() ([]byte, error) {
	if x == nil {
		return []byte("null"), nil
	}
	return protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}.Marshal(x)
}

func (x *Rule) UnmarshalJSON(b []byte) error {
	if x == nil {
		return nil
	}
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*x = Rule{}
		return nil
	}
	s = strings.ReplaceAll(s, `"bypass"`, `"direct"`)
	s = strings.ReplaceAll(s, `"reject"`, `"block"`)
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(s), x)
}

func (x *Rule) MarshalJSON() ([]byte, error) {
	if x == nil {
		return []byte("null"), nil
	}
	return protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}.Marshal(x)
}

func (x *Outbound) UnmarshalJSON(b []byte) error {
	var n int32
	if err := json.Unmarshal(b, &n); err == nil {
		*x = Outbound(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "direct", "bypass":
		*x = Outbound_direct
	case "direct_with_fragment", "fragment":
		*x = Outbound_direct_with_fragment
	case "block", "reject":
		*x = Outbound_block
	case "proxy":
		*x = Outbound_proxy
	default:
		if val, ok := Outbound_value[s]; ok {
			*x = Outbound(val)
		} else {
			*x = Outbound_proxy
		}
	}
	return nil
}

func (x Outbound) MarshalJSON() ([]byte, error) {
	s, ok := Outbound_name[int32(x)]
	if ok {
		return json.Marshal(s)
	}
	return json.Marshal("proxy")
}

func (x *Network) UnmarshalJSON(b []byte) error {
	var n int32
	if err := json.Unmarshal(b, &n); err == nil {
		*x = Network(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "tcp":
		*x = Network_tcp
	case "udp":
		*x = Network_udp
	default:
		*x = Network_all
	}
	return nil
}

func (x Network) MarshalJSON() ([]byte, error) {
	s, ok := Network_name[int32(x)]
	if ok {
		return json.Marshal(s)
	}
	return json.Marshal("all")
}

func (x *Protocol) UnmarshalJSON(b []byte) error {
	var n int32
	if err := json.Unmarshal(b, &n); err == nil {
		*x = Protocol(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "tls":
		*x = Protocol_tls
	case "http":
		*x = Protocol_http
	case "quic":
		*x = Protocol_quic
	case "stun":
		*x = Protocol_stun
	case "dns":
		*x = Protocol_dns
	case "bittorrent":
		*x = Protocol_bittorrent
	default:
		if val, ok := Protocol_value[s]; ok {
			*x = Protocol(val)
		} else {
			*x = Protocol_tls
		}
	}
	return nil
}

func (x Protocol) MarshalJSON() ([]byte, error) {
	s, ok := Protocol_name[int32(x)]
	if ok {
		return json.Marshal(s)
	}
	return json.Marshal("tls")
}
