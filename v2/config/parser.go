package config

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hiddify/ray2sing/ray2sing"
	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/batch"
	SJ "github.com/sagernet/sing/common/json"
	"github.com/xmdhs/clash2singbox/convert"
	clash2singmodel "github.com/xmdhs/clash2singbox/model"
	"github.com/xmdhs/clash2singbox/model/clash"
	"gopkg.in/yaml.v3"
)

//go:embed config.json.template
var configByte []byte

func ReadContent(ctx context.Context, opt *ReadOptions) ([]byte, error) {
	if opt.Content == "" {
		contentBytes, err := os.ReadFile(opt.Path)
		if err != nil {
			return nil, err
		}
		opt.Content = string(contentBytes)
	}
	return []byte(opt.Content), nil
}

func ParseConfig(ctx context.Context, opt *ReadOptions, debug bool, configOpt *HiddifyOptions, fullConfig bool) (*option.Options, error) {
	content, err := ReadContent(ctx, opt)
	if err != nil {
		return nil, err
	}
	return parseConfigContent(ctx, content, debug, nil, false)
}

func ParseConfigBytes(ctx context.Context, opt *ReadOptions, debug bool, configOpt *HiddifyOptions, fullConfig bool) ([]byte, error) {

	options, err := ParseConfig(ctx, opt, debug, configOpt, fullConfig)
	if err != nil {
		return nil, err
	}

	return options.MarshalJSONContext(ctx)

}
func parseConfigContent(ctx context.Context, content []byte, debug bool, configOpt *HiddifyOptions, fullConfig bool) (*option.Options, error) {
	if configOpt == nil {
		configOpt = DefaultHiddifyOptions()
	}

	var jsonObj map[string]interface{} = make(map[string]interface{})

	var tmpJsonResult any
	jsonDecoder := json.NewDecoder(SJ.NewCommentFilter(bytes.NewReader(content)))
	if err := jsonDecoder.Decode(&tmpJsonResult); err == nil {
		fmt.Printf("Convert using json\n")
		if tmpJsonObj, ok := tmpJsonResult.(map[string]interface{}); ok {
			if tmpJsonObj["outbounds"] == nil && tmpJsonObj["endpoints"] == nil {
				jsonObj["outbounds"] = []interface{}{jsonObj}
			} else {
				if fullConfig || (configOpt != nil && configOpt.EnableFullConfig) {
					jsonObj = tmpJsonObj
				} else {
					if tmpJsonObj["outbounds"] != nil {
						jsonObj["outbounds"] = tmpJsonObj["outbounds"]
					}
					if tmpJsonObj["endpoints"] != nil {
						jsonObj["endpoints"] = tmpJsonObj["endpoints"]
					}
				}
			}
		} else if jsonArray, ok := tmpJsonResult.([]map[string]interface{}); ok {
			jsonObj["outbounds"] = jsonArray
		} else {
			return nil, fmt.Errorf("[SingboxParser] Incorrect Json Format")
		}

		newContent, _ := json.MarshalIndent(jsonObj, "", "  ")

		return patchConfigStr(ctx, newContent, "SingboxParser", configOpt)
	}

	fmt.Printf("Convert using clash\n")
	clashObj := clash.Clash{}
	if err := yaml.Unmarshal(content, &clashObj); err == nil && clashObj.Proxies != nil {
		if len(clashObj.Proxies) == 0 {
			return nil, fmt.Errorf("[ClashParser] no outbounds found")
		}
		converted, endpoints, err := convert.Clash2sing(clashObj, clash2singmodel.SINGLATEST)
		if err != nil {
			return nil, fmt.Errorf("[ClashParser] converting clash to sing-box error: %w", err)
		}
		output := configByte
		output, err = convert.Patch(output, converted, endpoints, "", "", nil)
		if err != nil {
			return nil, fmt.Errorf("[ClashParser] patching clash config error: %w", err)
		}
		return patchConfigStr(ctx, output, "ClashParser", configOpt)
	}

	return parseSubscription(ctx, content, configOpt)
}

