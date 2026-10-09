# SIAKAD Mini — UTS Pemrograman Backend Lanjut

RESTful API backend untuk layanan akademik sederhana: data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibuat dengan **Go + Fiber v3 + PostgreSQL**.

## Menjalankan

```bash
# 1. siapkan database
psql -c "CREATE DATABASE siakad_mini;"
psql -d siakad_mini -f migrations/001_siakad.sql   # schema + seeder

# 2. siapkan environment
cp .env.example .env    # sesuaikan DB_USER/DB_PASSWORD/JWT_SECRET

# 3. jalankan
go run .
```

Server berjalan di `http://localhost:3000` (konfigurasi `APP_PORT`).

## Akun Seeder

| Peran | Email | Password |
|---|---|---|
| admin | `admin@siakad.ac.id` | `admin12345` |
| mahasiswa (20 akun) | `187221000001@student.unair.ac.id` s.d. `187221000020@student.unair.ac.id` | NIM masing-masing, contoh `187221000001` |

Seeder berisi 1 admin, 20 mahasiswa, dan 10 mata kuliah. Mata kuliah `SD201` sengaja berkuota 2 untuk menguji kuota penuh.

## Autentikasi

- Semua endpoint kecuali login memerlukan header `Authorization: Bearer <access_token>`.
- Token didapat dari `POST /api/v1/auth/login`, berlaku sesuai `JWT_ACCESS_TTL_MINUTES` (default 120 menit).
- Login dibatasi 5 percobaan per menit per IP (429).

## Business Rule KRS

Batas SKS per tahun akademik berdasarkan IPK terakhir:

| IPK terakhir | Batas SKS |
|---|---|
| ≥ 3,00 | 24 |
| 2,50 – 2,99 | 21 |
| < 2,50 | 18 |

Selain itu: mata kuliah yang sama tidak boleh diambil dua kali pada tahun akademik sama (409), kuota penuh tidak dapat diambil (422), dan mahasiswa hanya dapat mengakses KRS miliknya (403).

---

## Dokumentasi Endpoint

Semua response berformat JSON dengan struktur seragam:

```json
// sukses
{ "success": true, "message": "...", "data": ..., "meta": ... }

// gagal
{ "success": false, "message": "...", "errors": { "field": ["pesan"] } }
```

Status code yang dipakai: `200, 201, 204, 400, 401, 403, 404, 409, 422, 429, 500` (500 tidak membocorkan detail teknis; detail hanya tercatat di log).

### 1. POST /api/v1/auth/login — publik

Login, mengembalikan access token.

```json
{ "email": "admin@siakad.ac.id", "password": "admin12345" }
```

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOi...",
    "token_type": "Bearer",
    "expires_in": 7200,
    "user": { "id": 1, "email": "admin@siakad.ac.id", "role": "admin" }
  }
}
```

Error: `401` kredensial salah (pesan sama untuk email tak terdaftar / password salah / akun di-soft delete), `422` validasi (email format, password min 8), `429` lebih dari 5x per menit.

### 2. GET /api/v1/auth/me — semua role

Profil pengguna yang sedang login. Untuk mahasiswa menyertakan data akademiknya.

```json
{
  "success": true,
  "message": "Profil berhasil diambil",
  "data": {
    "id": 2, "email": "187221000001@student.unair.ac.id", "role": "mahasiswa",
    "student": { "id": 1, "nim": "187221000001", "nama": "Rina Putri",
                 "prodi": "Sistem Informasi", "angkatan": 2022, "ipk_terakhir": 3.45 }
  }
}
```

Error: `401` token tidak ada/salah/kedaluwarsa.

### 3. GET /api/v1/students — admin

Daftar mahasiswa (yang belum di-soft delete) dengan pagination, filter, dan search.

Query parameter:

| Parameter | Default | Keterangan |
|---|---|---|
| `page` | 1 | nomor halaman |
| `per_page` | 10 | maks 50 |
| `prodi` | — | filter persis, mis. `Teknik Informatika` |
| `angkatan` | — | filter persis, mis. `2023` |
| `search` | — | cari NIM atau nama (ILIKE) |
| `sort` | `nama` | `nama` (A-Z) atau `-ipk_terakhir` (IPK tertinggi dulu) |

```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [
    { "id": 1, "nim": "187221000001", "nama": "Rina Putri", "prodi": "Sistem Informasi",
      "angkatan": 2022, "ipk_terakhir": 3.45, "email": "187221000001@student.unair.ac.id" }
  ],
  "meta": { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
}
```

Error: `401`, `403` bukan admin.

### 4. POST /api/v1/students — admin

Menambah mahasiswa sekaligus akun user-nya dalam satu transaction. Password awal = NIM (disimpan sebagai hash bcrypt).

```json
{
  "nim": "187221000099",
  "nama": "Mahasiswa Baru",
  "email": "187221000099@student.unair.ac.id",
  "prodi": "Sains Data",
  "angkatan": 2026,
  "ipk_terakhir": 3.10
}
```

Aturan: `nim` wajib 12 digit angka & unik, `nama` wajib, `email` wajib format email & unik, `angkatan` 4 digit tidak melebihi tahun berjalan, `ipk_terakhir` opsional 0,00–4,00.

Error: `422` duplikat/validasi (contoh di bawah), `403`.

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": { "nim": ["NIM sudah terdaftar"], "email": ["Email sudah terdaftar"] }
}
```

