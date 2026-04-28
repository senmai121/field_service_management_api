package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON encodes v as JSON and writes it with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// WriteError writes a JSON error response: {"error": msg}.
// Exported so middleware can reuse it without an import cycle.
func WriteError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeError is the unexported alias for internal handler use.
func writeError(w http.ResponseWriter, status int, msg string) {
	WriteError(w, status, msg)
}
