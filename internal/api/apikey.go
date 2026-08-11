package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateAPIKey membuat key mentah acak 32 byte (256 bit), di-encode
// hex jadi string 64 karakter. Ini yang diberikan ke client sekali saja
// saat pembuatan — server hanya menyimpan hash-nya.
func generateAPIKey() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("gagal generate random key: %w", err)
	}
	return "ak_" + hex.EncodeToString(raw), nil
}

// hashAPIKey menghasilkan SHA-256 hex dari key mentah, dipakai baik saat
// penyimpanan maupun saat verifikasi request masuk.
func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// GenerateAndHashKey membuat key mentah baru sekaligus hash-nya.
// Dipakai oleh admin tool (cmd/genkey) untuk provisioning client baru.
func GenerateAndHashKey() (rawKey, hash string, err error) {
	rawKey, err = generateAPIKey()
	if err != nil {
		return "", "", err
	}
	return rawKey, hashAPIKey(rawKey), nil
}
