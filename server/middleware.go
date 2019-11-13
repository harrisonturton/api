package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Middleware functions wrap an existing request
// handler and layer additional functionality on
// top.
type Middleware func(http.Handler) http.Handler

type CorsConfig struct {
	AllowOrigin  string
	AllowHeaders string
	AllowMethods string
	MaxAge       int // Max time (in seconds) the preflight can be cached
}

// PreflightMiddleware will set the following HTTP headers:
//     Access-Control-Allow-Origin: *
//     Access-Control-Allow-Headers: content-type, Content-Type, Origin, Authorization
//     Contenxt-Type: application/json; charset=utf-8
// And respond with a 200 OK status for OPTIONS requests.
func SetPreflightHeaders(config CorsConfig) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", config.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Headers", config.AllowHeaders)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

// SetHeader will set a header
func SetHeader(key, value string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			h.ServeHTTP(w, r)
		})
	}
}

// SetNopCloserRequestBody takes the r.Body (a ReadCloser) and reconstructs
// it with a ReadCloser that has a Close() method that doesn't do anything.
// This allows all subsequent handlers to read the body.
func SetNopCloserRequestBody() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
			return
			// Only required if there is a body present
			if r.Body == nil {
				h.ServeHTTP(w, r)
				return
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Could not read body")
				return
			}
			bodyBytes := bytes.NewBuffer(body)
			r.Body = ioutil.NopCloser(bodyBytes)
			h.ServeHTTP(w, r)
		})
	}
}
