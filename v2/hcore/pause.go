package hcore

import (
	"context"
	"fmt"
	"time"

	hcommon "github.com/hiddify/hiddify-core/v2/hcommon"
	C "github.com/sagernet/sing-box/constant"
)

func (s *CoreService) Close(ctx context.Context, closeReq *CloseRequest) (*hcommon.Empty, error) {
	if closeReq == nil {
		return nil, nil
	}
	mode := closeReq.Mode
	if grpcServer[mode] == nil {
		Log(LogLevel_WARNING, LogType_CORE, "grpcServer already stoped")
		return nil, nil
	}

	CloseGrpcServer(mode)
	return &hcommon.Empty{}, nil
}

func Pause() {
	defer func() {
		if r := recover(); r != nil {
			Log(LogLevel_ERROR, LogType_CORE, fmt.Sprintf("Pause recovered panic: %v", r))
		}
	}()
	if box := static.Instance(); box != nil {
		if manager := box.PauseManager(); manager != nil {
			if !manager.IsDevicePaused() {
				manager.DevicePause()
			}
			if C.IsIos {
				if static.endPauseTimer == nil {
					static.endPauseTimer = time.AfterFunc(time.Minute, func() {
						Wake()
					})
				} else {
					static.endPauseTimer.Reset(time.Minute)
				}
			}
		}
	}
}

func Wake() {
	defer func() {
		if r := recover(); r != nil {
			Log(LogLevel_ERROR, LogType_CORE, fmt.Sprintf("Wake recovered panic: %v", r))
		}
	}()
	if box := static.Instance(); box != nil {
		if manager := box.PauseManager(); manager != nil {
			if manager.IsDevicePaused() {
				manager.DeviceWake()
			}
		}
	}
}
