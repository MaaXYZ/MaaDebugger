package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/gorilla/websocket"

	"github.com/MaaXYZ/MaaDebugger/frontend"
	"github.com/MaaXYZ/MaaDebugger/internal/buildinfo"
	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/logger"
	"github.com/MaaXYZ/MaaDebugger/internal/maaservice"
	"github.com/MaaXYZ/MaaDebugger/internal/platform"
	"github.com/MaaXYZ/MaaDebugger/internal/response"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/updater"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
)

var (
	httpLog       = logger.For(logger.ComponentHTTP)
	controllerLog = logger.For(logger.ComponentController)
	agentLog      = logger.For(logger.ComponentAgent)
	resourceLog   = logger.For(logger.ComponentResource)
	taskLog       = logger.For(logger.ComponentTask)
	interfaceLog  = logger.For(logger.ComponentInterface)
)

type Dependencies struct {
	StatusStore       *state.Store
	Hub               *ws.Hub
	ControllerService *maaservice.ControllerService
	ResourceService   *maaservice.ResourceService
	TaskerService     *maaservice.TaskerService
	AgentService      *maaservice.AgentService
	ScreenshotService *maaservice.ScreenshotService
	ConfigStore       *configstore.Store
	Channel           string
}

type router struct {
	deps     Dependencies
	upgrader websocket.Upgrader
}

func NewRouter(deps Dependencies) http.Handler {
	r := &router{
		deps: deps,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}

	// 设置事件回调：将 Tasker 事件通过 WS 广播（不在 sink 中渲染）
	deps.TaskerService.SetEventCallback(func(msg map[string]any) {
		deps.Hub.BroadcastJSON(ws.Message{Type: "task.event", Payload: msg})
	})

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/info/all", r.handleMaaDebuggerInfo)
	mux.HandleFunc("GET /api/info/status", r.handleInfoStatus)
	mux.HandleFunc("GET /api/info/uac", r.handleSystemUAC)
	mux.HandleFunc("GET /api/fw/version", r.handleMaaFrameworkVersion)
	mux.HandleFunc("GET /api/channel", r.handleChannel)
	mux.HandleFunc("GET /api/locale", r.handleLocale)
	mux.HandleFunc("GET /api/config", r.handleConfigAll)
	mux.HandleFunc("GET /api/config/{key}", r.handleConfigGet)
	mux.HandleFunc("PUT /api/config/{key}", r.handleConfigSet)
	mux.HandleFunc("PUT /api/config", r.handleConfigMerge)
	mux.HandleFunc("GET /api/controller/detect/adb", r.handleDetectAdb)
	mux.HandleFunc("GET /api/controller/detect/desktop", r.handleDetectDesktop)
	mux.HandleFunc("POST /api/controller/connect", r.handleControllerConnect)
	mux.HandleFunc("POST /api/controller/disconnect", r.handleControllerDisconnect)
	mux.HandleFunc("GET /api/controller/methods", r.handleGetControllerMethods)
	mux.HandleFunc("POST /api/path/exists", r.handlePathExists)
	mux.HandleFunc("POST /api/interface/parse", r.handleInterfaceParse)
	mux.HandleFunc("POST /api/resource/load", r.handleResourceLoad)
	mux.HandleFunc("POST /api/pipeline/start", r.handlePipelineCheckerStart)
	mux.HandleFunc("POST /api/pipeline/stop", r.handlePipelineCheckerStop)
	mux.HandleFunc("POST /api/pipeline/check", r.handlePipelineCheckerCheck)
	mux.HandleFunc("GET /api/pipeline/result", r.handlePipelineCheckerGet)
	mux.HandleFunc("POST /api/task/run", r.handleTaskRun)
	mux.HandleFunc("POST /api/task/stop", r.handleTaskStop)
	mux.HandleFunc("GET /api/task/nodes", r.handleTaskNodes)
	mux.HandleFunc("GET /api/task/node/{name}", r.handleTaskNodeDetail)
	mux.HandleFunc("GET /api/task/node-data/{name}", r.handleTaskNodeData)
	mux.HandleFunc("GET /api/task/reco/{id}", r.handleTaskRecoDetail)
	mux.HandleFunc("GET /api/task/action/{id}", r.handleTaskActionDetail)
	mux.HandleFunc("GET /api/task/image/{id}", r.handleTaskImage)
	mux.HandleFunc("POST /api/agent/connect", r.handleAgentConnect)
	mux.HandleFunc("POST /api/agent/disconnect", r.handleAgentDisconnect)
	mux.HandleFunc("GET /api/agent/list", r.handleAgentList)
	mux.HandleFunc("POST /api/screenshot/start", r.handleScreenshotStart)
	mux.HandleFunc("POST /api/screenshot/stop", r.handleScreenshotStop)
	mux.HandleFunc("POST /api/screenshot/pause", r.handleScreenshotPause)
	mux.HandleFunc("POST /api/screenshot/resume", r.handleScreenshotResume)
	mux.HandleFunc("PUT /api/screenshot/fps", r.handleScreenshotSetFPS)
	mux.HandleFunc("PUT /api/screenshot/output", r.handleScreenshotSetOutput)
	mux.HandleFunc("GET /api/screenshot/status", r.handleScreenshotStatus)
	mux.HandleFunc("GET /api/screenshot/raw", r.handleScreenshotRaw)
	mux.HandleFunc("POST /api/clear/cache", r.handleClearCache)
	mux.HandleFunc("GET /api/update/check", r.handleCheckUpdate)
	mux.HandleFunc("GET /ws", r.handleWS)

	// Serve embedded frontend SPA for all non-API routes
	frontendHandler := frontend.Handler()
	mux.Handle("/", frontendHandler)

	return recoverer(logging(cors(mux)))
}
func (r *router) handleCheckUpdate(w http.ResponseWriter, req *http.Request) {
	showPre := updater.LoadIncludePreRelease(r.deps.ConfigStore)
	if raw := req.URL.Query().Get("showPre"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			response.Fail(w, http.StatusBadRequest, "invalid showPre query parameter")
			return
		}
		showPre = parsed
	}

	result, err := updater.CheckUpdate(updater.CheckOptions{
		Nightly:           updater.IsNightlyBuild(r.deps.Channel, buildinfo.Version),
		IncludePreRelease: showPre,
	})
	if err != nil {
		httpLog.Error().Err(err).Msg("check update failed")
		response.Fail(w, http.StatusInternalServerError, "check update failed: "+err.Error())
		return
	}

	response.OK(w, result)
}

