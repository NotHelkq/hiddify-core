package hcore

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hiddify/hiddify-core/v2/config"
	"github.com/hiddify/hiddify-core/v2/db"
	hcommon "github.com/hiddify/hiddify-core/v2/hcommon"
	hutils "github.com/hiddify/hiddify-core/v2/hutils"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/option"
)

func BuildConfigJson(ctx context.Context, in *StartRequest) (string, error) {
	Log(LogLevel_DEBUG, LogType_CORE, "Stating Service ")

	parsedContent, err := BuildConfig(ctx, in)
	if err != nil {
		return "", err
	}
	res, err := parsedContent.MarshalJSONContext(ctx)
	return string(res), err
}

func BuildConfig(ctx context.Context, in *StartRequest) (*option.Options, error) {
	Log(LogLevel_DEBUG, LogType_CORE, "Building Config...")

	readOpt := &config.ReadOptions{Content: in.ConfigContent, Path: in.ConfigPath}
	if !in.EnableRawConfig {
		// hcontent, err := json.MarshalIndent(static.HiddifyOptions, "", " ")
		// if err != nil {
		// 	return nil, err
		// }

		// Log(LogLevel_DEBUG, LogType_CORE, "Building config ", string(hcontent))
		// Log(LogLevel_DEBUG, LogType_CORE, "Building config ")
		return config.BuildConfig(ctx, static.HiddifyOptions, readOpt)
	}
	return config.ReadSingOptions(ctx, readOpt)

}

func (s *CoreService) Parse(ctx context.Context, in *ParseRequest) (*ParseResponse, error) {
	return Parse(libbox.FromContext(ctx, nil), in)
}

func Parse(ctx context.Context, in *ParseRequest) (*ParseResponse, error) {
	defer config.DeferPanicToError("parse", func(err error) {
		Log(LogLevel_FATAL, LogType_CONFIG, err.Error())
		StopAndAlert(MessageType_UNEXPECTED_ERROR, err.Error())
	})

	path := in.TempPath
	if path == "" {
		path = in.ConfigPath
	}

	parsedConfig, err := config.ParseConfigBytes(ctx, &config.ReadOptions{Content: in.Content, Path: path}, true, static.HiddifyOptions, false)
	if err != nil {
		return &ParseResponse{
			ResponseCode: hcommon.ResponseCode_FAILED,
			Message:      err.Error(),
		}, err
	}
	if in.ConfigPath != "" {
		err = os.WriteFile(in.ConfigPath, parsedConfig, 0o644)
		if err != nil {
			return &ParseResponse{
				ResponseCode: hcommon.ResponseCode_FAILED,
				Message:      err.Error(),
			}, err
		}
		if activeOpt := config.DetectOLCRTCOptions(in.Content, path); activeOpt != nil {
			_ = config.SaveOLCRTCOptions(in.ConfigPath, activeOpt)
		}
	}
	return &ParseResponse{
		ResponseCode: hcommon.ResponseCode_OK,
		Content:      string(parsedConfig),
		Message:      "",
	}, err
}

func (s *CoreService) ChangeHiddifySettings(ctx context.Context, in *ChangeHiddifySettingsRequest) (*CoreInfoResponse, error) {
	return ChangeHiddifySettings(in, true)
}

