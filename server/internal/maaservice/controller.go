package maaservice

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/win32"
	"github.com/rs/zerolog"

	"github.com/MaaXYZ/MaaDebugger/internal/logger"
)

var maaServiceLog = logger.For(logger.ComponentMaaService)

type ControllerType string

const (
	ADB       ControllerType = "adb"
	Win32     ControllerType = "win32"
	Gamepad   ControllerType = "gamepad"
	PlayCover ControllerType = "playcover"
	WlRoot    ControllerType = "wlroot"
	Custom    ControllerType = "custom"
)

// ControllerService 管理 MaaFW Controller 实例的生命周期。
type ControllerService struct {
	controller     atomic.Pointer[maa.Controller]
	controllerType ControllerType
}

// NewControllerService 创建一个新的 ControllerService。
func NewControllerService() *ControllerService {
	return &ControllerService{}
}

// ConnectControllerResult 表示连接结果。
type ConnectControllerResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ConnectAdb 连接 ADB 设备。
func (s *ControllerService) ConnectAdb(
	adbPath, address, screencapMethod, inputMethod, config string,
) ConnectControllerResult {
	maaServiceLog.Info().
		Str("adb_path", adbPath).
		Str("address", address).
		Str("screencap_method", screencapMethod).
		Str("input_method", inputMethod).
		Str("config", config).
		Msg("connect adb request")

	scMethod, err := adb.ParseScreencapMethod(screencapMethod)
	if err != nil {
		maaServiceLog.Warn().Err(err).Str("raw", screencapMethod).Msg("parse adb screencap method failed, using raw uint64")
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", screencapMethod).Msg("invalid screencap method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = adb.ScreencapMethod(v)
	}
	maaServiceLog.Debug().Uint64("screencap_method_parsed", uint64(scMethod)).Msg("screencap method resolved")

	inMethod, err := adb.ParseInputMethod(inputMethod)
	if err != nil {
		maaServiceLog.Warn().Err(err).Str("raw", inputMethod).Msg("parse adb input method failed, using raw uint64")
		v, parseErr := strconv.ParseUint(inputMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", inputMethod).Msg("invalid input method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid input method: %s", inputMethod)}
		}
		inMethod = adb.InputMethod(v)
	}
	maaServiceLog.Debug().Uint64("input_method_parsed", uint64(inMethod)).Msg("input method resolved")

	maaServiceLog.Info().Msg("creating adb controller")
	ctrl, err := maa.NewAdbController(adbPath, address, scMethod, inMethod, config, "")
	if err != nil {
		maaServiceLog.Error().Err(err).Str("address", address).Msg("create adb controller failed")
		return ConnectControllerResult{Error: fmt.Sprintf("create adb controller failed: %v", err)}
	}

	maaServiceLog.Info().Msg("adb controller created, posting connection")
	ctrl.PostConnect().Wait()

	connected := ctrl.Connected()
	maaServiceLog.Info().Bool("connected", connected).Msg("adb post-connect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect ADB: %s", address)
		maaServiceLog.Warn().Str("address", address).Msg("adb connect returned disconnected")
		return ConnectControllerResult{Error: errMsg}
	}

	return s.finishConnect(ctrl, ADB, maaServiceLog.Info().Str("address", address), "adb controller connected")
}

