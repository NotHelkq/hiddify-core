package ray2sing

import (
	"strings"

	T "github.com/sagernet/sing-box/option"
)

func DnsttSingbox(vlessURL string) (*T.Outbound, error) {
	u, err := ParseUrl(vlessURL, 443)
	if err != nil {
		return nil, err
	}
	decoded := u.Params
	// uot := T.UDPOverTCPOptions{
	// 	Enabled: getOneOfN(decoded, "", "uot") != "false" && getOneOfN(decoded, "", "uot") != "0",
	// }
	pubKey := getOneOfN(decoded, "", "pubkey", "publickey", "serverpublickey")
	if pubKey == "" {
		pubKey = u.Username
	}
	domain := getOneOfN(decoded, "", "domain", "serveraddress", "address")
	if domain == "" {
		domain = u.Hostname
	}
	resolvers := strings.Split(getOneOfN(decoded, "", "resolver"), ",")
	if len(resolvers) == 0 || (len(resolvers) == 1 && resolvers[0] == "") {
		resolvers = []string{"1.1.1.1:53"}
	}

	d := &T.DnsttOptions{
		DialerOptions: getDialerOptions(decoded),
		PublicKey:     pubKey,
		Domain:        domain,
		Resolvers:     resolvers,
		// TunnelPerResolver: toInt(getOneOfN(decoded, "1", "tunnelperresolver")),

		PreTestDomain:     getOneOfN(decoded, "", "pretest-domain"),
		PreTestRecordType: getOneOfN(decoded, "", "pretest-record-type"),
		RecordType:        getOneOfN(decoded, "", "record-type"),
		UTLSClientHelloID: getOneOfN(decoded, "", "utls"),

		DnsttCompat:    toBool(getOneOfN(decoded, "false", "dnstt-compat"), false),
		ClientIDSize:   toIntN(getOneOfN(decoded, "", "clientid-size")),
		MaxQnameLen:    toIntN(getOneOfN(decoded, "", "max-qname-len")),
		MaxNumLabels:   toIntN(getOneOfN(decoded, "", "max-num-labels")),
		RPS:            toFloatN(getOneOfN(decoded, "", "rps")),
		SingleResolver: toBool(getOneOfN(decoded, "false", "single-resolver"), false),

		MTU: toIntN(getOneOfN(decoded, "0", "mtu")),

		MaxStreams: toIntN(getOneOfN(decoded, "", "max-streams")),

		// IdleTimeout          *badoption.Duration `json:"idle-timeout,omitempty"`
		// KeepAlive            *badoption.Duration `json:"keepalive,omitempty"`
		// OpenStreamTimeout    *badoption.Duration `json:"open-stream-timeout,omitempty"`

		// ReconnectMinDelay    *badoption.Duration `json:"reconnect-min,omitempty"`
		// ReconnectMaxDelay    *badoption.Duration `json:"reconnect-max,omitempty"`
		// SessionCheckInterval *badoption.Duration `json:"session-check-interval,omitempty"`
		// HandshakeTimeout     *badoption.Duration `json:"handshake-timeout,omitempty"`
		// UdpTimeout      *badoption.Duration `json:"udp-timeout,omitempty"`
		UdpAcceptErrors: toBool(getOneOfN(decoded, "false", "udp-accept-errors"), false),
		UdpSharedSocket: toBool(getOneOfN(decoded, "true", "udp-shared-socket"), true),
		UdpWorkers:      toIntN(getOneOfN(decoded, "", "udp-workers")),

		// UDPOverTCP: &uot,
	}

	if d.ClientIDSize == nil && d.RecordType == "" && getOneOfN(decoded, "", "dnstt-compat") == "" {
		d.DnsttCompat = true
	}
	return &T.Outbound{
		Tag:     u.Name,
		Type:    "dnstt",
		Options: d,
	}, nil
}
