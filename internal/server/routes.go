package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type CheckRequest struct {
	Key      string `json:"key"`
	Limit    int    `json:"limit"`
	WindowMS int64  `json:"window_ms"`
}

type CheckResponse struct {
	Allowed   bool `json:"allowed"`
	Remaining int  `json:"remaining"`
}

func Routes(s *Server) chi.Router {
	r := chi.NewRouter()

	// middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// routes
	r.Get("/health", s.Health)
	r.Post("/check", s.Check)
	return r
}
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
func (s *Server) Check(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Key == "" || req.Limit <= 0 || req.WindowMS <= 0 {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
	defer cancel()

	allowed, remaining, err := s.limiter.Allow(
		ctx,
		req.Key,
		req.Limit,
		time.Duration(req.WindowMS)*time.Millisecond,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests) // 429
	} else {
		w.WriteHeader(http.StatusOK)
	}

	resp := CheckResponse{
		Allowed:   allowed,
		Remaining: remaining,
	}

	json.NewEncoder(w).Encode(resp)
}
