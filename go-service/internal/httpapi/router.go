package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/internal/maaservice"
	"github.com/MaaXYZ/MaaDebugger/internal/response"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
)

type Dependencies struct {
	StatusStore       *state.Store
	Hub               *ws.Hub
	ControllerService *maaservice.ControllerService
}

type router struct {
	deps     Dependencies
	cfgMu    sync.RWMutex
	cfgStore map[string]interface{}
	upgrader websocket.Upgrader
}

func NewRouter(deps Dependencies) http.Handler {
	r := &router{
		deps:     deps,
		cfgStore: map[string]interface{}{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", r.handleRoot)
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
	mux.HandleFunc("GET /ws", r.handleWS)

	return recoverer(logging(cors(mux)))
}

func (r *router) handleRoot(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]interface{}{
		"name":    "maa-debugger-go-service",
		"version": "0.1.0",
	})
}

func (r *router) handleInfoVersion(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, "unknown")
}

func (r *router) handleInfoStatus(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, r.deps.StatusStore.Get())
}

func (r *router) handleConfigAll(w http.ResponseWriter, _ *http.Request) {
	r.cfgMu.RLock()
	defer r.cfgMu.RUnlock()

	dup := make(map[string]interface{}, len(r.cfgStore))
	for k, v := range r.cfgStore {
		dup[k] = v
	}
	response.OK(w, dup)
}

func (r *router) handleConfigGet(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")

	r.cfgMu.RLock()
	defer r.cfgMu.RUnlock()

	v, ok := r.cfgStore[key]
	if !ok {
		response.OK(w, nil)
		return
	}
	response.OK(w, v)
}

func (r *router) handleConfigSet(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	var payload interface{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	r.cfgMu.Lock()
	r.cfgStore[key] = payload
	r.cfgMu.Unlock()

	response.OK(w, nil)
}

func (r *router) handleConfigMerge(w http.ResponseWriter, req *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	r.cfgMu.Lock()
	for k, v := range payload {
		r.cfgStore[k] = v
	}
	r.cfgMu.Unlock()

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
	var payload map[string]interface{}
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
	} else {
		r.deps.StatusStore.SetController("disconnected")
		log.Warn().Str("type", ctrlType).Str("error", result.Error).Msg("[Controller] connect failed, status → disconnected")
	}
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	if !result.Success {
		response.Fail(w, http.StatusBadRequest, result.Error)
		return
	}

	response.OK(w, map[string]interface{}{"type": ctrlType})
}

func (r *router) handleControllerDisconnect(w http.ResponseWriter, _ *http.Request) {
	log.Info().Msg("[Controller] disconnect request")
	r.deps.ControllerService.Disconnect()
	r.deps.StatusStore.SetController("disconnected")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	log.Info().Msg("[Controller] disconnected, status → disconnected")
	response.OK(w, nil)
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
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(payload.Paths) == 0 {
		response.Fail(w, http.StatusBadRequest, "No resource paths provided")
		return
	}

	r.deps.StatusStore.SetResource("loaded")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	response.OK(w, nil)
}

func (r *router) handleWS(w http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	r.deps.Hub.Add(conn)

	_ = conn.WriteJSON(ws.Message{
		Type:    "status.update",
		Payload: r.deps.StatusStore.Get(),
	})

	go func() {
		defer func() {
			r.deps.Hub.Remove(conn)
			_ = conn.Close()
		}()

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
				_ = conn.WriteJSON(ws.Message{
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