func (r *router) handleMaaDebuggerInfo(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{
		"version":    buildinfo.Version,
		"commit_sha": buildinfo.CommitSHA,
		"build_time": buildinfo.BuildTime,
	})
}

func (r *router) handleMaaFrameworkVersion(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, maa.Version())
}

func (r *router) handleChannel(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, os.Getenv("MAADBG_CHANNEL"))
}

func (r *router) handleInfoStatus(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.StatusStore.Get())
}

func (r *router) handleConfigAll(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.ConfigStore.GetAll())
}

func (r *router) handleSystemUAC(w http.ResponseWriter, _ *http.Request) {
	enabled, err := platform.UACEnabled()
	if err != nil {
		httpLog.Error().Err(err).Msg("read windows UAC status failed")
		response.Fail(w, http.StatusInternalServerError, "read windows UAC status failed")
		return
	}

	response.OK(w, enabled)
}

func (r *router) handleConfigGet(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")

	v, ok := r.deps.ConfigStore.Get(key)
	if !ok {
		response.OK(w, nil)
		return
	}
	response.OK(w, v)
}

func (r *router) handleConfigSet(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	var payload json.RawMessage
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	var value any
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &value); err != nil {
			response.Fail(w, http.StatusBadRequest, "invalid json body")
			return
		}
	}

	r.deps.ConfigStore.Set(key, value)

	// 主动更新后端 cfg 值
	if key == "debugWorkspaceSettings" {
		if m, ok := value.(map[string]any); ok {
			if w, ok := m["watchResourceChange"].(bool); ok {
				r.deps.ResourceService.SetWatchEnabled(w)
			}
			if interval, ok := m["watchResourceChangeInterval"].(float64); ok {
				r.deps.ResourceService.SetWatchInterval(int(interval))
			}
		}
	}

	response.OK(w, nil)
}

func (r *router) handleConfigMerge(w http.ResponseWriter, req *http.Request) {
	var payload configMergePayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	entries := make(map[string]any, len(payload))
	for key, raw := range payload {
		if len(raw) == 0 {
			entries[key] = nil
			continue
		}
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			response.Fail(w, http.StatusBadRequest, fmt.Sprintf("invalid json body at key %q", key))
			return
		}
		entries[key] = value
	}

	r.deps.ConfigStore.Merge(entries)

	if v, ok := entries["debugWorkspaceSettings"]; ok {
		if m, ok := v.(map[string]any); ok {
			if w, ok := m["watchResourceChange"].(bool); ok {
				r.deps.ResourceService.SetWatchEnabled(w)
			}
			if interval, ok := m["watchResourceChangeInterval"].(float64); ok {
				r.deps.ResourceService.SetWatchInterval(int(interval))
			}
		}
	}

	response.OK(w, nil)
}

type configMergePayload map[string]json.RawMessage

type controllerConnectRequest struct {
	Type                 string `json:"type"`
	AdbPath              string `json:"adb_path"`
	AdbAddress           string `json:"adb_address"`
	AdbScreencapMethod   string `json:"adb_screencap_method"`
	AdbInputMethod       string `json:"adb_input_method"`
	AdbConfig            string `json:"adb_config"`
	Hwnd                 string `json:"hwnd"`
	Win32ScreencapMethod string `json:"win32_screencap_method"`
	Win32MouseMethod     string `json:"win32_mouse_method"`
	Win32KeyboardMethod  string `json:"win32_keyboard_method"`
	GamepadScreencap     string `json:"gamepad_screencap_method"`
	GamepadType          string `json:"gamepad_type"`
	PlayCoverAddress     string `json:"playcover_address"`
	PlayCoverUUID        string `json:"playcover_uuid"`
	WlrootSocketPath     string `json:"wlroot_socket_path"`
}

type controllerConnectResponse struct {
	Type string `json:"type"`
}

type taskRunRequest struct {
	Entry            string          `json:"entry"`
	PipelineOverride json.RawMessage `json:"pipeline_override"`
}

type taskCompletedPayload struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Entry   string `json:"entry,omitempty"`
	Stopped bool   `json:"stopped,omitempty"`
}

