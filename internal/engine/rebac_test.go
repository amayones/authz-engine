package engine_test

import (
	"context"
	"testing"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/memory"
)

func TestCheckRelation_DirectMatch(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	_ = e.WriteRelation(ctx, "document:123", "owner", "user:amayones")

	allowed, err := e.CheckRelation(ctx, "document:123", "owner", "user:amayones")
	if err != nil {
		t.Fatalf("CheckRelation gagal: %v", err)
	}
	if !allowed {
		t.Error("owner langsung harusnya diizinkan")
	}
}

func TestCheckRelation_DeniesUnrelatedSubject(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	_ = e.WriteRelation(ctx, "document:123", "owner", "user:amayones")

	allowed, _ := e.CheckRelation(ctx, "document:123", "owner", "user:orang-lain")
	if allowed {
		t.Error("subject yang tidak terkait harusnya ditolak")
	}
}

func TestCheckRelation_HierarchyGrantsLowerRelation(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())
	e.SetSchema(model.RelationSchema{
		"viewer": {"editor", "owner"},
		"editor": {"owner"},
	})

	_ = e.WriteRelation(ctx, "document:123", "owner", "user:amayones")

	// Owner harusnya otomatis juga viewer, walau tidak ada tuple
	// "viewer" eksplisit.
	allowed, err := e.CheckRelation(ctx, "document:123", "viewer", "user:amayones")
	if err != nil {
		t.Fatalf("CheckRelation gagal: %v", err)
	}
	if !allowed {
		t.Error("owner harusnya otomatis punya akses viewer lewat hierarki")
	}
}

func TestCheckRelation_UsersetGroupMembership(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	// document:123 diakses oleh semua member group:eng.
	_ = e.WriteRelation(ctx, "document:123", "viewer", "group:eng#member")
	// bob adalah member group:eng.
	_ = e.WriteRelation(ctx, "group:eng", "member", "user:bob")

	allowed, err := e.CheckRelation(ctx, "document:123", "viewer", "user:bob")
	if err != nil {
		t.Fatalf("CheckRelation gagal: %v", err)
	}
	if !allowed {
		t.Error("bob adalah member group:eng, harusnya jadi viewer lewat userset")
	}

	// user yang bukan member group:eng harus ditolak.
	allowed, _ = e.CheckRelation(ctx, "document:123", "viewer", "user:charlie")
	if allowed {
		t.Error("charlie bukan member group:eng, harusnya ditolak")
	}
}

func TestCheckRelation_RevokedRelationLosesAccess(t *testing.T) {
	ctx := context.Background()
	e := engine.New(memory.New())

	_ = e.WriteRelation(ctx, "document:123", "editor", "user:amayones")
	_ = e.DeleteRelation(ctx, "document:123", "editor", "user:amayones")

	allowed, _ := e.CheckRelation(ctx, "document:123", "editor", "user:amayones")
	if allowed {
		t.Error("setelah relation dihapus harusnya ditolak")
	}
}