// ConnectWin32 连接 Win32 控制器。
func (s *ControllerService) ConnectWin32(
	hwndStr, screencapMethod, mouseMethod, keyboardMethod string,
) ConnectControllerResult {
	maaServiceLog.Info().
		Str("hwnd", hwndStr).
		Str("screencap_method", screencapMethod).
		Str("mouse_method", mouseMethod).
		Str("keyboard_method", keyboardMethod).
		Msg("connect win32 request")

	hwnd, err := parseHwnd(hwndStr)
	if err != nil {
		maaServiceLog.Error().Err(err).Str("hwnd", hwndStr).Msg("invalid hwnd")
		return ConnectControllerResult{Error: fmt.Sprintf("invalid hwnd: %s", hwndStr)}
	}

	scMethod, err := win32.ParseScreencapMethod(screencapMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", screencapMethod).Msg("invalid win32 screencap method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = win32.ScreencapMethod(v)
	}

	mouseM, err := win32.ParseInputMethod(mouseMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(mouseMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", mouseMethod).Msg("invalid win32 mouse method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid mouse method: %s", mouseMethod)}
		}
		mouseM = win32.InputMethod(v)
	}

	keyboardM, err := win32.ParseInputMethod(keyboardMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(keyboardMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", keyboardMethod).Msg("invalid win32 keyboard method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid keyboard method: %s", keyboardMethod)}
		}
		keyboardM = win32.InputMethod(v)
	}

	maaServiceLog.Info().Msg("creating win32 controller")
	ctrl, err := maa.NewWin32Controller(hwnd, scMethod, mouseM, keyboardM)
	if err != nil {
		maaServiceLog.Error().Err(err).Str("hwnd", hwndStr).Msg("create win32 controller failed")
		return ConnectControllerResult{Error: fmt.Sprintf("create win32 controller failed: %v", err)}
	}

	maaServiceLog.Info().Msg("win32 controller created, posting connection")
	ctrl.PostConnect().Wait()

	connected := ctrl.Connected()
	maaServiceLog.Info().Bool("connected", connected).Msg("win32 post-connect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect Win32 hwnd: %s", hwndStr)
		maaServiceLog.Warn().Str("hwnd", hwndStr).Msg("win32 connect returned disconnected")
		return ConnectControllerResult{Error: errMsg}
	}

	return s.finishConnect(ctrl, Win32, maaServiceLog.Info().Str("hwnd", hwndStr), "win32 controller connected")
}

// ConnectGamepad 连接 Gamepad 控制器。
func (s *ControllerService) ConnectGamepad(
	hwndStr, screencapMethod, gamepadTypeStr string,
) ConnectControllerResult {
	maaServiceLog.Info().
		Str("hwnd", hwndStr).
		Str("screencap_method", screencapMethod).
		Str("gamepad_type", gamepadTypeStr).
		Msg("connect gamepad request")

	hwnd, err := parseHwnd(hwndStr)
	if err != nil {
		maaServiceLog.Error().Err(err).Str("hwnd", hwndStr).Msg("invalid hwnd")
		return ConnectControllerResult{Error: fmt.Sprintf("invalid hwnd: %s", hwndStr)}
	}

	scMethod, err := win32.ParseScreencapMethod(screencapMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			maaServiceLog.Error().Err(parseErr).Str("value", screencapMethod).Msg("invalid gamepad screencap method")
			return ConnectControllerResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = win32.ScreencapMethod(v)
	}

	gamepadType, err := strconv.ParseInt(gamepadTypeStr, 10, 32)
	if err != nil {
		maaServiceLog.Error().Err(err).Str("value", gamepadTypeStr).Msg("invalid gamepad type")
		return ConnectControllerResult{Error: fmt.Sprintf("invalid gamepad type: %s", gamepadTypeStr)}
	}

	maaServiceLog.Info().Msg("creating gamepad controller")
	ctrl, err := maa.NewGamepadController(hwnd, maa.GamepadType(int32(gamepadType)), scMethod)
	if err != nil {
		maaServiceLog.Error().Err(err).Str("hwnd", hwndStr).Msg("create gamepad controller failed")
		return ConnectControllerResult{Error: fmt.Sprintf("create gamepad controller failed: %v", err)}
	}

	maaServiceLog.Info().Msg("gamepad controller created, posting connection")
	ctrl.PostConnect().Wait()

	connected := ctrl.Connected()
	maaServiceLog.Info().Bool("connected", connected).Msg("gamepad post-connect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect Gamepad hwnd: %s", hwndStr)
		maaServiceLog.Warn().Str("hwnd", hwndStr).Msg("gamepad connect returned disconnected")
		return ConnectControllerResult{Error: errMsg}
	}

	return s.finishConnect(ctrl, Gamepad, maaServiceLog.Info().Str("hwnd", hwndStr), "gamepad controller connected")
}

// ConnectPlayCover 连接 PlayCover 控制器。
func (s *ControllerService) ConnectPlayCover(
	address, uuid string,
) ConnectControllerResult {
	maaServiceLog.Info().
		Str("address", address).
		Str("uuid", uuid).
		Msg("connect playcover request")

	maaServiceLog.Info().Msg("creating playcover controller")
	ctrl, err := maa.NewPlayCoverController(address, uuid)
	if err != nil {
		maaServiceLog.Error().Err(err).
			Str("address", address).
			Str("uuid", uuid).
			Msg("create playcover controller failed")
		return ConnectControllerResult{Error: fmt.Sprintf("create PlayCover controller failed: %v", err)}
	}

	maaServiceLog.Info().Msg("playcover controller created, posting connection")
	ctrl.PostConnect().Wait()

	connected := ctrl.Connected()
	maaServiceLog.Info().Bool("connected", connected).Msg("playcover post-connect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect PlayCover: %s", address)
		maaServiceLog.Warn().Str("address", address).Msg("playcover connect returned disconnected")
		return ConnectControllerResult{Error: errMsg}
	}

	return s.finishConnect(ctrl, PlayCover, maaServiceLog.Info().Str("address", address).Str("uuid", uuid), "playcover controller connected")
}

