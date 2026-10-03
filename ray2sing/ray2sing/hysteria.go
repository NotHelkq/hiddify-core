package ray2sing

import (
	"strconv"

	T "github.com/sagernet/sing-box/option"
)

func HysteriaSingbox(hysteriaURL string) (*T.Outbound, error) {
	u, err := ParseUrl(hysteriaURL, 443)
	if err != nil {
		return nil, err
	}
	SNI := u.Params["peer"]
	if SNI == "" {
		SNI = u.Params["sni"]
	}
	if SNI == "" {
		SNI = u.Hostname
	}
	insecure := u.Params["insecure"] == "1" || u.Params["insecure"] == "true" || u.Params["allowinsecure"] == "1" || isIPOnly(SNI) || isIPOnly(u.Hostname)
	opts := T.HysteriaOutboundOptions{
		ServerOptions: u.GetServerOption(),
		OutboundTLSOptionsContainer: T.OutboundTLSOptionsContainer{
			TLS: &T.OutboundTLSOptions{
				Enabled:    true,
				DisableSNI: isIPOnly(SNI),
				ServerName: SNI,
				Insecure:   insecure,
			},
		},
	}
	singOut := &T.Outbound{
		Type:    u.Scheme,
		Tag:     u.Name,
		Options: &opts,
	}

	opts.AuthString = u.Params["auth"]

	upMbps, err := strconv.Atoi(u.Params["upmbps"])
	if err == nil {
		opts.UpMbps = upMbps
	}

	downMbps, err := strconv.Atoi(u.Params["downmbps"])
	if err == nil {
		opts.DownMbps = downMbps
	}

	obfs := u.Params["obfs"]
	if obfs == "" {
		obfs = u.Params["obfsparam"]
	}
	if obfs == "" {
		obfs = u.Params["obfsParam"]
	}
	opts.Obfs = obfs
	return singOut, nil
}