### 5. GET /api/v1/students/{id} — admin / mahasiswa (data sendiri)

Detail mahasiswa beserta KRS-nya.

```json
{
  "success": true,
  "message": "Detail mahasiswa berhasil diambil",
  "data": {
    "id": 1, "nim": "187221000001", "nama": "Rina Putri", "prodi": "Sistem Informasi",
    "angkatan": 2022, "ipk_terakhir": 3.45, "email": "187221000001@student.unair.ac.id",
    "mata_kuliah": [
      { "id": 7, "tahun_akademik": "2026/2027-Ganjil", "kode_mk": "SD201",
        "nama_mk": "Data Mining", "sks": 3, "semester": 4, "created_at": "..." }
    ],
    "total_sks": 3,
    "batas_sks": 24
  }
}
```

`batas_sks` dihitung dari `ipk_terakhir` sesuai business rule. Error: `403` mahasiswa mengakses data orang lain, `404` tidak ditemukan / sudah di-soft delete.

### 6. PUT /api/v1/students/{id} — admin

Memperbarui data mahasiswa. Field sama dengan endpoint 4 kecuali `nim` dan `email` (tidak dapat diubah).

```json
{ "nama": "Rina Putri Utami", "prodi": "Sistem Informasi", "angkatan": 2022, "ipk_terakhir": 3.50 }
```

Error: `404`, `422`, `403`.

### 7. DELETE /api/v1/students/{id} — admin

Soft delete (mengisi `deleted_at`). Mahasiswa yang dihapus tidak muncul di endpoint 3, tidak dapat dilihat lewat endpoint 5, dan tidak dapat login.

Sukses: `204` tanpa body. Error: `404`, `403`.

### 8. GET /api/v1/courses — semua role

Daftar mata kuliah beserta `terisi` dan `sisa_kuota` yang dihitung dari tabel enrollments.

Query parameter: `semester` (1–8), `search` (kode_mk atau nama_mk), `available=true` (hanya yang kuotanya belum penuh).

```json
{
  "success": true,
  "message": "Data mata kuliah berhasil diambil",
  "data": [
    { "id": 10, "kode_mk": "SD201", "nama_mk": "Data Mining", "sks": 3,
      "semester": 4, "kuota": 2, "terisi": 2, "sisa_kuota": 0 }
  ],
  "meta": { "current_page": 1, "per_page": 10, "total": 10, "last_page": 1 }
}
```

Error: `401`.

### 9. POST /api/v1/enrollments — mahasiswa

Mengambil mata kuliah (menambah KRS).

```json
{ "course_id": 5, "tahun_akademik": "2026/2027-Ganjil" }
```

`tahun_akademik` wajib format `YYYY/YYYY+1-Ganjil|Genap`. Seluruh pemeriksaan berjalan dalam satu transaction dengan row locking (`SELECT ... FOR UPDATE`) pada baris course, sehingga dua mahasiswa tidak dapat merebut kuota terakhir secara bersamaan.

Error:

| Status | Situasi | Contoh pesan |
|---|---|---|
| `409` | sudah pernah diambil tahun akademik sama | `Mata kuliah IF101 sudah pernah diambil pada tahun akademik 2026/2027-Ganjil` |
| `422` | course tidak ada | `errors.course_id: ["Mata kuliah tidak ditemukan"]` |
| `422` | kuota penuh | `errors.course_id: ["Kuota mata kuliah SD201 penuh"]` |
| `422` | melebihi batas SKS | `errors.course_id: ["Total SKS melebihi batas: sisa SKS Anda 2, mata kuliah ini 3 SKS"]` |
| `403` | pemanggil bukan mahasiswa | |

### 10. DELETE /api/v1/enrollments/{id} — mahasiswa (milik sendiri)

Membatalkan mata kuliah dari KRS; kuota mata kuliah kembali bertambah.

Sukses: `204` tanpa body. Error: `403` enrollment milik mahasiswa lain, `404` tidak ditemukan.

---

## Pengujian dengan Postman

Folder ini dilengkapi `UTS.postman_collection.json` dan `UTS.postman_environment.json` (di luar git). Import keduanya, pilih environment "UTS SIAKAD - Local", lalu jalankan request "Login admin" / "Login mahasiswa" lebih dulu — test script-nya otomatis menyimpan token ke variabel `admin_token` / `mhs_token` / `mhs7_token`. Koleksi mencakup seluruh 10 endpoint beserta kasus gagalnya (401/403/404/409/422/429).

Catatan: login dibatasi 5 percobaan per menit per IP — bila mendapat 429 saat menguji, tunggu sebentar.

## Struktur Proyek

```
UTS/
├── app/
│   ├── model/        # struct + request + response envelope
│   ├── repository/   # query PostgreSQL (users, students, courses, enrollments)
│   └── service/      # business rule + handler
├── config/           # fiber app, env, logger, error handler terpusat
├── database/         # koneksi pgxpool
├── helper/           # jwt, bcrypt, validator, response, context
├── middleware/       # auth, role, rate limiter, access log
├── migrations/       # schema + seeder
└── route/            # peta 10 endpoint
```
