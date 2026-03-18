package logger

import (
	"io"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const componentFieldName = "component"

type Component string

const (
	ComponentApp         Component = "app"
	ComponentHTTP        Component = "http"
	ComponentController  Component = "controller"
	ComponentTask        Component = "task"
	ComponentAgent       Component = "agent"
	ComponentResource    Component = "resource"
	ComponentScreenshot  Component = "screenshot"
	ComponentConfigStore Component = "config_store"
	ComponentMaaService  Component = "maa_service"
	ComponentInterface   Component = "interface"
	ComponentUpdater     Component = "updater"
)

type switchableLevelWriter struct {
	mu     sync.RWMutex
	target zerolog.LevelWriter
}

func newSwitchableLevelWriter() *switchableLevelWriter {
	return &switchableLevelWriter{target: zerolog.MultiLevelWriter(io.Discard)}
}

func (w *switchableLevelWriter) Set(target io.Writer) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.target = toLevelWriter(target)
}

func (w *switchableLevelWriter) Write(p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.target.Write(p)
}

func (w *switchableLevelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.target.WriteLevel(level, p)
}

func toLevelWriter(w io.Writer) zerolog.LevelWriter {
	if lw, ok := w.(zerolog.LevelWriter); ok {
		return lw
	}
	return zerolog.MultiLevelWriter(w)
}

var rootWriter = newSwitchableLevelWriter()

func init() {
	log.Logger = zerolog.New(rootWriter).With().Timestamp().Logger()
}

func Init(w io.Writer) {
	rootWriter.Set(w)
	log.Logger = zerolog.New(rootWriter).With().Timestamp().Logger()
}

func For(component Component) zerolog.Logger {
	return log.With().Str(componentFieldName, string(component)).Logger()
}
