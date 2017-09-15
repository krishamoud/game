// Package middleware is reponsible for all api middlewares
package middleware

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// RecoverHandler recovers any panics and responds with a 500 error.
func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Panic")
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
