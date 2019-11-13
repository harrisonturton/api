package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Server takes an API router and responds to requests
// corresponding to those routes. It handles graceful
// shutdown, logging, and some basic generic middleware.
type Server struct {
	server *http.Server
	config Config
}

// Config configures the server timeouts
type Config struct {
	Host            string
	Port            int
	Router          http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Log             *log.Logger
	Middleware      []Middleware
}

// NewServer creates and configures a new Server
// instance, but doesn't start it.
func NewServer(config Config) *Server {
	server := &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      config.Router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		},
		config: config,
	}
	server.AddMiddleware(config.Middleware...)
	return server
}

// Run will start running the server and begin listening and responding
// to requests. It will gracefully shutdown when the done channel is closed.
// It will try to serve all remaining requests, but if it takes longer than
// Config.ShutdownTimeout, they will be killed.
func (server *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	go server.waitForShutdown(ctx)
	server.config.Log.Printf("Server starting on %s\n", server.server.Addr)
	err := server.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		server.config.Log.Fatalf("Server failed: %v\n", err)
	}
}

// waitForShutdown will wait until the done channel is closed,
// and then shutdown the server. It will attempt to server
// all remaining requests before stopping, but when ShutdownTimeout
// passes, they will be cancelled.
func (server *Server) waitForShutdown(ctx context.Context) {
	<-ctx.Done()
	server.config.Log.Println("Attempting to gracefully shutdown...")
	// Forcefully stop after shutdownTimeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), server.config.ShutdownTimeout)
	defer cancel()
	// Attempt to gracefully shutdown
	if err := server.server.Shutdown(shutdownCtx); err != nil {
		server.config.Log.Fatalf("Graceful shutdown failed: %v\n", err)
		return
	}
	server.config.Log.Println("Stopped gracefully!")
}

// AddMiddleware will add multipe middleware functions.
func (server *Server) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		server.server.Handler = middleware(server.server.Handler)
	}
}
