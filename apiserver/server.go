package apiserver

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/RaikaSurendra/go_asyncapi/store"

	"github.com/RaikaSurendra/go_asyncapi/config"
)

type ApiServer struct {
	// Config is the configuration for the API server
	config *config.Config
	// Logger is the logger for the API server
	logger *slog.Logger
	//store generic struct
	store *store.Store
}

func New(conf *config.Config, logger *slog.Logger, store *store.Store) *ApiServer {
	// Create a new instance of ApiServer with the provided configuration
	// and logger
	return &ApiServer{
		config: conf,
		logger: logger,
		store:  store,
	}
}

func (s *ApiServer) ping(w http.ResponseWriter, r *http.Request) {
	// Handle the ping request
	// This is where you would implement the logic for the ping endpoint
	// For example:
	// w.Write([]byte("pong"))
	// or
	// json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	// set status to StatusOk and write pong back in response
	//read the request body
	//and write it back to the response

	w.WriteHeader(http.StatusOK)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	w.Write(body)
	w.Write([]byte("pong"))
}

func (s *ApiServer) Start(ctx context.Context) error {
	// Start the API server
	// This is where you would set up your HTTP server, routes, etc.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("POST /auth/signup", s.signupHandler())

	middleware := NewLoggerMiddleware(s.logger)
	srv := &http.Server{
		Addr:    net.JoinHostPort(s.config.ApiServerHost, s.config.ApiServerPort),
		Handler: middleware(mux)}

	// Start the server in a goroutine and return the server.ListebnAndServe() error
	go func() {
		// Log server start message with the port
		port := srv.Addr
		if port == "" {
			port = ":80" // Default HTTP port
		}
		s.logger.Info("ApiServer Started", "port", s.config.ApiServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log the error and exit if the server fails to start
			s.logger.Error("Server failed to start", "error", err)
			panic(err)
		}
	}()
	// Wait for the context to be done (e.g., signal received)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// Log server shutdown message
		s.logger.Info("Server is shutting down", slog.String("component", "ApiServer"))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			// Log the error during server shutdown
			s.logger.Error("Server shutdown error", slog.String("error", err.Error()))
		}
	}()
	// Wait for the shutdown goroutine to finish
	wg.Wait()
	// Log server shutdown complete message
	s.logger.Info("Server shutdown complete", slog.String("status", "success"), slog.String("component", "ApiServer"), slog.String("event", "shutdown"))
	// Close the server gracefully
	// Return nil if everything went well
	return nil
}
