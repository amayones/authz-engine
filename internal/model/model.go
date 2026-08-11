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

// RelationTuple adalah unit dasar ReBAC: pernyataan bahwa `Subject`
// memiliki `Relation` terhadap `Object`.
//
// Format Object dan Subject biasanya "type:id", misal "document:123".
// Subject bisa juga berupa userset: "group:eng#member" — artinya semua
// subject yang punya relation "member" pada object "group:eng".
type RelationTuple struct {
	Object   string
	Relation string
	Subject  string
}

// RelationSchema mendefinisikan hierarki relasi. Key adalah nama relasi,
// value adalah daftar relasi lain yang JUGA memenuhi relasi ini.
//
// Contoh: RelationSchema{"viewer": {"editor", "owner"}} berarti siapa pun
// yang punya relasi "editor" atau "owner" otomatis dianggap "viewer" juga.
type RelationSchema map[string][]string