type screenshotSetFPSRequest struct {
	FPS int32 `json:"fps"`
}

type screenshotSetFPSResponse struct {
	FPS int32 `json:"fps"`
}

type screenshotSetOutputRequest struct {
	Output maaservice.ScreenshotOutput `json:"output"`
}

type screenshotOutputStatusResponse struct {
	Output maaservice.ScreenshotOutput `json:"output"`
	JPEG   bool                        `json:"jpeg"`
	H264   bool                        `json:"h264"`
	H265   bool                        `json:"h265"`
}

type screenshotStatusResponse struct {
	Running        bool                              `json:"running"`
	Paused         bool                              `json:"paused"`
	OutputActive   bool                              `json:"output_active"`
	FPS            int32                             `json:"fps"`
	Output         maaservice.ScreenshotOutput       `json:"output"`
	JPEG           bool                              `json:"jpeg"`
	H264           bool                              `json:"h264"`
	H265           bool                              `json:"h265"`
	OverlayState   maaservice.ScreenshotOverlayState `json:"overlay_state"`
	OverlayMessage string                            `json:"overlay_message"`
}

type adbDeviceInfo struct {
	Name             string `json:"name"`
	AdbPath          string `json:"adb_path"`
	Address          string `json:"address"`
	ScreencapMethods string `json:"screencap_methods"`
	InputMethods     string `json:"input_methods"`
	Config           string `json:"config"`
}

func (r *router) handleDetectAdb(w http.ResponseWriter, _ *http.Request) {
	controllerLog.Info().Msg("detect adb devices request")
	devices, err := maa.FindAdbDevices()
	if err != nil {
		controllerLog.Error().Err(err).Msg("find adb devices failed")
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("find adb devices failed: %v", err))
		return
	}

	result := make([]adbDeviceInfo, 0, len(devices))
	for _, d := range devices {
		if d == nil {
			continue
		}
		result = append(result, adbDeviceInfo{
			Name:             d.Name,
			AdbPath:          d.AdbPath,
			Address:          d.Address,
			ScreencapMethods: fmt.Sprintf("%v", d.ScreencapMethod),
			InputMethods:     fmt.Sprintf("%v", d.InputMethod),
			Config:           d.Config,
		})
	}

	controllerLog.Info().Int("count", len(result)).Msg("detect adb devices result")
	response.OK(w, result)
}

type desktopWindowInfo struct {
	Hwnd       string `json:"hwnd"`
	WindowName string `json:"window_name"`
	ClassName  string `json:"class_name"`
}

func (r *router) handleDetectDesktop(w http.ResponseWriter, req *http.Request) {
	classRegex := req.URL.Query().Get("class_regex")
	windowRegex := req.URL.Query().Get("window_regex")
	controllerLog.Info().Str("class_regex", classRegex).Str("window_regex", windowRegex).Msg("detect desktop windows request")

	windows, err := maa.FindDesktopWindows()
	if err != nil {
		controllerLog.Error().Err(err).Msg("find desktop windows failed")
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("find desktop windows failed: %v", err))
		return
	}

	result := make([]desktopWindowInfo, 0, len(windows))
	for _, win := range windows {
		if win == nil {
			continue
		}
		result = append(result, desktopWindowInfo{
			Hwnd:       fmt.Sprintf("%v", win.Handle),
			WindowName: win.WindowName,
			ClassName:  win.ClassName,
		})
	}

	if classRegex != "" {
		filtered := make([]desktopWindowInfo, 0, len(result))
		for _, w := range result {
			if containsOrRegexMatch(w.ClassName, classRegex) {
				filtered = append(filtered, w)
			}
		}
		result = filtered
	}
	if windowRegex != "" {
		filtered := make([]desktopWindowInfo, 0, len(result))
		for _, w := range result {
			if containsOrRegexMatch(w.WindowName, windowRegex) {
				filtered = append(filtered, w)
			}
		}
		result = filtered
	}

	controllerLog.Info().Int("total", len(windows)).Int("filtered", len(result)).Msg("detect desktop windows result")
	response.OK(w, result)
}

