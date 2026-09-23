package mobile

import (
	hcore "github.com/hiddify/hiddify-core/v2/hcore"
	olcrtc "github.com/openlibrecommunity/olcrtc/mobile"

	_ "net/http/pprof"

	_ "github.com/sagernet/gomobile"
	"github.com/sagernet/sing-box/experimental/libbox"
)

type SetupOptions struct {
	BasePath        string
	WorkingDir      string
	TempDir         string
	Listen          string
	Secret          string
	Debug           bool
	Mode            int
	FixAndroidStack bool
}

type olcrtcProtector struct {
	pi libbox.PlatformInterface
}

func (p *olcrtcProtector) Protect(fd int) bool {
	if p.pi != nil {
		return p.pi.AutoDetectInterfaceControl(int32(fd)) == nil
	}
	return false
}

func Setup(opt *SetupOptions, platformInterface libbox.PlatformInterface) error {
	if platformInterface != nil {
		olcrtc.SetProtector(&olcrtcProtector{pi: platformInterface})
	}
	olcrtc.SetProviders()

	return hcore.Setup(&hcore.SetupRequest{
		BasePath:          opt.BasePath,
		WorkingDir:        opt.WorkingDir,
		TempDir:           opt.TempDir,
		FlutterStatusPort: 0,
		Listen:            opt.Listen,
		Debug:             opt.Debug,
		Mode:              hcore.SetupMode(opt.Mode),
		Secret:            opt.Secret,
		FixAndroidStack:   opt.FixAndroidStack,
	}, platformInterface)
}

func Start(configPath string, configContent string) error {
	_, err := hcore.StartService(libbox.BaseContext(nil), &hcore.StartRequest{
		ConfigPath:    configPath,
		ConfigContent: configContent,
	})
	return err
}

func Stop() error {
	if olcrtc.IsRunning() {
		olcrtc.Stop()
	}
	_, err := hcore.Stop()
	return err
}

func GetServerPublicKey() []byte {
	return hcore.GetGrpcServerPublicKey()
}

func AddGrpcClientPublicKey(clientPublicKey []byte) error {
	return hcore.AddGrpcClientPublicKey(clientPublicKey)
}

func Close(mode int) {
	if olcrtc.IsRunning() {
		olcrtc.Stop()
	}
	hcore.Close(hcore.SetupMode(mode))
}

func Test() string {
	return "Hello from mobile with olcRTC"
}

func Pause() {
	hcore.Pause()
}

func Wake() {
	hcore.Wake()
}

// olcRTC specific mobile functions

func StartOLCRTC(carrierName, transportName, roomID, clientID, keyHex string, socksPort int, socksUser, socksPass string) error {
	return olcrtc.StartWithTransport(carrierName, transportName, roomID, clientID, keyHex, socksPort, socksUser, socksPass)
}

func StopOLCRTC() error {
	olcrtc.Stop()
	return nil
}

func WaitOLCRTCReady(timeoutMillis int64) error {
	return olcrtc.WaitReady(int(timeoutMillis))
}

func IsOLCRTCRunning() bool {
	return olcrtc.IsRunning()
}

func SetOLCRTCDNS(dnsServer string) {
	olcrtc.SetDNS(dnsServer)
}

func SetOLCRTCWBToken(token string) {
	olcrtc.SetWBToken(token)
}

func SetOLCRTCVP8Options(fps, batchSize int) {
	olcrtc.SetVP8Options(fps, batchSize)
}

func SetOLCRTCLivenessOptions(intervalMillis, timeoutMillis, failures int) {
	olcrtc.SetLivenessOptions(intervalMillis, timeoutMillis, failures)
}
