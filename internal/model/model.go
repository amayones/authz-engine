// Package model berisi tipe data inti untuk authorization engine.
package model

// Subject merepresentasikan entitas yang meminta akses (biasanya user,
// tapi bisa juga service-account atau API key).
type Subject struct {
	ID    string
	Roles []string
}

// Permission merepresentasikan satu izin atomik dalam format
// "resource:action", misalnya "invoice:read" atau "invoice:delete".
type Permission string

// Attributes adalah kumpulan key-value milik subject atau context request,
// dipakai untuk evaluasi ABAC (misal: department, level, region).
type Attributes map[string]string

// Condition adalah syarat ABAC sederhana: sebuah attribute key harus
// sama dengan value tertentu.
type Condition struct {
	AttrKey   string
	AttrValue string
}

// Role merepresentasikan kumpulan permission yang bisa di-assign ke Subject.
// Conditions opsional — kosong berarti role ini murni RBAC tanpa syarat
// tambahan.
type Role struct {
	Name        string
	Permissions []Permission
	Conditions  []Condition
}

// AccessRequest adalah input untuk pertanyaan "apakah subject ini boleh
// melakukan action pada resource ini?"
type AccessRequest struct {
	SubjectID string
	Resource  string
	Action    string
	Context   Attributes // atribut tambahan untuk evaluasi ABAC
}

// Permission gabungan dari Resource + Action, dipakai untuk pencocokan cepat.
func (r AccessRequest) Permission() Permission {
	return Permission(r.Resource + ":" + r.Action)
}
