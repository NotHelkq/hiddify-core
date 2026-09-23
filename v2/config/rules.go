package config

import (
	"strconv"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
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
		if strings.HasPrefix(item, "geosite:") {
			raw.Geosite = append(raw.Geosite, strings.TrimPrefix(item, "geosite:"))
		} else if strings.HasPrefix(item, "geoip:") {
			raw.GeoIP = append(raw.GeoIP, strings.TrimPrefix(item, "geoip:"))
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
		if strings.HasPrefix(item, "geosite:") {
			raw.Geosite = append(raw.Geosite, strings.TrimPrefix(item, "geosite:"))
		} else {
			raw.RuleSet = append(raw.RuleSet, item)
		}
	}

	if len(raw.Domain) == 0 && len(raw.DomainSuffix) == 0 && len(raw.DomainKeyword) == 0 &&
		len(raw.DomainRegex) == 0 && len(raw.Geosite) == 0 && len(raw.RuleSet) == 0 && len(raw.PackageName) == 0 {
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
