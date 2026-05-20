package api

import (
	"encoding/json"
	"net/http"
)

// Helper because we always write a response as a json
func writeResponse(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(body)
	if err != nil { // We are trying to write the response and fail, we have nothing that we can do with the error
		panic("Error trying to write response: " + err.Error())
	}
}
