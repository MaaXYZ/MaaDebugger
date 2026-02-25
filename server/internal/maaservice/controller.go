package maaservice

import (
	"fmt"
	"strconv"
	"sync"
	"unsafe"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/win32"
	"github.com/rs/zerolog/log"
)

// ControllerService 管理 MaaFW Controller 实例的生命周期。
type ControllerService struct {
	mu         sync.Mutex
	controller *maa.Controller
}

// NewControllerService 创建一个新的 ControllerService。
func NewControllerService() *ControllerService {
	return &ControllerService{}
}

// ConnectAdbResult 表示 ADB 连接结果。
type ConnectAdbResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ConnectAdb 连接 ADB 设备。
func (s *ControllerService) ConnectAdb(
	adbPath, address, screencapMethod, inputMethod, config string,
) ConnectAdbResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().
		Str("adb_path", adbPath).
		Str("address", address).
		Str("screencap_method", screencapMethod).
		Str("input_method", inputMethod).
		Str("config", config).
		Msg("[MaaService] ConnectAdb called")

	// 销毁旧连接
	if s.controller != nil {
		log.Info().Msg("[MaaService] destroying previous controller")
		s.controller.Destroy()
		s.controller = nil
	}

	scMethod, err := adb.ParseScreencapMethod(screencapMethod)
	if err != nil {
		log.Warn().Err(err).Str("raw", screencapMethod).Msg("[MaaService] parse adb screencap method failed, using raw uint64")
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", screencapMethod).Msg("[MaaService] invalid screencap method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = adb.ScreencapMethod(v)
	}
	log.Debug().Uint64("screencap_method_parsed", uint64(scMethod)).Msg("[MaaService] screencap method resolved")

	inMethod, err := adb.ParseInputMethod(inputMethod)
	if err != nil {
		log.Warn().Err(err).Str("raw", inputMethod).Msg("[MaaService] parse adb input method failed, using raw uint64")
		v, parseErr := strconv.ParseUint(inputMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", inputMethod).Msg("[MaaService] invalid input method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid input method: %s", inputMethod)}
		}
		inMethod = adb.InputMethod(v)
	}
	log.Debug().Uint64("input_method_parsed", uint64(inMethod)).Msg("[MaaService] input method resolved")

	log.Info().Msg("[MaaService] creating ADB controller...")
	ctrl, err := maa.NewAdbController(adbPath, address, scMethod, inMethod, config, "")
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("[MaaService] create adb controller failed")
		return ConnectAdbResult{Error: fmt.Sprintf("create adb controller failed: %v", err)}
	}

	log.Info().Msg("[MaaService] ADB controller created, posting connection...")
	job := ctrl.PostConnect()
	job.Wait()

	connected := ctrl.Connected()
	log.Info().Bool("connected", connected).Msg("[MaaService] ADB PostConnect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect ADB: %s", address)
		log.Warn().Str("address", address).Msg("[MaaService] " + errMsg)
		return ConnectAdbResult{Error: errMsg}
	}

	s.controller = ctrl
	log.Info().Str("address", address).Msg("[MaaService] ADB controller connected successfully")
	return ConnectAdbResult{Success: true}
}

// ConnectWin32 连接 Win32 控制器。
func (s *ControllerService) ConnectWin32(
	hwndStr, screencapMethod, mouseMethod, keyboardMethod string,
) ConnectAdbResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().
		Str("hwnd", hwndStr).
		Str("screencap_method", screencapMethod).
		Str("mouse_method", mouseMethod).
		Str("keyboard_method", keyboardMethod).
		Msg("[MaaService] ConnectWin32 called")

	// 销毁旧连接
	if s.controller != nil {
		log.Info().Msg("[MaaService] destroying previous controller")
		s.controller.Destroy()
		s.controller = nil
	}

	hwnd, err := parseHwnd(hwndStr)
	if err != nil {
		log.Error().Err(err).Str("hwnd", hwndStr).Msg("[MaaService] invalid hwnd")
		return ConnectAdbResult{Error: fmt.Sprintf("invalid hwnd: %s", hwndStr)}
	}

	scMethod, err := win32.ParseScreencapMethod(screencapMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", screencapMethod).Msg("[MaaService] invalid win32 screencap method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = win32.ScreencapMethod(v)
	}

	mouseM, err := win32.ParseInputMethod(mouseMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(mouseMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", mouseMethod).Msg("[MaaService] invalid win32 mouse method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid mouse method: %s", mouseMethod)}
		}
		mouseM = win32.InputMethod(v)
	}

	keyboardM, err := win32.ParseInputMethod(keyboardMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(keyboardMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", keyboardMethod).Msg("[MaaService] invalid win32 keyboard method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid keyboard method: %s", keyboardMethod)}
		}
		keyboardM = win32.InputMethod(v)
	}

	log.Info().Msg("[MaaService] creating Win32 controller...")
	ctrl, err := maa.NewWin32Controller(hwnd, scMethod, mouseM, keyboardM)
	if err != nil {
		log.Error().Err(err).Str("hwnd", hwndStr).Msg("[MaaService] create win32 controller failed")
		return ConnectAdbResult{Error: fmt.Sprintf("create win32 controller failed: %v", err)}
	}

	log.Info().Msg("[MaaService] Win32 controller created, posting connection...")
	job := ctrl.PostConnect()
	job.Wait()

	connected := ctrl.Connected()
	log.Info().Bool("connected", connected).Msg("[MaaService] Win32 PostConnect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect Win32 hwnd: %s", hwndStr)
		log.Warn().Str("hwnd", hwndStr).Msg("[MaaService] " + errMsg)
		return ConnectAdbResult{Error: errMsg}
	}

	s.controller = ctrl
	log.Info().Str("hwnd", hwndStr).Msg("[MaaService] Win32 controller connected successfully")
	return ConnectAdbResult{Success: true}
}