func (r *router) handleControllerConnect(w http.ResponseWriter, req *http.Request) {
	var payload controllerConnectRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		controllerLog.Warn().Err(err).Msg("connect request: invalid json body")
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	ctrlType := strings.TrimSpace(payload.Type)
	if ctrlType == "" {
		controllerLog.Warn().Msg("connect request: missing controller type")
		response.Fail(w, http.StatusBadRequest, "missing controller type")
		return
	}

	// 设置 connecting 状态并广播
	r.deps.StatusStore.SetController("connecting")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	controllerLog.Info().Str("type", ctrlType).Str("status", "connecting").Msg("controller status updated")

	var result maaservice.ConnectControllerResult

	switch ctrlType {
	case "adb":
		adbPath := strings.TrimSpace(payload.AdbPath)
		adbAddress := strings.TrimSpace(payload.AdbAddress)
		screencapMethod := orDefault(strings.TrimSpace(payload.AdbScreencapMethod), maaservice.ADBScreencapDefault)
		inputMethod := orDefault(strings.TrimSpace(payload.AdbInputMethod), maaservice.ADBInputDefault)
		adbConfig := strings.TrimSpace(payload.AdbConfig)

		controllerLog.Info().
			Str("type", ctrlType).
			Str("adb_path", adbPath).
			Str("adb_address", adbAddress).
			Str("screencap_method", screencapMethod).
			Str("input_method", inputMethod).
			Str("adb_config", adbConfig).
			Msg("connecting controller")

		result = r.deps.ControllerService.ConnectAdb(
			adbPath, adbAddress, screencapMethod, inputMethod, adbConfig,
		)
		// r.deps.ControllerService.Controller().SetScreenshot(maa.WithScreenshotUseRawSize(true))

	case "win32":
		hwnd := strings.TrimSpace(payload.Hwnd)
		if hwnd == "" {
			controllerLog.Warn().Str("type", ctrlType).Msg("connect request: hwnd is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "hwnd is required for Win32 controller")
			return
		}
		screencapMethod := orDefault(strings.TrimSpace(payload.Win32ScreencapMethod), maaservice.WindowScreencapGDI)
		mouseMethod := orDefault(strings.TrimSpace(payload.Win32MouseMethod), maaservice.Win32InputSeize)
		keyboardMethod := orDefault(strings.TrimSpace(payload.Win32KeyboardMethod), maaservice.Win32InputSeize)

		controllerLog.Info().
			Str("type", ctrlType).
			Str("hwnd", hwnd).
			Str("screencap_method", screencapMethod).
			Str("mouse_method", mouseMethod).
			Str("keyboard_method", keyboardMethod).
			Msg("connecting controller")

		result = r.deps.ControllerService.ConnectWin32(
			hwnd, screencapMethod, mouseMethod, keyboardMethod,
		)
		// r.deps.ControllerService.Controller().SetScreenshot(maa.WithScreenshotUseRawSize(true))

	case "gamepad":
		hwnd := strings.TrimSpace(payload.Hwnd)
		if hwnd == "" {
			controllerLog.Warn().Str("type", ctrlType).Msg("connect request: hwnd is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "hwnd is required for Gamepad controller")
			return
		}
		screencapMethod := orDefault(strings.TrimSpace(payload.GamepadScreencap), maaservice.WindowScreencapGDI)
		gamepadType := orDefault(strings.TrimSpace(payload.GamepadType), maaservice.GamepadTypeXbox360)

		controllerLog.Info().
			Str("type", ctrlType).
			Str("hwnd", hwnd).
			Str("screencap_method", screencapMethod).
			Str("gamepad_type", gamepadType).
			Msg("connecting controller")

		result = r.deps.ControllerService.ConnectGamepad(
			hwnd, screencapMethod, gamepadType,
		)
		// r.deps.ControllerService.Controller().SetScreenshot(maa.WithScreenshotUseRawSize(true))

	case "playcover":
		address := strings.TrimSpace(payload.PlayCoverAddress)
		uuid := strings.TrimSpace(payload.PlayCoverUUID)
		if address == "" {
			controllerLog.Warn().Str("type", ctrlType).Msg("connect request: address is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "address is required for PlayCover controller")
			return
		}

		controllerLog.Info().
			Str("type", ctrlType).
			Str("address", address).
			Str("uuid", uuid).
			Msg("connecting controller")

		result = r.deps.ControllerService.ConnectPlayCover(address, uuid)
		// r.deps.ControllerService.Controller().SetScreenshot(maa.WithScreenshotUseRawSize(true))

	case "wlroot":
		wlrSocketPath := strings.TrimSpace(payload.WlrootSocketPath)
		if wlrSocketPath == "" {
			controllerLog.Warn().Str("type", ctrlType).Msg("connect request: socket path is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "socket path is required for WlRoot controller")
			return
		}

		controllerLog.Info().
			Str("type", ctrlType).
			Str("socket_path", wlrSocketPath).
			Msg("connecting controller")

		result = r.deps.ControllerService.ConnectWlRoot(wlrSocketPath)
		// r.deps.ControllerService.Controller().SetScreenshot(maa.WithScreenshotUseRawSize(true))

	case "custom":
		// TODO

	default:
		controllerLog.Warn().Str("type", ctrlType).Msg("unsupported controller type")
		r.deps.StatusStore.SetController("disconnected")
		r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("unsupported controller type: %s", ctrlType))
		return

	}

	// 记录连接结果
	if result.Success {
		r.deps.StatusStore.SetController("connected")
		controllerLog.Info().Str("type", ctrlType).Str("status", "connected").Msg("connect succeeded")
		r.deps.ScreenshotService.OnConnected()
	} else {
		r.deps.StatusStore.SetController("disconnected")
		controllerLog.Warn().Str("type", ctrlType).Str("status", "disconnected").Str("error", result.Error).Msg("connect failed")
	}
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	if !result.Success {
		response.Fail(w, http.StatusBadRequest, result.Error)
		return
	}

	response.OK(w, controllerConnectResponse{Type: ctrlType})
}

func (r *router) handleControllerDisconnect(w http.ResponseWriter, _ *http.Request) {
	controllerLog.Info().Msg("disconnect request")
	r.deps.ScreenshotService.Stop()
	r.deps.ControllerService.Disconnect()
	r.deps.ScreenshotService.OnDisconnected()
	r.deps.StatusStore.SetController("disconnected")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	controllerLog.Info().Str("status", "disconnected").Msg("controller disconnected")
	response.OK(w, nil)
}

func (r *router) handleAgentConnect(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if payload.Identifier == "" {
		response.Fail(w, http.StatusBadRequest, "identifier is required")
		return
	}

	agentLog.Info().Str("identifier", payload.Identifier).Msg("connect request")

	result := r.deps.AgentService.Connect(payload.Identifier)

	r.deps.Hub.BroadcastJSON(ws.Message{Type: "agent.update", Payload: r.deps.AgentService.List()})

	if !result.Success {
		agentLog.Warn().Str("identifier", payload.Identifier).Str("error", result.Error).Msg("connect failed")
		response.Fail(w, http.StatusBadRequest, result.Error)
		return
	}

	agentLog.Info().Str("identifier", payload.Identifier).Msg("connect succeeded")
	response.OK(w, nil)
}

func (r *router) handleAgentDisconnect(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if payload.Identifier == "" {
		response.Fail(w, http.StatusBadRequest, "identifier is required")
		return
	}

	agentLog.Info().Str("identifier", payload.Identifier).Msg("disconnect request")
	r.deps.AgentService.Disconnect(payload.Identifier)

	r.deps.Hub.BroadcastJSON(ws.Message{Type: "agent.update", Payload: r.deps.AgentService.List()})
	response.OK(w, nil)
}

func (r *router) handleAgentList(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.AgentService.List())
}

func orDefault(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func parseOptionalInt64Query(req *http.Request, key string) (int64, error) {
	value := strings.TrimSpace(req.URL.Query().Get(key))
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func (r *router) handlePathExists(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	path := strings.TrimSpace(payload.Path)
	if path == "" {
		response.Fail(w, http.StatusBadRequest, "path is required")
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			response.OKWithMsg(w, map[string]bool{"exists": false}, "Path does not exist")
			return
		}
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	switch payload.Type {
	case "file":
		if info.IsDir() {
			response.OKWithMsg(w, map[string]bool{"exists": false}, "Path is a directory, expected a file")
			return
		}
	case "dir":
		if !info.IsDir() {
			response.OKWithMsg(w, map[string]bool{"exists": false}, "Path is a file, expected a directory")
			return
		}
	case "", "any":
		// no type restriction
	default:
		response.Fail(w, http.StatusBadRequest, "invalid type, must be 'file' or 'dir'")
		return
	}

	response.OKWithMsg(w, map[string]bool{"exists": true}, "Path exists")
}

func (r *router) handleResourceLoad(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		resourceLog.Warn().Err(err).Msg("load request: invalid json body")
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(payload.Paths) == 0 {
		resourceLog.Warn().Msg("load request: no paths provided")
		response.Fail(w, http.StatusBadRequest, "No resource paths provided")
		return
	}

	resourceLog.Info().Strs("paths", payload.Paths).Msg("load request")

	// 设置 loading 状态并广播
	r.deps.StatusStore.SetResource("loading")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	result := r.deps.ResourceService.LoadBundles(payload.Paths)

	if result.Success {
		r.deps.StatusStore.SetResource("loaded")
		resourceLog.Info().Str("status", "loaded").Msg("load succeeded")
	} else {
		r.deps.StatusStore.SetResource("failed")
		resourceLog.Warn().Str("status", "failed").Str("failed_path", result.FailedPath).Msg("load failed")
	}
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	if !result.Success {
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("Failed to load resource: %s", result.FailedPath))
		return
	}

	response.OK(w, nil)
}

func (r *router) handleTaskRun(w http.ResponseWriter, req *http.Request) {
	// 防止重复提交：如果已经在 running 状态则拒绝
	if r.deps.StatusStore.GetTask() == "running" {
		response.Fail(w, http.StatusConflict, "A task is already running")
		return
	}

	var payload taskRunRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		taskLog.Warn().Err(err).Msg("run request: invalid json body")
		response.Fail(w, http.StatusBadRequest, "Invalid json body")
		return
	}
	if payload.Entry == "" {
		response.Fail(w, http.StatusBadRequest, "Entry is required")
		return
	}

	taskLog.Info().Str("entry", payload.Entry).Msg("run request")
	r.deps.ScreenshotService.OnTaskStarted()
	r.deps.StatusStore.SetTask("running")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	// 异步执行任务，立即返回
	go func() {
		defer func() {
			if rv := recover(); rv != nil {
				taskLog.Error().
					Interface("panic", rv).
					Str("stack", string(debug.Stack())).
					Msg("run panic in goroutine")
				r.deps.ScreenshotService.OnTaskEnded()
				r.deps.StatusStore.SetTask("failed")
				r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
				r.deps.Hub.BroadcastJSON(ws.Message{
					Type:    "task.completed",
					Payload: taskCompletedPayload{Success: false, Error: fmt.Sprintf("internal panic: %v", rv)},
				})
			}
		}()

		result := r.deps.TaskerService.RunTask(payload.Entry, payload.PipelineOverride)

		if result.Success {
			r.deps.ScreenshotService.OnTaskEnded()
			r.deps.StatusStore.SetTask("success")
			taskLog.Info().Str("entry", payload.Entry).Str("status", "success").Msg("run succeeded")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			r.deps.Hub.BroadcastJSON(ws.Message{
				Type:    "task.completed",
				Payload: taskCompletedPayload{Success: true, Entry: payload.Entry},
			})
			return
		}

		// 用户主动停止任务后 RunTask 也会返回失败，此时状态已被 handleTaskStop 设为 stopped，
		// 不应覆盖为 failed。
		if r.deps.StatusStore.GetTask() == "stopped" {
			r.deps.ScreenshotService.OnTaskEnded()
			taskLog.Info().Str("entry", payload.Entry).Str("status", "stopped").Msg("run ended after user stop")
			r.deps.Hub.BroadcastJSON(ws.Message{
				Type:    "task.completed",
				Payload: taskCompletedPayload{Success: false, Stopped: true, Entry: payload.Entry},
			})
			return
		}

		r.deps.ScreenshotService.OnTaskEnded()
		r.deps.StatusStore.SetTask("failed")
		taskLog.Warn().Str("entry", payload.Entry).Str("status", "failed").Str("error", result.Error).Msg("run failed")
		r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
		r.deps.Hub.BroadcastJSON(ws.Message{
			Type:    "task.completed",
			Payload: taskCompletedPayload{Success: false, Error: result.Error, Entry: payload.Entry},
		})
	}()

	// 立即返回，告知前端任务已提交
	response.OK(w, nil)
}

func (r *router) handleTaskStop(w http.ResponseWriter, _ *http.Request) {
	taskLog.Info().Msg("stop request")
	r.deps.StatusStore.SetTask("stopped")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	r.deps.TaskerService.StopTask()
	response.OK(w, nil)
}

func (r *router) handleTaskNodes(w http.ResponseWriter, _ *http.Request) {
	nodes := r.deps.TaskerService.GetNodeList()
	response.OK(w, nodes)
}

func (r *router) handleTaskNodeDetail(w http.ResponseWriter, req *http.Request) {
	name := req.PathValue("name")
	if name == "" {
		response.Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	detail, err := r.deps.TaskerService.GetLatestNodeDetail(name)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(w, detail)
}

func (r *router) handleTaskNodeData(w http.ResponseWriter, req *http.Request) {
	name := req.PathValue("name")
	if name == "" {
		response.Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	recoID, err := parseOptionalInt64Query(req, "reco_id")
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid reco id")
		return
	}
	actionID, err := parseOptionalInt64Query(req, "action_id")
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid action id")
		return
	}

	detail, err := r.deps.TaskerService.GetNode(name, recoID, actionID)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(w, detail)
}

func (r *router) handleTaskRecoDetail(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid reco id")
		return
	}
	detail, err := r.deps.TaskerService.GetRecognitionDetailByID(id)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, detail)
}

