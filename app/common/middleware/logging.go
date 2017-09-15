// Package middleware is reponsible for all api middlewares
package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// LoggingHandler prints the method, url, status code, and latency of each
// request
func LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		lrw := newLoggingResponseWriter(w)
		t1 := time.Now()
		next.ServeHTTP(lrw, r)
		t2 := time.Now()
		statusCode := lrw.statusCode
		l := log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL.String(),
			"statusCode": statusCode,
			"latency":    t2.Sub(t1),
		})
		switch {
		case statusCode >= 200 && statusCode < 300:
			l.Info(time.Now().UTC().Format("2006-01-02 15:04:05"))
		case statusCode >= 300 && statusCode < 400:
			l.Warn(time.Now().UTC().Format("2006-01-02 15:04:05"))
		case statusCode >= 400:
			l.Error(time.Now().UTC().Format("2006-01-02 15:04:05"))
		}
	}
	return http.HandlerFunc(fn)
}

// AccessOriginHandler adds the correct access-origin header to each request
func AccessOriginHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		next.ServeHTTP(w, r)
		return
	}
	return http.HandlerFunc(fn)
}
