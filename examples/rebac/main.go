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

	// Hierarki: owner otomatis editor, editor otomatis viewer.
	e.SetSchema(model.RelationSchema{
		"viewer": {"editor", "owner"},
		"editor": {"owner"},
	})

	// amayones adalah owner document:proposal.
	if err := e.WriteRelation(ctx, "document:proposal", "owner", "user:amayones"); err != nil {
		log.Fatal(err)
	}

	// document:proposal di-share ke semua member group:engineering sebagai viewer.
	if err := e.WriteRelation(ctx, "document:proposal", "viewer", "group:engineering#member"); err != nil {
		log.Fatal(err)
	}

	// bob adalah member group:engineering.
	if err := e.WriteRelation(ctx, "group:engineering", "member", "user:bob"); err != nil {
		log.Fatal(err)
	}

	// Berbagai pengecekan:
	checks := []struct {
		subject  string
		relation string
	}{
		{"user:amayones", "owner"},  // true, langsung
		{"user:amayones", "viewer"}, // true, lewat hierarki (owner -> viewer)
		{"user:bob", "viewer"},      // true, lewat userset (member grup)
		{"user:bob", "owner"},       // false, bob cuma viewer, bukan owner
		{"user:charlie", "viewer"},  // false, charlie tidak terkait sama sekali
	}

	for _, c := range checks {
		allowed, err := e.CheckRelation(ctx, "document:proposal", c.relation, c.subject)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s has %q on document:proposal -> %v\n", c.subject, c.relation, allowed)
	}
}
