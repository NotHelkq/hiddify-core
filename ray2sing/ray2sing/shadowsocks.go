package ray2sing

import (
	"strings"

	T "github.com/sagernet/sing-box/option"
)

func ShadowsocksSingbox(shadowsocksUrl string) (*T.Outbound, error) {
	u, err := ParseUrl(shadowsocksUrl, 443)
	if err != nil {
		return nil, err
	}

	decoded := u.Params
	
	defaultMethod := u.Username
	pass:=u.Password
	if u.Password == "" {
		pass = u.Username
		defaultMethod = "none"
	}
	

	plugin := decoded["plugin"]
	pluginOpts := decoded["pluginopts"]
	if pluginOpts == "" {
		pluginOpts = decoded["plugin-opts"]
	}
	if strings.Contains(plugin, ";") {
		parts := strings.SplitN(plugin, ";", 2)
		plugin = parts[0]
		if pluginOpts == "" {
			pluginOpts = parts[1]
		}
	}

	result := T.Outbound{
		Type: "shadowsocks",
		Tag:  u.Name,
		Options: &T.ShadowsocksOutboundOptions{
			ServerOptions: u.GetServerOption(),
			Method:        defaultMethod,
			Password:      pass,
			Plugin:        plugin,
			PluginOptions: pluginOpts,
		},
	}

	return &result, nil
}
