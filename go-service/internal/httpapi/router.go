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

	"github.com/MaaXYZ/MaaDebugger/internal/response"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
)

type Dependencies struct {
	StatusStore *state.Store
	Hub         *ws.Hub
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
	devices, err := maa.FindAdbDevices()
	if err != nil {
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

	response.OK(w, result)
}

type desktopWindowInfo struct {
	Hwnd       string `json:"hwnd"`
	WindowName string `json:"window_name"`
	ClassName  string `json:"class_name"`
}

func (r *router) handleDetectDesktop(w http.ResponseWriter, req *http.Request) {
	windows, err := maa.FindDesktopWindows()
	if err != nil {
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

	classRegex := req.URL.Query().Get("class_regex")
	windowRegex := req.URL.Query().Get("window_regex")
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

	response.OK(w, result)
}

func (r *router) handleControllerConnect(w http.ResponseWriter, req *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	r.deps.StatusStore.SetController("connected")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})

	response.OK(w, map[string]interface{}{"type": payload["type"]})
}

func (r *router) handleControllerDisconnect(w http.ResponseWriter, _ *http.Request) {
	r.deps.StatusStore.SetController("disconnected")
	r.deps.Hub.BroadcastJSON(ws.Message{Type: "status.update", Payload: r.deps.StatusStore.Get()})
	response.OK(w, nil)
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
