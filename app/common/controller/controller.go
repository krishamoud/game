// Package common handles all common controller functions
package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Controller handle all base methods
type Controller struct {
}

// SendJSON marshals v to a json struct and sends appropriate headers to w
func (c *Controller) SendJSON(w http.ResponseWriter, r *http.Request, v interface{}, code int) {
	w.Header().Add("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		log.Print(fmt.Sprintf("Error while encoding JSON: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": "Internal server error"}`)
	} else {
		w.WriteHeader(code)
		io.WriteString(w, string(b))
	}
}

// CheckError checks the error, and writes to w if there is an error.  Returns
// true if err is non-nil
func (c *Controller) CheckError(err error, statusCode int, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return true
	}
	return false
}
