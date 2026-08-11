// Package api berisi HTTP layer di atas authorization engine —
// supaya engine bisa dipakai lintas bahasa (bukan cuma import Go).
package api

import "github.com/amayones/authz-engine/internal/model"

// --- Role & RBAC ---

type createRoleRequest struct {
	Name        string         `json:"name"`
	Permissions []string       `json:"permissions"`
	Conditions  []conditionDTO `json:"conditions,omitempty"`
}

type conditionDTO struct {
	AttrKey   string `json:"attr_key"`
	AttrValue string `json:"attr_value"`
}

type assignRoleRequest struct {
	SubjectID string `json:"subject_id"`
	RoleName  string `json:"role_name"`
}

// --- Can (RBAC + ABAC + ReBAC gabungan) ---

type canRequest struct {
	SubjectID string            `json:"subject_id"`
	Resource  string            `json:"resource,omitempty"`
	Action    string            `json:"action"`
	Object    string            `json:"object,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
}

type canResponse struct {
	Allowed bool `json:"allowed"`
}

// --- ReBAC relations ---

type relationRequest struct {
	Object   string `json:"object"`
	Relation string `json:"relation"`
	Subject  string `json:"subject"`
}

// --- Attributes (ABAC) ---

type setAttributeRequest struct {
	SubjectID string `json:"subject_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

// --- Generic response ---

type errorResponse struct {
	Error string `json:"error"`
}

type okResponse struct {
	Status string `json:"status"`
}

func toPermissions(strs []string) []model.Permission {
	perms := make([]model.Permission, len(strs))
	for i, s := range strs {
		perms[i] = model.Permission(s)
	}
	return perms
}

func toConditions(dtos []conditionDTO) []model.Condition {
	conds := make([]model.Condition, len(dtos))
	for i, d := range dtos {
		conds[i] = model.Condition{AttrKey: d.AttrKey, AttrValue: d.AttrValue}
	}
	return conds
}