func (r *router) handleTaskActionDetail(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid action id")
		return
	}
	detail, err := r.deps.TaskerService.GetActionDetailByID(id)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, detail)
}

func (r *router) handleTaskImage(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSpace(req.PathValue("id"))
	if id == "" {
		response.Fail(w, http.StatusBadRequest, "image id is required")
		return
	}
	item, ok := r.deps.TaskerService.GetTaskImage(id)
	if !ok {
		response.Fail(w, http.StatusNotFound, "image not found")
		return
	}
	if err := maaservice.WriteTaskImageResponse(w, req, item); err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}
}

// --- Clear ---
func (r *router) handleClearCache(w http.ResponseWriter, req *http.Request) {
	r.deps.TaskerService.ClearTaskImages()
	r.deps.TaskerService.ClearCache()
	response.OK(w, nil)
}

func (r *router) handleGetControllerMethods(w http.ResponseWriter, req *http.Request) {
	methodType := req.URL.Query().Get("method_type")

	methods := r.deps.ControllerService.ControllerMethod(maaservice.MethodType(methodType))
	response.OK(w, methods)
}

// --- Screenshot handlers ---

func (r *router) handleScreenshotStart(w http.ResponseWriter, _ *http.Request) {
	r.deps.ScreenshotService.Start()
	response.OK(w, nil)
}

