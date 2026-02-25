package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the unified HTTP response payload.
type Envelope struct {
	Succeed bool        `json:"succeed"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func JSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Envelope{
		Succeed: true,
		Msg:     "ok",
		Data:    data,
	})
}

func Fail(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, Envelope{
		Succeed: false,
		Msg:     msg,
		Data:    nil,
	})
}
