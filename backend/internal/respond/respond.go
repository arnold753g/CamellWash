package respond

import (
	"encoding/json"
	"net/http"
	"time"
)

type Envelope struct {
	Success   bool       `json:"success"`
	Data      any        `json:"data,omitempty"`
	Message   string     `json:"message,omitempty"`
	Error     *ErrorData `json:"error,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

type ErrorData struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any, message string) {
	write(w, status, Envelope{Success: true, Data: data, Message: message, Timestamp: time.Now().UTC()})
}

func Error(w http.ResponseWriter, r *http.Request, status int, code, message string, fields map[string]string) {
	write(w, status, Envelope{Success: false, Error: &ErrorData{Code: code, Message: message, Fields: fields},
		RequestID: r.Header.Get("X-Request-ID"), Timestamp: time.Now().UTC()})
}

func write(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
