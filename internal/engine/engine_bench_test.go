package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/memory"
)

// setupNestedGroups membuat graph ReBAC bertingkat: document dibagikan
// ke group besar, dan user ada di paling dalam nested group — kasus
// yang paling berat untuk traversal.
func setupNestedGroups(b *testing.B, e *engine.Engine, depth int) {
	ctx := context.Background()
	_ = e.WriteRelation(ctx, "document:1", "viewer", "group:g0#member")
	for i := 0; i < depth; i++ {
		_ = e.WriteRelation(ctx, "group:g"+string(rune('0'+i)), "member", "group:g"+string(rune('0'+i+1))+"#member")
	}
	_ = e.WriteRelation(ctx, "group:g"+string(rune('0'+depth)), "member", "user:target")
}

func BenchmarkCheckRelation_NoCache(b *testing.B) {
	e := engine.New(memory.New())
	setupNestedGroups(b, e, 5)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.CheckRelation(ctx, "document:1", "viewer", "user:target")
	}
}

func BenchmarkCheckRelation_WithCache(b *testing.B) {
	e := engine.NewWithCache(memory.New(), 30*time.Second)
	setupNestedGroups(b, e, 5)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.CheckRelation(ctx, "document:1", "viewer", "user:target")
	}
}

func BenchmarkCan_WithCache(b *testing.B) {
	e := engine.NewWithCache(memory.New(), 30*time.Second)
	ctx := context.Background()
	_ = e.CreateRole(ctx, "editor", []model.Permission{"invoice:read"})
	_ = e.AssignRole(ctx, "user-1", "editor")

	req := model.AccessRequest{SubjectID: "user-1", Resource: "invoice", Action: "read"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Can(ctx, req)
	}
}