func (r *router) handleScreenshotStop(w http.ResponseWriter, _ *http.Request) {
	r.deps.ScreenshotService.Stop()
	response.OK(w, nil)
}

func (r *router) handleScreenshotPause(w http.ResponseWriter, _ *http.Request) {
	r.deps.ScreenshotService.Pause()
	response.OK(w, nil)
}

func (r *router) handleScreenshotResume(w http.ResponseWriter, _ *http.Request) {
	r.deps.ScreenshotService.Resume()
	response.OK(w, nil)
}

func (r *router) handleScreenshotSetFPS(w http.ResponseWriter, req *http.Request) {
	var payload screenshotSetFPSRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	r.deps.ScreenshotService.SetFPS(payload.FPS)
	response.OK(w, screenshotSetFPSResponse{FPS: r.deps.ScreenshotService.GetFPS()})
}

func (r *router) handleScreenshotSetOutput(w http.ResponseWriter, req *http.Request) {
	var payload screenshotSetOutputRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	outputs := r.deps.ScreenshotService.SetOutputDemand(payload.Output)
	response.OK(w, screenshotOutputStatusResponse{
		Output: outputs,
		JPEG:   r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputJPEG),
		H264:   r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputH264),
		H265:   r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputH265),
	})
}

func (r *router) handleScreenshotStatus(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, screenshotStatusResponse{
		Running:        r.deps.ScreenshotService.Running(),
		Paused:         r.deps.ScreenshotService.Paused(),
		OutputActive:   r.deps.ScreenshotService.OutputActive(),
		FPS:            r.deps.ScreenshotService.GetFPS(),
		Output:         r.deps.ScreenshotService.OutputDemand(),
		JPEG:           r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputJPEG),
		H264:           r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputH264),
		H265:           r.deps.ScreenshotService.OutputEnabled(maaservice.ScreenshotOutputH265),
		OverlayState:   r.deps.ScreenshotService.OverlayState(),
		OverlayMessage: r.deps.ScreenshotService.OverlayMessage(),
	})
}

