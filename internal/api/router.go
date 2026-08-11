package api

import "net/http"

// NewRouter membangun http.Handler lengkap dengan semua endpoint dan
// middleware terpasang. apiKey dipakai untuk proteksi X-API-Key.
func NewRouter(s *Server, apiKey string) http.Handler {
	mux := http.NewServeMux()

	// Go 1.22+ method-based routing pattern.
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
		apiKeyMiddleware(apiKey),
	)
}
