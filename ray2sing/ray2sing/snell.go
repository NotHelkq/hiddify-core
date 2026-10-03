package ray2sing

import (
	"strings"

	C "github.com/sagernet/sing-box/constant"
	T "github.com/sagernet/sing-box/option"
)

func SnellSingbox(snellURL string) (*T.Outbound, error) {
	u, err := ParseUrl(snellURL, 443)
	if err != nil {
		return nil, err
	}
	decoded := u.Params

	psk := getOneOfN(decoded, "", "psk")
	if psk == "" {
		psk = u.Password
	}
	if psk == "" {
		psk = u.Username
	}

	version := toInt(getOneOfN(decoded, "4", "version", "v"))
	obfsMode := getOneOfN(decoded, "", "obfs", "obfs-mode", "obfsmode")
	obfsHost := getOneOfN(decoded, "", "obfs-host", "obfshost", "host")

	return &T.Outbound{
		Tag:  u.Name,
		Type: C.TypeSnell,
		Options: &T.SnellOutboundOptions{
			DialerOptions: getDialerOptions(decoded),
			ServerOptions: u.GetServerOption(),
			PSK:           psk,
			Version:       version,
			Reuse:         toBool(getOneOfN(decoded, "true", "reuse"), true),
			ObfsMode:      strings.ToLower(obfsMode),
			ObfsHost:      obfsHost,
		},
	}, nil
}