func (r *router) handleScreenshotRaw(w http.ResponseWriter, req *http.Request) {
	img := r.deps.ScreenshotService.RawFrame()
	if img == nil {
		response.Fail(w, http.StatusNotFound, "no raw frame available")
		return
	}
	if err := maaservice.WriteImageResponse(w, req, img); err != nil {
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}
}

func (r *router) handlePipelineCheckerStart(w http.ResponseWriter, _ *http.Request) {
	r.deps.ResourceService.StartPipelineCheck()
	response.OK(w, nil)
}

func (r *router) handlePipelineCheckerStop(w http.ResponseWriter, _ *http.Request) {
	r.deps.ResourceService.StopPipelineCheck()
	response.OK(w, nil)
}

type pipelineAddPathRequest struct {
	Paths []string `json:"paths"`
}

func (r *router) handlePipelineCheckerGet(w http.ResponseWriter, _ *http.Request) {
	result := r.deps.ResourceService.PipelineChecker.GetResult()
	response.OK(w, result)
}

func (r *router) handlePipelineCheckerCheck(w http.ResponseWriter, _ *http.Request) {
	result := r.deps.ResourceService.PipelineChecker.CheckOnce()
	response.OK(w, result)
}

func (r *router) handleWS(w http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	client := r.deps.Hub.Register(conn)

	client.SendJSON(ws.Message{
		Type:    "status.update",
		Payload: r.deps.StatusStore.Get(),
	})

	go func() {
		defer r.deps.Hub.Remove(client)

		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var msg ws.Message
			if err := json.Unmarshal(payload, &msg); err != nil {
				continue
			}

			if msg.Type == "status.subscribe" {
				client.SendJSON(ws.Message{
					Type:    "status.update",
					Payload: r.deps.StatusStore.Get(),
				})
			}
		}
	}()
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				httpLog.Error().
					Any("panic", r).
					Bytes("stack", debug.Stack()).
					Str("path", req.URL.Path).
					Msg("request panic")
				response.Fail(w, http.StatusInternalServerError, fmt.Sprintf("internal error: %v", r))
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// WebSocket 升级要求底层 ResponseWriter 支持 Hijacker。
		// 包装器可能丢失该接口，导致 /ws 返回 500。
		if req.URL.Path == "/ws" {
			next.ServeHTTP(w, req)
			httpLog.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", http.StatusSwitchingProtocols).
				Dur("latency", time.Since(start)).
				Msg("http request")
			return
		}

		var requestBody string
		if isAPIPath(req.URL.Path) && req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err == nil {
				requestBody = formatLogBody(req.URL.Path, body)
				req.Body = io.NopCloser(bytes.NewReader(body))
			} else {
				requestBody = fmt.Sprintf("[read body failed: %v]", err)
			}
		}

		recvEvt := httpLog.Info().
			Str("method", req.Method).
			Str("path", req.URL.Path)
		if req.URL.RawQuery != "" {
			recvEvt = recvEvt.Str("query", truncateLogField(req.URL.RawQuery, 4096))
		}
		if requestBody != "" {
			recvEvt = recvEvt.Str("request_body", truncateLogField(requestBody, 4096))
		}
		recvEvt.Msg("http request received")

		rw := &statusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			captureBody:    shouldCaptureResponseBody(req.URL.Path),
		}
		next.ServeHTTP(rw, req)

		respEvt := httpLog.Info().
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Int("status", rw.status).
			Dur("latency", time.Since(start))
		if rw.captureBody && rw.body.Len() > 0 {
			respEvt = respEvt.Str("response_body", truncateLogField(formatLogBody(req.URL.Path, rw.body.Bytes()), 4096))
		}
		respEvt.Msg("http request completed")
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	captureBody bool
	body        bytes.Buffer
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.captureBody && len(b) > 0 {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

func shouldCaptureResponseBody(path string) bool {
	if !isAPIPath(path) {
		return false
	}
	return !strings.HasPrefix(path, "/api/task/image/")
}

func formatLogBody(path string, body []byte) string {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ""
	}
	if !looksLikeJSONBody(trimmed) {
		return string(trimmed)
	}
	if shouldSanitizeLogBody(path) && bytes.Contains(bytes.ToLower(trimmed), []byte("raw_image")) {
		return sanitizeAndCompactLogJSON(trimmed)
	}
	return compactLogJSON(trimmed)
}

func compactLogJSON(body []byte) string {
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, body); err == nil {
		return compacted.String()
	}
	return string(body)
}

func sanitizeAndCompactLogJSON(body []byte) string {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}

	sanitizeLogValue(payload)

	sanitized, err := json.Marshal(payload)
	if err != nil {
		return string(body)
	}
	return compactLogJSON(sanitized)
}

