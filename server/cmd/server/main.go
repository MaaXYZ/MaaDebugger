package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/internal/cliargs"
	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/httpapi"
	"github.com/MaaXYZ/MaaDebugger/internal/maaservice"
	"github.com/MaaXYZ/MaaDebugger/internal/state"
	"github.com/MaaXYZ/MaaDebugger/internal/ws"
)

const DEFAULT_HOST = "localhost"
const DEFAULT_PORT = 8011

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	parser := cliargs.New("MaaDebugger", "The Offical Github repo: https://github.com/MaaXYZ/MaaDebugger")
	parser.AddInt("port", "p", "server port (default: auto-detect from 8011)", 0, false)
	parser.AddString("host", "H", "service host", "", false)
	parser.AddString("lib-path", "b", "path to maa framework binary", "", false)
	parser.AddBool("dev", "D", "Enable Dev Mode.", false)

	parsed, err := parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, parser.Help())
		os.Exit(2)
	}
	if parsed.HelpRequested {
		fmt.Println(parser.Help())
		return
	}

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

	// the config file saved to {CWD}/.maa/dbg.json
	cfgStore := configstore.New(getCwd())
	defer cfgStore.Close()
	defer agentService.DisconnectAll()
	defer screenshotService.Stop()

	// Get args
	devMode, _ := parsed.Bool("dev")
	argPath, _ := parsed.String("lib-path")

	// Set the release channel and channel path
	channel := getenv("MAADBG_CHANNEL", "github") // npm | pypi | github
	channelPath := getenv("MAADBG_CHANNEL_PATH", "")
	cfgStore.Merge(map[string]any{"channel": channel, "channel_path": channelPath})

	// Load Maa
	loadMaaFramework(devMode, argPath, channel, channelPath)
	defer func() {
		if err := maa.Release(); err != nil {
			log.Error().Err(err).Msg("maa release failed")
		}
	}()

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

	// Arg > env > DEFAULT
	host := getenv("MAADBG_HOST", DEFAULT_HOST)
	if v, ok := parsed.String("host"); ok && v != "" {
		host = v
	}
	portArg, err := strconv.Atoi(getenv("MAADBG_PORT", strconv.Itoa(DEFAULT_PORT)))
	if err != nil {
		portArg = DEFAULT_PORT
	}
	if v, ok := parsed.Int("port"); ok {
		portArg = v
	}
	port := resolvePort(host, portArg)
	addr := host + ":" + strconv.Itoa(port)

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

	if !devMode {
		openBrowser("http://" + addr)
	}

	waitForShutdown(srv)
}

// Load Maa Lib
func loadMaaFramework(devMode bool, argPath string, channel string, channelPath string) {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to run `os.Executable`.")
	}

	exePath = filepath.Dir(exePath)
	root := filepath.Join(exePath, "bin")
	if devMode {
		root = filepath.Dir(getCwd())
		root = filepath.Join(root, "bin")
	}

	// argPath > envPath > channelPath
	libPath := ""
	if channel != "" && channelPath != "" {
		libPath = channelPath
	}
	if envPath := getenv("MAAFW_BINARY_PATH", ""); envPath != "" {
		libPath = envPath
	}
	if argPath != "" {
		libPath = argPath
	}
	if libPath == "" {
		libPath = root
	}

	// Load Maa
	if err := maa.Init(maa.WithDebugMode(true), maa.WithLibDir(libPath)); err != nil {
		log.Fatal().Err(err).Msg("maa init failed")
		os.Exit(1)
	}
	if err := maa.ConfigInitOption(root, "{}"); err != nil {
		log.Warn().
			Str("userPath", root).
			Err(err).
			Msg("Failed to init toolkit config option")
	} else {
		log.Info().
			Str("userPath", root).
			Msg("Toolkit config option initialized")
	}
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
