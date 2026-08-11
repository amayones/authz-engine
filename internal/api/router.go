package api

import (
	"net/http"
	"time"

	"github.com/amayones/authz-engine/internal/store"
)

func NewRouter(s *Server, st store.APIKeyStore, limiter *RateLimiter) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /roles", s.handleCreateRole)
	mux.HandleFunc("POST /roles/assign", s.handleAssignRole)
	mux.HandleFunc("POST /roles/revoke", s.handleRevokeRole)

	mux.HandleFunc("POST /can", s.handleCan)

	mux.HandleFunc("POST /relations", s.handleWriteRelation)
	mux.HandleFunc("DELETE /relations", s.handleDeleteRelation)
	mux.HandleFunc("POST /relations/check", s.handleCheckRelation)

	mux.HandleFunc("POST /attributes", s.handleSetAttribute)

	return chain(mux,
		recoveryMiddleware,
		loggingMiddleware,
		timeoutMiddleware(8*time.Second),
		authMiddleware(st, limiter),
	)
}