func sanitizeLogValue(v any) {
	switch value := v.(type) {
	case map[string]any:
		for key, item := range value {
			if strings.EqualFold(key, "raw_image") {
				value[key] = "[skipped]"
				continue
			}
			sanitizeLogValue(item)
		}
	case []any:
		for _, item := range value {
			sanitizeLogValue(item)
		}
	}
}

func looksLikeJSONBody(body []byte) bool {
	if len(body) < 2 {
		return false
	}
	return (body[0] == '{' && body[len(body)-1] == '}') || (body[0] == '[' && body[len(body)-1] == ']')
}

func shouldSanitizeLogBody(path string) bool {
	return strings.Contains(path, "/controller/connect") || strings.Contains(path, "/task/")
}

func truncateLogField(v string, max int) string {
	if len(v) <= max {
		return v
	}
	return v[:max] + "...(truncated)"
}

func containsOrRegexMatch(value, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}

	matcher, err := compileDesktopFilter(pattern)
	if err != nil {
		controllerLog.Warn().Err(err).Str("pattern", pattern).Msg("invalid desktop filter pattern, fallback to plain contains")
		return value == pattern || strings.Contains(value, pattern)
	}

	return matcher(value)
}

func compileDesktopFilter(pattern string) (func(string) bool, error) {
	tokens := splitDesktopFilter(pattern)
	matcher, next, err := parseDesktopOrExpr(tokens, 0)
	if err != nil {
		return nil, err
	}
	if next != len(tokens) {
		return nil, fmt.Errorf("unexpected token %q", tokens[next])
	}
	return matcher, nil
}

func parseDesktopOrExpr(tokens []string, index int) (func(string) bool, int, error) {
	left, next, err := parseDesktopAndExpr(tokens, index)
	if err != nil {
		return nil, index, err
	}

	for next < len(tokens) && strings.EqualFold(tokens[next], "or") {
		right, afterRight, parseErr := parseDesktopAndExpr(tokens, next+1)
		if parseErr != nil {
			return nil, index, parseErr
		}
		prev := left
		left = func(value string) bool {
			return prev(value) || right(value)
		}
		next = afterRight
	}

	return left, next, nil
}

func parseDesktopAndExpr(tokens []string, index int) (func(string) bool, int, error) {
	left, next, err := parseDesktopPrimary(tokens, index)
	if err != nil {
		return nil, index, err
	}

	for next < len(tokens) && strings.EqualFold(tokens[next], "and") {
		right, afterRight, parseErr := parseDesktopPrimary(tokens, next+1)
		if parseErr != nil {
			return nil, index, parseErr
		}
		prev := left
		left = func(value string) bool {
			return prev(value) && right(value)
		}
		next = afterRight
	}

	return left, next, nil
}

func parseDesktopPrimary(tokens []string, index int) (func(string) bool, int, error) {
	if index >= len(tokens) {
		return nil, index, fmt.Errorf("missing operand")
	}

	token := tokens[index]
	if token == "(" {
		matcher, next, err := parseDesktopOrExpr(tokens, index+1)
		if err != nil {
			return nil, index, err
		}
		if next >= len(tokens) || tokens[next] != ")" {
			return nil, index, fmt.Errorf("missing closing parenthesis")
		}
		return matcher, next + 1, nil
	}

	if token == ")" {
		return nil, index, fmt.Errorf("unexpected closing parenthesis")
	}

	if strings.EqualFold(token, "and") || strings.EqualFold(token, "or") {
		return nil, index, fmt.Errorf("unexpected operator %q", token)
	}

	matcher, err := compileDesktopTerm(token)
	if err != nil {
		return nil, index, err
	}
	return matcher, index + 1, nil
}

func compileDesktopTerm(token string) (func(string) bool, error) {
	term := strings.TrimSpace(token)
	if term == "" {
		return nil, fmt.Errorf("empty term")
	}

	re, err := regexp.Compile(term)
	if err != nil {
		return nil, fmt.Errorf("compile regex %q: %w", term, err)
	}

	return func(value string) bool {
		return value == term || strings.Contains(value, term) || re.MatchString(value)
	}, nil
}

func splitDesktopFilter(pattern string) []string {
	var tokens []string
	var current strings.Builder
	var escaped bool
	var inCharClass bool

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range pattern {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case inCharClass:
			current.WriteRune(r)
			if r == ']' {
				inCharClass = false
			}
		case r == '[':
			current.WriteRune(r)
			inCharClass = true
		case r == '(' || r == ')':
			flush()
			tokens = append(tokens, string(r))
		case unicode.IsSpace(r):
			flush()
		default:
			current.WriteRune(r)
		}
	}

	flush()
	return mergeDesktopOperatorTokens(tokens)
}

func mergeDesktopOperatorTokens(tokens []string) []string {
	merged := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if i+2 < len(tokens) && isDesktopOperatorWord(tokens[i]) && isDesktopOperatorWord(tokens[i+1]) && isDesktopOperatorWord(tokens[i+2]) {
			candidate := tokens[i] + tokens[i+1] + tokens[i+2]
			if strings.EqualFold(candidate, "and") || strings.EqualFold(candidate, "or") {
				merged = append(merged, candidate)
				i += 2
				continue
			}
		}
		merged = append(merged, tokens[i])
	}
	return merged
}

func isDesktopOperatorWord(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(token) == 1
}