func ChangeHiddifySettings(in *ChangeHiddifySettingsRequest, insert bool) (*CoreInfoResponse, error) {
	static.HiddifyOptions = config.DefaultHiddifyOptions()
	defer func() {
		switch static.HiddifyOptions.LogLevel {
		case "debug":
			static.logLevel = LogLevel_DEBUG
		case "info":
			static.logLevel = LogLevel_INFO
		case "warn":
			static.logLevel = LogLevel_WARNING
		case "error":
			static.logLevel = LogLevel_ERROR
		case "fatal":
			static.logLevel = LogLevel_FATAL
		case "trace":
			static.logLevel = LogLevel_TRACE
		default:
			static.logLevel = LogLevel_INFO
		}
		static.debug = static.debug || static.logLevel <= LogLevel_DEBUG
	}()

	if in.HiddifySettingsJson == "" {
		return &CoreInfoResponse{}, nil
	}
	if insert {
		settings := db.GetTable[hcommon.AppSettings]()
		settings.UpdateInsert(&hcommon.AppSettings{
			Id:    "HiddifySettingsJson",
			Value: in.HiddifySettingsJson,
		})
	}

	err := json.Unmarshal([]byte(in.HiddifySettingsJson), static.HiddifyOptions)
	if err != nil {
		return nil, err
	}

	// Map chain-status / unblocker / extra-security into Warp options if configured
	if (static.HiddifyOptions.ChainStatus == "unblocker" || static.HiddifyOptions.ChainStatus == "both" || static.HiddifyOptions.Unblocker.Mode != "") && static.HiddifyOptions.Unblocker.Mode == "warp" {
		static.HiddifyOptions.Warp.EnableWarp = true
		static.HiddifyOptions.Warp.Mode = "warp_over_proxy"
		if static.HiddifyOptions.Unblocker.Warp.CleanIP != "" {
			static.HiddifyOptions.Warp.CleanIP = static.HiddifyOptions.Unblocker.Warp.CleanIP
		}
		if static.HiddifyOptions.Unblocker.Warp.CleanPort != 0 {
			static.HiddifyOptions.Warp.CleanPort = uint16(static.HiddifyOptions.Unblocker.Warp.CleanPort)
		}
		static.HiddifyOptions.Warp.FakePackets = static.HiddifyOptions.Unblocker.Warp.Noise
		static.HiddifyOptions.Warp.FakePacketMode = static.HiddifyOptions.Unblocker.Warp.NoiseMode
		static.HiddifyOptions.Warp.FakePacketSize = static.HiddifyOptions.Unblocker.Warp.NoiseSize
		static.HiddifyOptions.Warp.FakePacketDelay = static.HiddifyOptions.Unblocker.Warp.NoiseDelay
	} else if (static.HiddifyOptions.ChainStatus == "extra_security" || static.HiddifyOptions.ChainStatus == "both" || static.HiddifyOptions.ExtraSecurity.Mode != "") && static.HiddifyOptions.ExtraSecurity.Mode == "warp" {
		static.HiddifyOptions.Warp.EnableWarp = true
		static.HiddifyOptions.Warp.Mode = "proxy_over_warp"
	} else if static.HiddifyOptions.ChainStatus == "off" {
		static.HiddifyOptions.Warp.EnableWarp = false
	}

	if static.HiddifyOptions.Warp.WireguardConfigStr != "" {
		err := json.Unmarshal([]byte(static.HiddifyOptions.Warp.WireguardConfigStr), &static.HiddifyOptions.Warp.WireguardConfig)
		if err != nil {
			return nil, err
		}
	}
	if static.HiddifyOptions.Warp2.WireguardConfigStr != "" {
		err := json.Unmarshal([]byte(static.HiddifyOptions.Warp2.WireguardConfigStr), &static.HiddifyOptions.Warp2.WireguardConfig)
		if err != nil {
			return nil, err
		}
	}
	return &CoreInfoResponse{}, nil
}

func (s *CoreService) GenerateConfig(ctx context.Context, in *GenerateConfigRequest) (*GenerateConfigResponse, error) {
	return GenerateConfig(libbox.FromContext(ctx, nil), in)
}

func GenerateConfig(ctx context.Context, in *GenerateConfigRequest) (*GenerateConfigResponse, error) {
	defer config.DeferPanicToError("generateConfig", func(err error) {
		Log(LogLevel_FATAL, LogType_CONFIG, err.Error())
		StopAndAlert(MessageType_UNEXPECTED_ERROR, err.Error())
	})
	if static.HiddifyOptions == nil {
		static.HiddifyOptions = config.DefaultHiddifyOptions()
	}
	config, err := config.ParseBuildConfigBytes(ctx, static.HiddifyOptions, &config.ReadOptions{Path: in.Path})
	if err != nil {
		return nil, err
	}

	return &GenerateConfigResponse{
		ConfigContent: string(config),
	}, nil
}

func removeTunnelIfNeeded(options *option.Options) (tuninb *option.TunInboundOptions) {
	if hutils.TunAllowed() {
		return nil
	}

	// Create a new slice to hold the remaining inbounds
	newInbounds := make([]option.Inbound, 0, len(options.Inbounds))

	for _, inb := range options.Inbounds {
		if inb.Type == C.TypeTun {
			if d, ok := inb.Options.(option.TunInboundOptions); ok {
				tuninb = &d
			}

		} else {
			newInbounds = append(newInbounds, inb)
		}
	}

	options.Inbounds = newInbounds
	return tuninb
}
