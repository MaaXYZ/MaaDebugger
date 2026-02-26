package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
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

	// 获取当前工作目录
	userPath := getCwd()

	host := getenv("GO_SERVICE_HOST", "127.0.0.1")
	port := getenv("GO_SERVICE_PORT", "8011")
	addr := host + ":" + port

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
	cfgStore := configstore.New(userPath)
	defer cfgStore.Close()
	defer agentService.DisconnectAll()

	router := httpapi.NewRouter(httpapi.Dependencies{
		StatusStore:       statusStore,
		Hub:               hub,
		ControllerService: ctrlService,
		ResourceService:   resService,
		TaskerService:     taskerService,
		AgentService:      agentService,
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