// ConnectGamepad 连接 Gamepad 控制器。
func (s *ControllerService) ConnectGamepad(
	hwndStr, screencapMethod, gamepadTypeStr string,
) ConnectAdbResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().
		Str("hwnd", hwndStr).
		Str("screencap_method", screencapMethod).
		Str("gamepad_type", gamepadTypeStr).
		Msg("[MaaService] ConnectGamepad called")

	// 销毁旧连接
	if s.controller != nil {
		log.Info().Msg("[MaaService] destroying previous controller")
		s.controller.Destroy()
		s.controller = nil
	}

	hwnd, err := parseHwnd(hwndStr)
	if err != nil {
		log.Error().Err(err).Str("hwnd", hwndStr).Msg("[MaaService] invalid hwnd")
		return ConnectAdbResult{Error: fmt.Sprintf("invalid hwnd: %s", hwndStr)}
	}

	scMethod, err := win32.ParseScreencapMethod(screencapMethod)
	if err != nil {
		v, parseErr := strconv.ParseUint(screencapMethod, 10, 64)
		if parseErr != nil {
			log.Error().Err(parseErr).Str("value", screencapMethod).Msg("[MaaService] invalid gamepad screencap method")
			return ConnectAdbResult{Error: fmt.Sprintf("invalid screencap method: %s", screencapMethod)}
		}
		scMethod = win32.ScreencapMethod(v)
	}

	gamepadType, err := strconv.ParseInt(gamepadTypeStr, 10, 32)
	if err != nil {
		log.Error().Err(err).Str("value", gamepadTypeStr).Msg("[MaaService] invalid gamepad type")
		return ConnectAdbResult{Error: fmt.Sprintf("invalid gamepad type: %s", gamepadTypeStr)}
	}

	log.Info().Msg("[MaaService] creating Gamepad controller...")
	ctrl, err := maa.NewGamepadController(hwnd, maa.GamepadType(int32(gamepadType)), scMethod)
	if err != nil {
		log.Error().Err(err).Str("hwnd", hwndStr).Msg("[MaaService] create gamepad controller failed")
		return ConnectAdbResult{Error: fmt.Sprintf("create gamepad controller failed: %v", err)}
	}

	log.Info().Msg("[MaaService] Gamepad controller created, posting connection...")
	job := ctrl.PostConnect()
	job.Wait()

	connected := ctrl.Connected()
	log.Info().Bool("connected", connected).Msg("[MaaService] Gamepad PostConnect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect Gamepad hwnd: %s", hwndStr)
		log.Warn().Str("hwnd", hwndStr).Msg("[MaaService] " + errMsg)
		return ConnectAdbResult{Error: errMsg}
	}

	s.controller = ctrl
	log.Info().Str("hwnd", hwndStr).Msg("[MaaService] Gamepad controller connected successfully")
	return ConnectAdbResult{Success: true}
}

// Disconnect 断开当前 Controller 连接。
func (s *ControllerService) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.controller != nil {
		log.Info().Msg("[MaaService] disconnecting controller...")
		s.controller.Destroy()
		s.controller = nil
		log.Info().Msg("[MaaService] controller disconnected")
	} else {
		log.Info().Msg("[MaaService] disconnect called but no active controller")
	}
}

// ConnectPlayCover 连接 PlayCover 控制器。
func (s *ControllerService) ConnectPlayCover(
	address, uuid string,
) ConnectAdbResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().
		Str("address", address).
		Str("uuid", uuid).
		Msg("[MaaService] ConnectPlayCover called")

	// 销毁旧连接
	if s.controller != nil {
		log.Info().Msg("[MaaService] destroying previous controller")
		s.controller.Destroy()
		s.controller = nil
	}

	log.Info().Msg("[MaaService] creating PlayCover controller...")
	ctrl, err := maa.NewPlayCoverController(address, uuid)
	if err != nil {
		log.Error().Err(err).
			Str("address", address).
			Str("uuid", uuid).
			Msg("[MaaService] create PlayCover controller failed")
		return ConnectAdbResult{Error: fmt.Sprintf("create PlayCover controller failed: %v", err)}
	}

	log.Info().Msg("[MaaService] PlayCover controller created, posting connection...")
	job := ctrl.PostConnect()
	job.Wait()

	connected := ctrl.Connected()
	log.Info().Bool("connected", connected).Msg("[MaaService] PlayCover PostConnect completed")

	if !connected {
		ctrl.Destroy()
		errMsg := fmt.Sprintf("failed to connect PlayCover: %s", address)
		log.Warn().Str("address", address).Msg("[MaaService] " + errMsg)
		return ConnectAdbResult{Error: errMsg}
	}

	s.controller = ctrl
	log.Info().Str("address", address).Str("uuid", uuid).Msg("[MaaService] PlayCover controller connected successfully")
	return ConnectAdbResult{Success: true}
}

// Connected 返回当前是否已连接。
func (s *ControllerService) Connected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.controller == nil {
		return false
	}
	return s.controller.Connected()
}

// Controller 返回当前的 Controller 实例（可能为 nil）。
func (s *ControllerService) Controller() *maa.Controller {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.controller
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
