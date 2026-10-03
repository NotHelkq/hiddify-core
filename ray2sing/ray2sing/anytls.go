package ray2sing

import (
	C "github.com/sagernet/sing-box/constant"
	T "github.com/sagernet/sing-box/option"
)

func AnyTLSSingbox(anytlsURL string) (*T.Outbound, error) {
	u, err := ParseUrl(anytlsURL, 443)
	if err != nil {
		return nil, err
	}
	decoded := u.Params
	if decoded["security"] == "" {
		decoded["security"] = "tls"
	}
	tlsOptions := getTLSOptions(decoded)

	pass := u.Password
	if pass == "" {
		pass = u.Username
	}

	return &T.Outbound{
		Tag:  u.Name,
		Type: C.TypeAnyTLS,
		Options: &T.AnyTLSOutboundOptions{
			DialerOptions:               getDialerOptions(decoded),
			ServerOptions:               u.GetServerOption(),
			Password:                    pass,
			OutboundTLSOptionsContainer: tlsOptions,
		},
	}, nil
}
