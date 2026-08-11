// File ini berisi logic ReBAC: evaluasi relationship graph, termasuk
// dukungan userset (grup) dan hierarki relasi.
package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/amayones/authz-engine/internal/model"
)

// SetSchema mengatur hierarki relasi. Contoh:
//
//	e.SetSchema(model.RelationSchema{
//	    "viewer": {"editor", "owner"}, // editor & owner otomatis viewer
//	    "editor": {"owner"},           // owner otomatis editor
//	})
func (e *Engine) SetSchema(schema model.RelationSchema) {
	e.schema = schema
}

// WriteRelation membuat satu relation tuple, misal:
// WriteRelation(ctx, "document:123", "owner", "user:amayones")
func (e *Engine) WriteRelation(ctx context.Context, object, relation, subject string) error {
	if err := e.store.WriteTuple(ctx, model.RelationTuple{
		Object: object, Relation: relation, Subject: subject,
	}); err != nil {
		return err
	}
	if e.cache != nil {
		e.cache.Clear()
	}
	return nil
}

// DeleteRelation menghapus satu relation tuple.
func (e *Engine) DeleteRelation(ctx context.Context, object, relation, subject string) error {
	if err := e.store.DeleteTuple(ctx, model.RelationTuple{
		Object: object, Relation: relation, Subject: subject,
	}); err != nil {
		return err
	}
	if e.cache != nil {
		e.cache.Clear()
	}
	return nil
}

// CheckRelation menjawab pertanyaan inti ReBAC: apakah `subject` memiliki
// `relation` terhadap `object`?
//
// Contoh: CheckRelation(ctx, "document:123", "viewer", "user:bob")
// akan bernilai true kalau bob adalah viewer langsung, ATAU editor/owner
// (lewat hierarki schema), ATAU member dari grup yang jadi viewer
// (lewat userset traversal).
func (e *Engine) CheckRelation(ctx context.Context, object, relation, subject string) (bool, error) {
	key := "rel:" + object + "|" + relation + "|" + subject
	if e.cache != nil {
		if val, found := e.cache.Get(key); found {
			return val, nil
		}
	}

	allowed, err := e.checkRelation(ctx, object, relation, subject, make(map[string]bool))
	if err != nil {
		return false, err
	}

	if e.cache != nil {
		e.cache.Set(key, allowed)
	}
	return allowed, nil
}

func (e *Engine) checkRelation(ctx context.Context, object, relation, subject string, visited map[string]bool) (bool, error) {
	// Cegah infinite loop kalau relationship graph punya siklus.
	key := object + "#" + relation
	if visited[key] {
		return false, nil
	}
	visited[key] = true

	// Relasi yang harus dicek: relasi itu sendiri, plus semua relasi lain
	// yang menurut schema juga memenuhi relasi ini (hierarki).
	relationsToCheck := append([]string{relation}, e.schema[relation]...)

	for _, rel := range relationsToCheck {
		tuples, err := e.store.ReadTuples(ctx, object, rel)
		if err != nil {
			return false, fmt.Errorf("engine: gagal baca relation tuples: %w", err)
		}

		for _, t := range tuples {
			// Match langsung.
			if t.Subject == subject {
				return true, nil
			}

			// Userset: subject berformat "object#relation", misal
			// "group:eng#member" — perlu traversal rekursif ke grup itu.
			if objPart, relPart, ok := strings.Cut(t.Subject, "#"); ok {
				allowed, err := e.checkRelation(ctx, objPart, relPart, subject, visited)
				if err != nil {
					return false, err
				}
				if allowed {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
