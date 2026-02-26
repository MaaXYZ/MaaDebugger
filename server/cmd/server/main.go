package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/httpapi"
	"github.com/MaaXYZ/MaaDebugger/internal/maaservice"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	var flagPort int
	flag.IntVar(&flagPort, "port", 0, "server port (default: auto-detect from 8011)")
	flag.Parse()

	userPath := getCwd()

	host := getenv("GO_SERVICE_HOST", "localhost")

	port := resolvePort(host, flagPort)
	addr := host + ":" + strconv.Itoa(port)

	if err := maa.Init(maa.WithDebugMode(true), maa.WithLibDir(userPath+"/bin")); err != nil {
		log.Fatal().Err(err).Msg("maa init failed")
	}

	if err := maa.ConfigInitOption(userPath, "{}"); err != nil {
		log.Warn().
			Str("userPath", userPath).
			Err(err).
			Msg("Failed to init toolkit config option")
	} else {
		log.Info().
			Str("userPath", userPath).
			Msg("Toolkit config option initialized")
	}

	defer func() {
		if err := maa.Release(); err != nil {
			log.Error().Err(err).Msg("maa release failed")
		}
	}()

	statusStore := state.NewStore()
	hub := ws.NewHub()
	ctrlService := maaservice.NewControllerService()
	resService := maaservice.NewResourceService()
	taskerService := maaservice.NewTaskerService(ctrlService, resService)
	agentService := maaservice.NewAgentService(resService)
	screenshotService := maaservice.NewScreenshotService(ctrlService)
	screenshotService.SetOnFrame(func(data []byte) {
		hub.BroadcastBinary(data)
	})
	screenshotService.SetOnError(func(reason string) {
		hub.BroadcastJSON(ws.Message{
			Type:    "screenshot.error",
			Payload: map[string]string{"reason": reason},
		})
	})
	cfgStore := configstore.New(userPath)
	defer cfgStore.Close()
	defer agentService.DisconnectAll()
	defer screenshotService.Stop()

	router := httpapi.NewRouter(httpapi.Dependencies{
		StatusStore:       statusStore,
		Hub:               hub,
		ControllerService: ctrlService,
		ResourceService:   resService,
		TaskerService:     taskerService,
		AgentService:      agentService,
		ScreenshotService: screenshotService,
		ConfigStore:       cfgStore,
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Str("http", "http://"+addr).Str("ws", "ws://"+addr+"/ws").Msg("Go service started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server listen failed")
		}
	}()

	openBrowser("http://" + addr)

	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
		_ = srv.Close()
	}

	log.Info().Msg("server stopped")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Warn().Err(err).Str("url", url).Msg("failed to open browser")
	}
}

func resolvePort(host string, flagPort int) int {
	const basePort = 8011
	const maxAttempts = 100

	if flagPort > 0 {
		return flagPort
	}

	if v := os.Getenv("GO_SERVICE_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			return p
		}
	}

	for i := range maxAttempts {
		p := basePort + i
		if isPortAvailable(host, p) {
			if i > 0 {
				log.Info().Int("port", p).Msg("default port 8011 occupied, using alternative")
			}
			return p
		}
	}
	log.Fatal().
		Int("from", basePort).
		Int("to", basePort+maxAttempts-1).
		Msg("no available port found")
	return 0
}

func isPortAvailable(host string, port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
