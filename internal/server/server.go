package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"fin/internal/ai"
)

type Server struct {
	port     int
	provider ai.AIProvider
}

func New(port int, provider ai.AIProvider) *Server {
	return &Server{
		port:     port,
		provider: provider,
	}
}

func (s *Server) Start(ctx context.Context) error {
	if s.provider == nil {
		return errors.New("AI provider is nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRedirect)

	httpServer := &http.Server{
		Addr:              ":" + strconv.Itoa(s.port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw := r.URL.Query().Get("text")
	if raw == "" {
		http.Error(w, "missing text parameter", http.StatusBadRequest)
		return
	}

	optimized, err := s.provider.OptimizeQuery(r.Context(), raw)
	if err != nil || optimized == "" {
		http.Redirect(w, r, buildSearchURL(raw), http.StatusFound)
		return
	}

	parsed, parseErr := ai.ParseRouterResponse(optimized)
	if parseErr != nil || parsed.Query == "" {
		http.Redirect(w, r, buildSearchURL(optimized), http.StatusFound)
		return
	}

	http.Redirect(w, r, buildSearchURL(parsed.Query), http.StatusFound)
}
