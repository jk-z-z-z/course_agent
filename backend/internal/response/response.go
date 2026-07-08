package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func WriteJSON(w http.ResponseWriter, status int, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Code: code, Message: message, Data: data})
}

func Success(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, 0, "success", data)
}

func Fail(w http.ResponseWriter, status int, code int, message string) {
	WriteJSON(w, status, code, message, nil)
}