func parseSubscription(ctx context.Context, content []byte, configOpt *HiddifyOptions) (*option.Options, error) {
	decodedStr := TryDecodeSubscription(string(content))
	rawLines := strings.Split(strings.ReplaceAll(decodedStr, "\r\n", "\n"), "\n")

	var olcrtcLines []string
	var otherLines []string

	for _, rawLine := range rawLines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if isOLCRTCUri(line) {
			olcrtcLines = append(olcrtcLines, line)
		} else {
			otherLines = append(otherLines, line)
		}
	}

	// Case 1: No olcRTC lines found -> standard ray2sing parser
	if len(olcrtcLines) == 0 {
		v2ray, err := ray2sing.Ray2SingboxOptions(ctx, string(content), configOpt.UseXrayCoreWhenPossible)
		if err == nil {
			return patchConfigOptions(ctx, v2ray, "V2rayParser", configOpt)
		}
		return nil, fmt.Errorf("unable to determine config format: %w", err)
	}

	// Case 2: One or more olcRTC lines found!
	store := &OLCRTCStore{Options: make(map[string]*OLCRTCOptions)}
	var olcOutbounds []option.Outbound
	var olcOutboundsJSON []map[string]interface{}
	basePort := 10808

	for i, line := range olcrtcLines {
		opt, err := ParseOLCRTCURI(line)
		if err != nil {
			continue
		}
		port := opt.SocksPort
		if port <= 0 {
			port = basePort + i
			opt.SocksPort = port
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
		if store.Active == "" {
			store.Active = tag
			ActiveOLCRTCOptions = opt
		}
		RegisterOLCRTCOption(tag, opt)

		socksMap := map[string]interface{}{
			"type":        "socks",
			"tag":         tag,
			"server":      "127.0.0.1",
			"server_port": port,
			"version":     "5",
		}
		olcOutboundsJSON = append(olcOutboundsJSON, socksMap)

		var ob option.Outbound
		b, _ := json.Marshal(socksMap)
		if err := ob.UnmarshalJSONContext(ctx, b); err == nil {
			olcOutbounds = append(olcOutbounds, ob)
		}
	}

	if len(store.Options) > 0 {
		ActiveOLCRTCStore = store
	}

	// Sub-case 2a: Mixed subscription - other proxy protocols are also present
	if len(otherLines) > 0 {
		otherContent := strings.Join(otherLines, "\n")
		v2ray, err := ray2sing.Ray2SingboxOptions(ctx, otherContent, configOpt.UseXrayCoreWhenPossible)
		if err == nil && v2ray != nil {
			v2ray.Outbounds = append(v2ray.Outbounds, olcOutbounds...)
			return patchConfigOptions(ctx, v2ray, "V2rayParser", configOpt)
		}
	}

	// Sub-case 2b: Pure olcRTC subscription (or ray2sing produced no outbounds)
	if len(olcOutboundsJSON) > 0 {
		res := map[string]interface{}{
			"outbounds": olcOutboundsJSON,
		}
		b, _ := json.Marshal(res)
		return patchConfigStr(ctx, b, "OLCRTCParser", configOpt)
	}

	return nil, fmt.Errorf("failed to parse subscription with olcrtc")
}

func patchConfigStr(ctx context.Context, content []byte, name string, configOpt *HiddifyOptions) (*option.Options, error) {
	options := option.Options{}
	err := options.UnmarshalJSONContext(ctx, content)

	if err != nil {
		return nil, fmt.Errorf("[SingboxParser] unmarshal error: %w", err)
	}

	return patchConfigOptions(ctx, &options, name, configOpt)
}
func patchConfigOptions(ctx context.Context, options *option.Options, name string, configOpt *HiddifyOptions) (*option.Options, error) {
	b, _ := batch.New(ctx, batch.WithConcurrencyNum[*option.Endpoint](2))
	for _, base := range options.Endpoints {
		out := base
		b.Go(base.Tag, func() (*option.Endpoint, error) {
			err := patchWarp(&out, configOpt, false, nil)
			if err != nil {
				return nil, fmt.Errorf("[Warp] patch warp error: %w", err)
			}
			// options.Outbounds[i] = base
			return &out, nil
		})
	}
	if res, err := b.WaitAndGetResult(); err != nil {
		return nil, err
	} else {
		for i, base := range options.Endpoints {
			options.Endpoints[i] = *res[base.Tag].Value
		}
	}

	// fmt.Printf("%s\n", content)
	return validateResult(ctx, options, name)
}

func validateResult(ctx context.Context, options *option.Options, name string) (*option.Options, error) {
	err := libbox.CheckConfigOptions(options)
	if err != nil {
		return nil, fmt.Errorf("[%s] invalid sing-box config: %w", name, err)
	}
	return options, nil
}
