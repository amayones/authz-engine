# Panduan Deployment Gratis Tanpa Kartu Kredit

Panduan ini menjelaskan cara hosting `authz-engine` (aplikasi Go) secara **gratis tanpa kartu kredit** menggunakan:

1. **Supabase** — database PostgreSQL gratis (500MB)
2. **Replit** — hosting & menjalankan aplikasi Go gratis (import dari GitHub)

> **Catatan penting**: Ada banyak platform bernama "Glitch" (mis. glitch.com, glitchonline.com). Untuk **aplikasi Go**, platform gratis tanpa kartu kredit yang **benar-benar berfungsi** adalah **Replit**. Glitch.com lebih cocok untuk Node.js/HTML, bukan Go.

---

## Ringkasan

| Komponen | Platform | Gratis | Tanpa Kartu Kredit |
|----------|----------|--------|-------------------|
| Database PostgreSQL | [Supabase](https://supabase.com) | 500MB | ✅ |
| Hosting App Go | [Replit](https://replit.com) | ✅ | ✅ |
| HTTPS Otomatis | Replit | ✅ | ✅ |

---

## Langkah 1: Persiapkan Proyek di GitHub

### 1.1 Push proyek ke GitHub

Buka terminal di folder proyek Anda:

```bash
cd c:\Apache2462\htdocs\authz-engine

# Tambahkan semua perubahan (termasuk dukungan PostgreSQL yang sudah ditambahkan)
git add .
git commit -m "tambah dukungan PostgreSQL + panduan deployment"
git push origin main
```

> Jika repo belum ada di GitHub, buat dulu di https://github.com/new lalu hubungkan:
> ```bash
> git remote add origin https://github.com/USERNAME/authz-engine.git
> git branch -M main
> git push -u origin main
> ```

---

## Langkah 2: Buat Database PostgreSQL Gratis di Supabase

### 2.1 Daftar Supabase

1. Buka https://supabase.com
2. Klik **Start your project**
3. Daftar dengan **GitHub** atau **Email** (bukan kartu kredit)

### 2.2 Buat Project

1. Setelah login, klik **New project**
2. Isi form:
   - **Name**: `authz-engine` (atau nama lain)
   - **Database Password**: buat password kuat, simpan baik-baik
   - **Region**: pilih **Southeast Asia (Singapore)** — terdekat dengan Indonesia
3. Klik **Create new project**

### 2.3 Ambil Connection String

1. Tunggu project selesai dibuat (±2 menit)
2. Di sidebar kiri, klik **Settings** → **Database**
3. Scroll ke bagian **Connection string**
4. Pilih **URI** (bukan Pooler)
5. Salin connection string, formatnya seperti ini:

   ```
   postgresql://postgres.XXXXX:YOUR-PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

   > **PENTING**: Ganti `YOUR-PASSWORD` dengan password database yang tadi Anda buat.
   > Simpan connection string ini — akan dipakai di Langkah 3.

---

## Langkah 3: Deploy Aplikasi ke Replit

### 3.1 Daftar Replit

1. Buka https://replit.com
2. Klik **Sign up**
3. Daftar dengan **GitHub** atau **Email** (bukan kartu kredit)

### 3.2 Import Proyek dari GitHub

1. Setelah login, klik **Create** (atau **New Repl**)
2. Pilih **Import from GitHub**
3. Masukkan repo Anda: `amayones/authz-engine`
4. Replit akan otomatis mendeteksi proyek Go dan melakukan build

### 3.3 Set Environment Variables (Secrets)

1. Di panel Replit (workspace), cari **Secrets** (ikon kunci/gembok di sidebar kiri)
2. Tambahkan secret berikut:

   | Key | Value |
   |-----|-------|
   | `AUTHZ_DB_DRIVER` | `postgres` |
   | `AUTHZ_DB_CONN` | *(isi dengan connection string dari Supabase, Langkah 2.3)* |
   | `AUTHZ_ADDR` | `:8080` |
   | `AUTHZ_AUTO_MIGRATE` | `true` |

   Contoh value `AUTHZ_DB_CONN`:
   ```
   postgresql://postgres.abc123:password123@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

3. Pastikan semua secret tersimpan

### 3.4 Jalankan Aplikasi

1. Klik tombol **Run** di Replit
2. Replit akan build dan menjalankan `cmd/server`
3. Aplikasi akan berjalan di URL seperti: `https://nama-repl.your-username.repl.co`

> **Catatan**: Replit mungkin perlu file konfigurasi agar tahu cara menjalankan Go app. Jika belum ada, Replit biasanya mendeteksi `go.mod` dan menjalankan `go run ./...` otomatis. Jika tidak, buat file `replit.nix` dan `.replit` (lihat bagian konfigurasi di bawah).

### 3.5 Konfigurasi Tambahan (jika perlu)

Jika Replit tidak otomatis menjalankan server, buat file `.replit` di root proyek:

```
run = "go build -o /tmp/authz-server ./cmd/server && /tmp/authz-server"

[deployment]
run = ["sh", "-c", "go build -o /tmp/authz-server ./cmd/server && /tmp/authz-server"]

[languages]
[languages.go]
pattern = "**/*.go"
```

> **Catatan**: Menggunakan `go build` (compile sekali) lalu jalankan binary jauh lebih cepat daripada `go run` yang compile ulang setiap kali — penting untuk tier gratis Replit.

Dan file `replit.nix`:

```nix
{ pkgs }: {
  deps = [
    pkgs.go
    pkgs.git
  ];
}
```

> **PENTING**: Gunakan `pkgs.go` (bukan `pkgs.go_1_26`). Versi `go_1_26` tidak tersedia di channel nixpkgs yang dipakai Replit (stable-25_05) dan akan menyebabkan error `couldn't get nix env`.
>
> Jangan tambahkan section `[nix] channel = ...` di `.replit` karena akan menimpa channel default Replit.

Lalu klik **Run** lagi.

### 3.6 Verifikasi

Buka URL aplikasi Anda di browser:

```
https://nama-repl.your-username.repl.co/health
```

Response yang benar:

```json
{"status":"ok"}
```

> Jika muncul `{"error":"..."}` atau halaman error, cek tab **Console/Logs** di Replit untuk melihat error.

---

## Langkah 4: Buat API Key

API key dibuat dengan menjalankan `cmd/genkey` di komputer lokal. Anda perlu `go` terpasang.

### 4.1 Jalankan genkey

```bash
cd c:\Apache2462\htdocs\authz-engine

# Ganti <CONNECTION_STRING> dengan connection string Supabase Anda
go run ./cmd/genkey \
  -driver postgres \
  -db "postgresql://postgres.abc123:password123@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres" \
  -name "client1" \
  -rpm 120
```

Output:

```
API key berhasil dibuat. SIMPAN SEKARANG — tidak akan ditampilkan lagi:
ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```

> **PENTING**: Simpan API key ini dengan aman — hanya ditampilkan sekali.

---

## Langkah 5: Uji API

### 5.1 Buat role

```bash
curl -X POST https://nama-repl.your-username.repl.co/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "editor",
    "permissions": ["invoice:read", "invoice:write"]
  }'
```

### 5.2 Assign role ke user

```bash
curl -X POST https://nama-repl.your-username.repl.co/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin (Can)

```bash
curl -X POST https://nama-repl.your-username.repl.co/can \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "resource": "invoice", "action": "read"}'
```

Response:

```json
{"allowed":true}
```

---

## Troubleshooting

### Error: `connection refused` saat migration

Pastikan connection string Supabase benar dan password tidak mengandung karakter khusus yang perlu di-encode. Jika password mengandung `@` atau `:` gunakan URL-encoding.

### Error: `role "postgres" does not exist`

Pastikan Anda memilih **URI** (bukan **Transaction** / **Pooler**) di Supabase. Untuk production, bisa gunakan **Transaction pooler**.

### App mati (idle) di Replit

Replit gratis akan mematikan app setelah tidak digunakan (idle). Saat diakses lagi, akan restart otomatis. Ini normal untuk tier gratis.

### Replit tidak menjalankan server Go

Pastikan file `.replit` dan `replit.nix` ada (lihat Langkah 3.5), atau gunakan tombol **Run** manual.

### Error: `Security scan skipped: connection lost`

Ini **bukan error dari kode Anda** — ini masalah infrastruktur Replit. Build berhenti karena koneksi Replit ke server build terputus. Penyebab umum:

1. **Dependensi terlalu besar** — `go.mod` sebelumnya menarik ratusan package (MySQL, SQLite, MongoDB, dll.) karena directive `tool github.com/golang-migrate/migrate/v4/cmd/migrate`. Ini membuat build sangat lambat dan timeout.
   - **Solusi**: Sudah diperbaiki — directive `tool` dihapus dan `go mod tidy` dijalankan. `go.mod` sekarang hanya 38 baris (sebelumnya 197 baris).

2. **Versi Go tidak tersedia di Replit** — `go 1.26.5` di `go.mod` mungkin tidak tersedia persis di Replit.
   - **Solusi**: `replit.nix` menggunakan `pkgs.go` (default) yang pasti tersedia di semua channel nixpkgs.

3. **Koneksi internet Replit tidak stabil** — kadang build gagal karena jaringan.
   - **Solusi**: Klik **Run** lagi. Jika masih gagal, coba **Stop** lalu **Run** ulang, atau refresh halaman.

4. **Cache build rusak** — Replit menyimpan cache build yang kadang korup.
   - **Solusi**: Di Replit, buka **Shell** dan jalankan:
     ```bash
     rm -rf ~/.cache/go-build
     go clean -cache
     ```
     Lalu klik **Run** lagi.

5. **Replit sedang down** — cek status di https://status.replit.com

### Error: `couldn't get nix env` / `evaluating file '<nix/derivation-internal.nix>'`

Error ini terjadi karena **konfigurasi `replit.nix` atau `.replit` salah**. Penyebab umum:

1. **`pkgs.go_1_26` tidak tersedia** di channel nixpkgs Replit (stable-25_05).
   - **Solusi**: Gunakan `pkgs.go` (default) yang pasti tersedia di semua channel:
     ```nix
     { pkgs }: {
       deps = [
         pkgs.go
         pkgs.git
       ];
     }
     ```

2. **Section `[nix] channel = ...` di `.replit`** menimpa channel default Replit dengan channel yang tidak tersedia.
   - **Solusi**: Hapus section `[nix]` dari `.replit` dan biarkan Replit memakai channel default-nya.

3. **Cache nix korup** di workspace Replit.
   - **Solusi**: Di Replit, buka **Shell**, lalu jalankan:
     ```bash
     rm -rf ~/.nix-profile 2>/dev/null
     ```
     Lalu klik **Stop** dan **Run** ulang.

4. Jika masih gagal, **buat Repl baru** (Import from GitHub) dengan repo yang sama — ini membersihkan semua cache lama.

**Langkah yang harus dilakukan setelah perbaikan ini:**

1. Push perubahan ke GitHub:
   ```bash
   git add .
   git commit -m "perbaiki konfigurasi Replit: kurangi dependensi, tambah nix channel"
   git push origin main
   ```

2. Di Replit, buka **Shell** dan jalankan:
   ```bash
   rm -rf ~/.cache/go-build
   go clean -cache
   ```

3. Klik **Run** lagi.

4. Jika masih gagal, buat **Repl baru** (Import from GitHub) dengan repo yang sama — ini membersihkan semua cache lama.

---

## Catatan Penting

- **Supabase** gratis 500MB — cukup untuk aplikasi kecil. Bisa upgrade kapan saja.
- **Replit** gratis dengan limit waktu aktif — mati saat idle, nyala lagi saat diakses.
- **Migration otomatis** berjalan saat app start (`AUTHZ_AUTO_MIGRATE=true`). Setelah pertama kali, table sudah dibuat.
- Untuk development lokal, Anda bisa pakai PostgreSQL lokal atau Docker.

---

## Alternatif Lain (Juga Tanpa Kartu Kredit)

| Kombinasi | Database | App |
|-----------|----------|-----|
| Neon (512MB) + Replit | [Neon](https://neon.tech) | [Replit](https://replit.com) |
| Supabase (500MB) + Railway ($5 kredit) | [Supabase](https://supabase.com) | [Railway](https://railway.app) |
| Aiven (1GB) + Deta Space | [Aiven](https://aiven.io) | [Deta Space](https://deta.space) |
| Supabase (500MB) + Fly.io (perlu kartu) | [Supabase](https://supabase.com) | [Fly.io](https://fly.io) |

> **Jujur**: Untuk aplikasi Go gratis **tanpa kartu kredit**, pilihan hosting sangat terbatas. **Replit** adalah yang paling andal dan mendukung Go. Alternatif lain (Railway, Fly.io, Render) umumnya memerlukan kartu kredit untuk verifikasi, bahkan di tier gratis.
