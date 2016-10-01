package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// ResponseJSONOK writes ok response.
func ResponseJSONOK(w http.ResponseWriter, v interface{}) error {
	return ResponseJSON(w, http.StatusOK, v)
}

// ResponseJSON encodes value to json and write as response.
func ResponseJSON(w http.ResponseWriter, code int, v interface{}) error {
	// Encode JSON.
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// Write response.
	s := string(b)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	io.WriteString(w, s)
	return nil
}

// RequestBind binds request data into value.
func RequestBind(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
