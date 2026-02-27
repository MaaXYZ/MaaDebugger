package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/frontend"
	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/maaservice"
	"github.com/MaaXYZ/MaaDebugger/internal/response"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
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

		// On recognition events, notify screenshot service to read from cache
		if msgStr, _ := msg["msg"].(string); len(msgStr) >= 11 && msgStr[:11] == "Recognition" {
			deps.ScreenshotService.NotifyRecoUpdate()
		}
	})

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/info/version", r.handleInfoVersion)
	mux.HandleFunc("GET /api/info/status", r.handleInfoStatus)
	mux.HandleFunc("GET /api/config", r.handleConfigAll)
	mux.HandleFunc("GET /api/config/{key}", r.handleConfigGet)
	mux.HandleFunc("PUT /api/config/{key}", r.handleConfigSet)
	mux.HandleFunc("PUT /api/config", r.handleConfigMerge)
	mux.HandleFunc("GET /api/controller/detect/adb", r.handleDetectAdb)
	mux.HandleFunc("GET /api/controller/detect/desktop", r.handleDetectDesktop)
	mux.HandleFunc("POST /api/controller/connect", r.handleControllerConnect)
	mux.HandleFunc("POST /api/controller/disconnect", r.handleControllerDisconnect)
	mux.HandleFunc("POST /api/resource/load", r.handleResourceLoad)
	mux.HandleFunc("POST /api/task/run", r.handleTaskRun)
	mux.HandleFunc("POST /api/task/stop", r.handleTaskStop)
	mux.HandleFunc("GET /api/task/nodes", r.handleTaskNodes)
	mux.HandleFunc("GET /api/task/node/{name}", r.handleTaskNodeDetail)
	mux.HandleFunc("GET /api/task/reco/{id}", r.handleTaskRecoDetail)
	mux.HandleFunc("GET /api/task/action/{id}", r.handleTaskActionDetail)
	mux.HandleFunc("POST /api/agent/connect", r.handleAgentConnect)
	mux.HandleFunc("POST /api/agent/disconnect", r.handleAgentDisconnect)
	mux.HandleFunc("GET /api/agent/list", r.handleAgentList)
	mux.HandleFunc("POST /api/screenshot/start", r.handleScreenshotStart)
	mux.HandleFunc("POST /api/screenshot/stop", r.handleScreenshotStop)
	mux.HandleFunc("POST /api/screenshot/pause", r.handleScreenshotPause)
	mux.HandleFunc("POST /api/screenshot/resume", r.handleScreenshotResume)
	mux.HandleFunc("PUT /api/screenshot/fps", r.handleScreenshotSetFPS)
	mux.HandleFunc("GET /api/screenshot/status", r.handleScreenshotStatus)
	mux.HandleFunc("GET /ws", r.handleWS)

	// Serve embedded frontend SPA for all non-API routes
	frontendHandler := frontend.Handler()
	mux.Handle("/", frontendHandler)

	return recoverer(logging(cors(mux)))
}

func (r *router) handleInfoVersion(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, "unknown")
}

func (r *router) handleInfoStatus(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.StatusStore.Get())
}

func (r *router) handleConfigAll(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.ConfigStore.GetAll())
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
	var payload any
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	r.deps.ConfigStore.Set(key, payload)
	response.OK(w, nil)
}

