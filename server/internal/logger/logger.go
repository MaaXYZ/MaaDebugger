package logger

import (
	"io"

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

func Init(w io.Writer) {
	log.Logger = zerolog.New(w).With().Timestamp().Logger()
}

func For(component Component) zerolog.Logger {
	return log.With().Str(componentFieldName, string(component)).Logger()
}
