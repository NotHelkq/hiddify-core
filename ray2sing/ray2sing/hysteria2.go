package ray2sing

import (
	T "github.com/sagernet/sing-box/option"
)

func Hysteria2Singbox(hysteria2Url string) (*T.Outbound, error) {
	u, err := ParseUrl(hysteria2Url, 443)
	if err != nil {
		return nil, err
	}
	decoded := u.Params
	var ObfsOpts *T.Hysteria2Obfs
	ObfsOpts = nil
	if obfs, ok := decoded["obfs"]; ok && obfs != "" {
		pass := decoded["obfs-password"]
		if pass == "" {
			pass = decoded["obfspassword"]
		}
		if pass == "" {
			pass = decoded["obfs_password"]
		}
		if pass == "" {
			pass = decoded["obfsparam"]
		}
		ObfsOpts = &T.Hysteria2Obfs{
			Type:     obfs,
			Password: pass,
		}
	}

	valECH, hasECH := decoded["ech"]
	hasECH = hasECH && (valECH != "0")
	var ECHOpts *T.OutboundECHOptions
	ECHOpts = nil
	if hasECH {
		ECHOpts = &T.OutboundECHOptions{
			Enabled: hasECH,
		}
	}

	SNI := decoded["sni"]
	if SNI == "" {
		SNI = decoded["hostname"]
	}
	if SNI == "" {
		SNI = u.Hostname
	}
	insecure := decoded["insecure"] == "1" || decoded["insecure"] == "true" || decoded["allowinsecure"] == "1" || isIPOnly(SNI) || isIPOnly(u.Hostname)
	// turnRelay, err := u.GetRelayOptions()
	// if err != nil {
	// 	return nil, err
	// }
	pass := u.Username
	if u.Password != "" {
		pass += ":" + u.Password
	}
	result := T.Outbound{
		Type: "hysteria2",
		Tag:  u.Name,
		Options: &T.Hysteria2OutboundOptions{
			ServerOptions: u.GetServerOption(),
			Obfs:          ObfsOpts,
			Password:      pass,
			OutboundTLSOptionsContainer: T.OutboundTLSOptionsContainer{
				TLS: &T.OutboundTLSOptions{
					Enabled:    true,
					Insecure:   insecure,
					DisableSNI: isIPOnly(SNI),
					ServerName: SNI,
					ECH:        ECHOpts,
				},
			},
			// TurnRelay: turnRelay,
		},
	}

	return &result, nil
}