func (r *router) handleConfigMerge(w http.ResponseWriter, req *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	r.deps.ConfigStore.Merge(payload)
	response.OK(w, nil)
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
	log.Info().Msg("[Controller] detect ADB devices request")
	devices, err := maa.FindAdbDevices()
	if err != nil {
		log.Error().Err(err).Msg("[Controller] find adb devices failed")
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

	log.Info().Int("count", len(result)).Msg("[Controller] detect ADB devices result")
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
	log.Info().Str("class_regex", classRegex).Str("window_regex", windowRegex).Msg("[Controller] detect desktop windows request")

	windows, err := maa.FindDesktopWindows()
	if err != nil {
		log.Error().Err(err).Msg("[Controller] find desktop windows failed")
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

	log.Info().Int("total", len(windows)).Int("filtered", len(result)).Msg("[Controller] detect desktop windows result")
	response.OK(w, result)
}

func (r *router) handleControllerConnect(w http.ResponseWriter, req *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Warn().Err(err).Msg("[Controller] connect: invalid json body")
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	// 记录请求参数
	log.Info().Interface("params", payload).Msg("[Controller] connect request")

	ctrlType, _ := payload["type"].(string)
	if ctrlType == "" {
		log.Warn().Msg("[Controller] connect: missing controller type")
		response.Fail(w, http.StatusBadRequest, "missing controller type")
		return
	}

	// 设置 connecting 状态并广播
	r.deps.StatusStore.SetController("connecting")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	log.Info().Str("type", ctrlType).Msg("[Controller] status → connecting")

	getString := func(key string) string {
		v, _ := payload[key].(string)
		return v
	}

	var result maaservice.ConnectAdbResult

	switch ctrlType {
	case "adb":
		adbPath := getString("adb_path")
		adbAddress := getString("adb_address")
		screencapMethod := orDefault(getString("adb_screencap_method"), "18446744073709551559")
		inputMethod := orDefault(getString("adb_input_method"), "18446744073709551607")
		adbConfig := getString("adb_config")

		log.Info().
			Str("adb_path", adbPath).
			Str("adb_address", adbAddress).
			Str("screencap_method", screencapMethod).
			Str("input_method", inputMethod).
			Str("adb_config", adbConfig).
			Msg("[Controller] connecting ADB")

		result = r.deps.ControllerService.ConnectAdb(
			adbPath, adbAddress, screencapMethod, inputMethod, adbConfig,
		)

	case "win32":
		hwnd := getString("hwnd")
		if hwnd == "" {
			log.Warn().Msg("[Controller] connect win32: hwnd is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "hwnd is required for Win32 controller")
			return
		}
		screencapMethod := orDefault(getString("win32_screencap_method"), "1")
		mouseMethod := orDefault(getString("win32_mouse_method"), "1")
		keyboardMethod := orDefault(getString("win32_keyboard_method"), "1")

		log.Info().
			Str("hwnd", hwnd).
			Str("screencap_method", screencapMethod).
			Str("mouse_method", mouseMethod).
			Str("keyboard_method", keyboardMethod).
			Msg("[Controller] connecting Win32")

		result = r.deps.ControllerService.ConnectWin32(
			hwnd, screencapMethod, mouseMethod, keyboardMethod,
		)

	case "gamepad":
		hwnd := getString("hwnd")
		if hwnd == "" {
			log.Warn().Msg("[Controller] connect gamepad: hwnd is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "hwnd is required for Gamepad controller")
			return
		}
		screencapMethod := orDefault(getString("gamepad_screencap_method"), "1")
		gamepadType := orDefault(getString("gamepad_type"), "0")

		log.Info().
			Str("hwnd", hwnd).
			Str("screencap_method", screencapMethod).
			Str("gamepad_type", gamepadType).
			Msg("[Controller] connecting Gamepad")

		result = r.deps.ControllerService.ConnectGamepad(
			hwnd, screencapMethod, gamepadType,
		)

	case "playcover":
		address := getString("playcover_address")
		uuid := getString("playcover_uuid")
		if address == "" {
			log.Warn().Msg("[Controller] connect playcover: address is empty")
			r.deps.StatusStore.SetController("disconnected")
			r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
			response.Fail(w, http.StatusBadRequest, "address is required for PlayCover controller")
			return
		}

		log.Info().
			Str("address", address).
			Str("uuid", uuid).
			Msg("[Controller] connecting PlayCover")

		result = r.deps.ControllerService.ConnectPlayCover(address, uuid)

	default:
		log.Warn().Str("type", ctrlType).Msg("[Controller] unsupported controller type")
		r.deps.StatusStore.SetController("disconnected")
		r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("unsupported controller type: %s", ctrlType))
		return
	}

	// 记录连接结果
	if result.Success {
		r.deps.StatusStore.SetController("connected")
		log.Info().Str("type", ctrlType).Msg("[Controller] connect succeeded, status → connected")
		r.deps.ScreenshotService.Start()
	} else {
		r.deps.StatusStore.SetController("disconnected")
		log.Warn().Str("type", ctrlType).Str("error", result.Error).Msg("[Controller] connect failed, status → disconnected")
	}
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	if !result.Success {
		response.Fail(w, http.StatusBadRequest, result.Error)
		return
	}

	response.OK(w, map[string]any{"type": ctrlType})
}

func (r *router) handleControllerDisconnect(w http.ResponseWriter, _ *http.Request) {
	log.Info().Msg("[Controller] disconnect request")
	r.deps.ScreenshotService.Stop()
	r.deps.ControllerService.Disconnect()
	r.deps.StatusStore.SetController("disconnected")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	log.Info().Msg("[Controller] disconnected, status → disconnected")
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

	log.Info().Str("identifier", payload.Identifier).Msg("[Agent] connect request")

	result := r.deps.AgentService.Connect(payload.Identifier)

	r.deps.Hub.BroadcastJSON(ws.Message{Type: "agent.update", Payload: r.deps.AgentService.List()})

	if !result.Success {
		log.Warn().Str("identifier", payload.Identifier).Str("error", result.Error).Msg("[Agent] connect failed")
		response.Fail(w, http.StatusBadRequest, result.Error)
		return
	}

	log.Info().Str("identifier", payload.Identifier).Msg("[Agent] connect succeeded")
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

	log.Info().Str("identifier", payload.Identifier).Msg("[Agent] disconnect request")
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

func (r *router) handleResourceLoad(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Warn().Err(err).Msg("[Resource] load: invalid json body")
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(payload.Paths) == 0 {
		log.Warn().Msg("[Resource] load: no paths provided")
		response.Fail(w, http.StatusBadRequest, "No resource paths provided")
		return
	}

	log.Info().Strs("paths", payload.Paths).Msg("[Resource] load request")

	// 设置 loading 状态并广播
	r.deps.StatusStore.SetResource("loading")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	result := r.deps.ResourceService.LoadBundles(payload.Paths)

	if result.Success {
		r.deps.StatusStore.SetResource("loaded")
		log.Info().Msg("[Resource] load succeeded, status → loaded")
	} else {
		r.deps.StatusStore.SetResource("failed")
		log.Warn().Str("failed_path", result.FailedPath).Msg("[Resource] load failed, status → failed")
	}
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	if !result.Success {
		response.Fail(w, http.StatusBadRequest, fmt.Sprintf("Failed to load resource: %s", result.FailedPath))
		return
	}

	response.OK(w, nil)
}

func (r *router) handleTaskRun(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Entry            string          `json:"entry"`
		PipelineOverride json.RawMessage `json:"pipeline_override"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Warn().Err(err).Msg("[Task] run: invalid json body")
		response.Fail(w, http.StatusBadRequest, "Invalid json body")
		return
	}
	if payload.Entry == "" {
		response.Fail(w, http.StatusBadRequest, "Entry is required")
		return
	}

	log.Info().Str("entry", payload.Entry).Msg("[Task] run request")
	r.deps.StatusStore.SetTask("running")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	result := r.deps.TaskerService.RunTask(payload.Entry, payload.PipelineOverride)
	if result.Success {
		r.deps.StatusStore.SetTask("success")
		log.Info().Str("entry", payload.Entry).Msg("[Task] run succeeded, status → success")
		r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
		response.OK(w, nil)
		return
	}

	// 用户主动停止任务后 RunTask 也会返回失败，此时状态已被 handleTaskStop 设为 stopped，
	// 不应覆盖为 failed，也不需要向前端报错。
	if r.deps.StatusStore.GetTask() == "stopped" {
		log.Info().Str("entry", payload.Entry).Msg("[Task] run ended after user stop, keeping stopped status")
		response.OK(w, nil)
		return
	}

	r.deps.StatusStore.SetTask("failed")
	log.Warn().Str("entry", payload.Entry).Str("error", result.Error).Msg("[Task] run failed, status → failed")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	response.Fail(w, http.StatusBadRequest, result.Error)
}

func (r *router) handleTaskStop(w http.ResponseWriter, _ *http.Request) {
	log.Info().Msg("[Task] stop request")
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
	var payload struct {
		FPS int32 `json:"fps"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	r.deps.ScreenshotService.SetFPS(payload.FPS)
	response.OK(w, map[string]any{"fps": r.deps.ScreenshotService.GetFPS()})
}

func (r *router) handleScreenshotStatus(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]any{
		"running": r.deps.ScreenshotService.Running(),
		"paused":  r.deps.ScreenshotService.Paused(),
		"fps":     r.deps.ScreenshotService.GetFPS(),
	})
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
				log.Error().
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
			log.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", http.StatusSwitchingProtocols).
				Dur("latency", time.Since(start)).
				Msg("http request")
			return
		}

		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, req)
		log.Info().
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Int("status", rw.status).
			Dur("latency", time.Since(start)).
			Msg("http request")
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
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func parsePort(v string, fallback int) int {
	p, err := strconv.Atoi(v)
	if err != nil || p <= 0 {
		return fallback
	}
	return p
}

func containsOrRegexMatch(value, pattern string) bool {
	// 当前先按包含匹配，后续可升级为编译正则并忽略无效表达式
	return value == pattern || (pattern != "" && contains(value, pattern))
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