func (s *ControllerService) ConnectWlRoot(wlrSocketPath string) ConnectControllerResult {
	wlrSocketPath = strings.TrimSpace(wlrSocketPath)
	if wlrSocketPath == "" {
		return ConnectControllerResult{Error: "socket path is required"}
	}

	maaServiceLog.Info().Str("socket_path", wlrSocketPath).Msg("connect wlroot request")

	maaServiceLog.Info().Msg("creating wlroot controller")
	ctrl, err := maa.NewWlRootsController(wlrSocketPath, false) // TODO: provide `useWin32VkCode` option
	if err != nil {
		maaServiceLog.Error().Err(err).Str("socket_path", wlrSocketPath).Msg("create wlroot controller failed")
		return ConnectControllerResult{Error: fmt.Sprintf("create WlRoot controller failed: %v", err)}
	}

	maaServiceLog.Info().Msg("wlroot controller created, posting connection")
	ctrl.PostConnect().Wait()

	connected := ctrl.Connected()
	maaServiceLog.Info().Bool("connected", connected).Msg("wlroot post-connect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect WlRoot: %s", wlrSocketPath)
		maaServiceLog.Warn().Str("socket_path", wlrSocketPath).Msg("wlroot connect returned disconnected")
		return ConnectControllerResult{Error: errMsg}
	}

	return s.finishConnect(ctrl, WlRoot, maaServiceLog.Info().Str("socket_path", wlrSocketPath), "wlroot controller connected")
}

func (s *ControllerService) finishConnect(
	ctrl *maa.Controller,
	controllerType ControllerType,
	event *zerolog.Event,
	successMsg string,
) ConnectControllerResult {
	if old := s.controller.Swap(ctrl); old != nil {
		maaServiceLog.Info().Str("type", string(controllerType)).Msg("destroying previous controller")
		old.Destroy()
	}

	s.controllerType = controllerType
	event.Msg(successMsg)
	return ConnectControllerResult{Success: true}
}

// Disconnect 断开当前 Controller 连接。
func (s *ControllerService) Disconnect() {
	if old := s.controller.Swap(nil); old != nil {
		maaServiceLog.Info().Str("type", string(s.controllerType)).Msg("disconnecting controller")
		old.Destroy()
		s.controllerType = ""
		maaServiceLog.Info().Msg("controller disconnected")
	} else {
		s.controllerType = ""
		maaServiceLog.Debug().Msg("disconnect called with no active controller")
	}
}

// Connected 返回当前是否已连接。
func (s *ControllerService) Connected() bool {
	ctrl := s.controller.Load()
	if ctrl == nil {
		return false
	}
	return ctrl.Connected()
}

// Controller 返回当前的 Controller 实例（可能为 nil）。
func (s *ControllerService) Controller() *maa.Controller {
	return s.controller.Load()
}

// ControllerType 返回当前 Controller 的类型
func (s *ControllerService) ControllerType() ControllerType {
	return s.controllerType
}

func (s *ControllerService) ControllerMethod(method MethodType) []ControllerMethod {
	switch method {
	case ADBScreencap:
		return adbScreencapMethods
	case ADBInput:
		return adbInputMethods
	case WindowScreencap:
		return windowScreencapMethods
	case Win32Input:
		return win32InputMethods
	case GamepadInput:
		return gamepadInputMethods
	}

	return []ControllerMethod{}
}

// parseHwnd 将 hwnd 字符串解析为 unsafe.Pointer。
// 支持十进制（如 "12345"）和十六进制（如 "0x12345"）格式。
func parseHwnd(s string) (unsafe.Pointer, error) {
	var v uint64
	var err error

	if len(s) > 2 && (s[:2] == "0x" || s[:2] == "0X") {
		v, err = strconv.ParseUint(s[2:], 16, 64)
	} else {
		v, err = strconv.ParseUint(s, 10, 64)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid hwnd %q: %w", s, err)
	}
	return unsafe.Pointer(uintptr(v)), nil
}
