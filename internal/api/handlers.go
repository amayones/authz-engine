package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
)

// Server membungkus engine dan menyediakan http.Handler untuk tiap endpoint.
type Server struct {
	engine *engine.Engine
}

func NewServer(e *engine.Engine) *Server {
	return &Server{engine: e}
}

// writeJSON adalah helper supaya semua response konsisten formatnya.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// --- Role & RBAC handlers ---

func (s *Server) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	role := model.Role{
		Name:        req.Name,
		Permissions: toPermissions(req.Permissions),
		Conditions:  toConditions(req.Conditions),
	}

	// Engine.CreateRole saat ini menerima (ctx, name, permissions) saja —
	// kalau mau ikut kirim Conditions, panggil store langsung lewat helper
	// tambahan di engine (lihat catatan di bawah kode ini).
	if err := s.engine.CreateRoleWithConditions(r.Context(), role); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "role already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, okResponse{Status: "created"})
}

func (s *Server) handleAssignRole(w http.ResponseWriter, r *http.Request) {
	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.SubjectID == "" || req.RoleName == "" {
		writeError(w, http.StatusBadRequest, "subject_id and role_name are required")
		return
	}

	if err := s.engine.AssignRole(r.Context(), req.SubjectID, req.RoleName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, okResponse{Status: "assigned"})
}

func (s *Server) handleRevokeRole(w http.ResponseWriter, r *http.Request) {
	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.SubjectID == "" || req.RoleName == "" {
		writeError(w, http.StatusBadRequest, "subject_id and role_name are required")
		return
	}

	if err := s.engine.RevokeRole(r.Context(), req.SubjectID, req.RoleName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, okResponse{Status: "revoked"})
}

// --- Can (titik keputusan utama) ---

func (s *Server) handleCan(w http.ResponseWriter, r *http.Request) {
	var req canRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.SubjectID == "" || req.Action == "" {
		writeError(w, http.StatusBadRequest, "subject_id and action are required")
		return
	}

	allowed, err := s.engine.Can(r.Context(), model.AccessRequest{
		SubjectID: req.SubjectID,
		Resource:  req.Resource,
		Action:    req.Action,
		Object:    req.Object,
		Context:   req.Context,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, canResponse{Allowed: allowed})
}

// --- ReBAC relations ---

func (s *Server) handleWriteRelation(w http.ResponseWriter, r *http.Request) {
	var req relationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Object == "" || req.Relation == "" || req.Subject == "" {
		writeError(w, http.StatusBadRequest, "object, relation, and subject are required")
		return
	}

	if err := s.engine.WriteRelation(r.Context(), req.Object, req.Relation, req.Subject); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, okResponse{Status: "created"})
}

func (s *Server) handleDeleteRelation(w http.ResponseWriter, r *http.Request) {
	var req relationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Object == "" || req.Relation == "" || req.Subject == "" {
		writeError(w, http.StatusBadRequest, "object, relation, and subject are required")
		return
	}

	if err := s.engine.DeleteRelation(r.Context(), req.Object, req.Relation, req.Subject); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, okResponse{Status: "deleted"})
}

func (s *Server) handleCheckRelation(w http.ResponseWriter, r *http.Request) {
	var req relationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Object == "" || req.Relation == "" || req.Subject == "" {
		writeError(w, http.StatusBadRequest, "object, relation, and subject are required")
		return
	}

	allowed, err := s.engine.CheckRelation(r.Context(), req.Object, req.Relation, req.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, canResponse{Allowed: allowed})
}

// --- Attributes (ABAC) ---

func (s *Server) handleSetAttribute(w http.ResponseWriter, r *http.Request) {
	var req setAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.SubjectID == "" || req.Key == "" {
		writeError(w, http.StatusBadRequest, "subject_id and key are required")
		return
	}

	if err := s.engine.SetAttribute(r.Context(), req.SubjectID, req.Key, req.Value); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, okResponse{Status: "set"})
}

// --- Health check ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, okResponse{Status: "ok"})
}
