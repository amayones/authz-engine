package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/memory"
)

func main() {
	ctx := context.Background()
	e := engine.New(memory.New())

	e.SetSchema(model.RelationSchema{
		"viewer": {"editor", "owner"},
		"editor": {"owner"},
	})

	// --- Skenario 1: Admin lewat RBAC klasik (role global) ---
	_ = e.CreateRole(ctx, "admin", []model.Permission{"invoice:*"})
	_ = e.AssignRole(ctx, "user:amayones", "admin")

	// --- Skenario 2: alice tidak punya role, tapi owner satu document spesifik ---
	_ = e.WriteRelation(ctx, "document:proposal", "owner", "user:alice")

	// --- Skenario 3: bob akses lewat keanggotaan grup ---
	_ = e.WriteRelation(ctx, "document:proposal", "viewer", "group:eng#member")
	_ = e.WriteRelation(ctx, "group:eng", "member", "user:bob")

	checks := []model.AccessRequest{
		{SubjectID: "user:amayones", Resource: "invoice", Action: "delete"},        // RBAC (wildcard admin)
		{SubjectID: "user:alice", Action: "owner", Object: "document:proposal"},    // ReBAC langsung
		{SubjectID: "user:alice", Action: "viewer", Object: "document:proposal"},   // ReBAC via hierarki
		{SubjectID: "user:bob", Action: "viewer", Object: "document:proposal"},     // ReBAC via grup
		{SubjectID: "user:charlie", Action: "viewer", Object: "document:proposal"}, // ditolak semua jalur
	}

	for _, req := range checks {
		allowed, err := e.Can(ctx, req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("subject=%s action=%s object=%s -> allowed=%v\n",
			req.SubjectID, req.Action, req.Object, allowed)
	}
}
