package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
}

func Error(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	JSON(w, code, errorResponse{Error: msg})
}
