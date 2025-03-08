package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, errMsg string, status int) {
	SetContentJSON(w)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: errMsg})
}

func SetContentJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
