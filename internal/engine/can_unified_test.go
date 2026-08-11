package engine_test

import (
	"context"
	"testing"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/memory"
)

func TestCan_AllowsViaRBACOnly(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	_ = e.CreateRole(ctx, "editor", []model.Permission{"invoice:read"})
	_ = e.AssignRole(ctx, "user:alice", "editor")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:alice", Resource: "invoice", Action: "read",
		// Object kosong -> hanya RBAC yang dievaluasi
	})
	if err != nil {
		t.Fatalf("Can gagal: %v", err)
	}
	if !allowed {
		t.Error("harusnya diizinkan lewat RBAC")
	}
}

func TestCan_AllowsViaReBACOnly(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	// alice tidak punya role apapun, tapi punya relasi "viewer" langsung
	// ke document:123 -- RBAC akan gagal, ReBAC harus menyelamatkan.
	_ = e.WriteRelation(ctx, "document:123", "viewer", "user:alice")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:alice",
		Resource:  "document", // tidak dipakai untuk pengecekan RBAC krn tidak match
		Action:    "viewer",   // dipakai sebagai relation name untuk ReBAC
		Object:    "document:123",
	})
	if err != nil {
		t.Fatalf("Can gagal: %v", err)
	}
	if !allowed {
		t.Error("harusnya diizinkan lewat ReBAC walau tidak punya role")
	}
}

func TestCan_AllowsViaReBACUserset(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())
	e.SetSchema(model.RelationSchema{"viewer": {"editor", "owner"}})

	_ = e.WriteRelation(ctx, "document:123", "viewer", "group:eng#member")
	_ = e.WriteRelation(ctx, "group:eng", "member", "user:bob")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:bob",
		Action:    "viewer",
		Object:    "document:123",
	})
	if err != nil {
		t.Fatalf("Can gagal: %v", err)
	}
	if !allowed {
		t.Error("bob harusnya diizinkan lewat userset ReBAC")
	}
}

func TestCan_DeniesWhenNeitherRBACNorReBACMatch(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	allowed, _ := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:charlie",
		Resource:  "invoice",
		Action:    "read",
		Object:    "document:123", // ada Object, tapi tidak ada tuple sama sekali
	})
	if allowed {
		t.Error("charlie tidak punya akses lewat cara manapun, harusnya ditolak")
	}
}

func TestCan_RBACAndReBACCombinedOnSameRequest(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	// dave TIDAK punya role apapun.
	// Tapi dave punya relasi "editor" langsung ke document:999.
	_ = e.WriteRelation(ctx, "document:999", "editor", "user:dave")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:dave",
		Resource:  "document",
		Action:    "editor",
		Object:    "document:999",
	})
	if err != nil {
		t.Fatalf("Can gagal: %v", err)
	}
	if !allowed {
		t.Error("dave harusnya diizinkan murni lewat ReBAC walau RBAC kosong")
	}
}
