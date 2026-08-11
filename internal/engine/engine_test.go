package engine_test

import (
	"context"
	"testing"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/memory"
)

func setup(t *testing.T) *engine.Engine {
	t.Helper()
	return engine.New(memory.New())
}

func TestCan_AllowsWhenPermissionMatches(t *testing.T) {
	ctx := context.Background()
	e := setup(t)

	_ = e.CreateRole(ctx, "editor", []model.Permission{"invoice:read", "invoice:write"})
	_ = e.AssignRole(ctx, "user-1", "editor")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user-1", Resource: "invoice", Action: "read",
	})
	if err != nil {
		t.Fatalf("Can gagal: %v", err)
	}
	if !allowed {
		t.Error("harusnya diizinkan, tapi ditolak")
	}
}

func TestCan_DeniesWhenPermissionMissing(t *testing.T) {
	ctx := context.Background()
	e := setup(t)

	_ = e.CreateRole(ctx, "viewer", []model.Permission{"invoice:read"})
	_ = e.AssignRole(ctx, "user-1", "viewer")

	allowed, _ := e.Can(ctx, model.AccessRequest{
		SubjectID: "user-1", Resource: "invoice", Action: "delete",
	})
	if allowed {
		t.Error("harusnya ditolak, tapi malah diizinkan")
	}
}

func TestCan_WildcardPermission(t *testing.T) {
	ctx := context.Background()
	e := setup(t)

	_ = e.CreateRole(ctx, "admin", []model.Permission{"invoice:*"})
	_ = e.AssignRole(ctx, "user-1", "admin")

	allowed, _ := e.Can(ctx, model.AccessRequest{
		SubjectID: "user-1", Resource: "invoice", Action: "delete",
	})
	if !allowed {
		t.Error("wildcard admin harusnya boleh semua action")
	}
}

func TestCan_RevokedRoleLosesAccess(t *testing.T) {
	ctx := context.Background()
	e := setup(t)

	_ = e.CreateRole(ctx, "editor", []model.Permission{"invoice:write"})
	_ = e.AssignRole(ctx, "user-1", "editor")
	_ = e.RevokeRole(ctx, "user-1", "editor")

	allowed, _ := e.Can(ctx, model.AccessRequest{
		SubjectID: "user-1", Resource: "invoice", Action: "write",
	})
	if allowed {
		t.Error("setelah revoke harusnya ditolak")
	}
}
