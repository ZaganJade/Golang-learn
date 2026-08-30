# API Students — Tugas 3 Praktikum Pemrograman Backend Lanjut

API mahasiswa dari tugas 2, sekarang datanya pindah dari memori ke PostgreSQL lewat pola repository. Perilaku HTTP-nya tidak berubah; yang berubah hanya tempat datanya disimpan.

## Skema tabel

```sql
CREATE TABLE IF NOT EXISTS students (
    id         SERIAL       PRIMARY KEY,
    nim        VARCHAR(20)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    grade      NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key ON students (nim);
CREATE INDEX IF NOT EXISTS students_grade_idx ON students (grade);
```

Alasan pemilihan tipe dan indeks:

- **NIM unik dijaga `students_nim_key` di basis data**, bukan lewat pengecekan di Go. Alasannya: pengecekan di aplikasi punya celah race condition — dua permintaan yang datang hampir bersamaan bisa sama-sama lolos pengecekan sebelum salah satunya menyimpan. Batasan UNIQUE di basis data tidak punya celah itu karena basis datalah yang memutuskan siapa yang lebih dulu. Ditambah lagi, kalau ada dua aplikasi yang nulis ke tabel yang sama, batasan ini tetap bekerja.
- **Indeks `students_grade_idx`** dipakai untuk filter `grade_min`/`grade_max` pada endpoint daftar. Query rentang (`grade >= $1 AND grade <= $2`) bisa memakai indeks B-tree ini, jadi penyaringan tidak harus memindai seluruh tabel.
- **`NUMERIC(5,2)`** untuk grade karena nilainya pasti (tidak ada pembulatan floating point), maksimal 999.99 — cukup untuk skala 0-100.

## Menyiapkan basis data dari nol

Anggap pembacanya rekan sekelas yang baru mengklona repositori ini:

```bash
# 1. Buat database (contoh pakai psql)
psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE praktikum_backend;"

# 2. Jalankan migrasi
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/001_create_students.sql

# 3. Salin contoh konfigurasi lalu isi nilainya
cp .env.example .env

# 4. Jalankan
go mod tidy
go run .
```

Migrasi ini sekalian mengisi 5 mahasiswa awal (Amba, Rusdi, Mas Fuad, Azril, Mas Gatot) dengan `ON CONFLICT (nim) DO NOTHING`, jadi aman dijalankan berulang.

## Variabel environment

| Variabel | Arti | Contoh |
|---|---|---|
| `APP_PORT` | Port server | `3000` |
| `DB_HOST` | Alamat PostgreSQL | `127.0.0.1` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User basis data | `postgres` |
| `DB_PASSWORD` | Kata sandi basis data (kosong kalau auth-nya trust) | |
| `DB_NAME` | Nama database | `praktikum_backend` |
| `DB_SSLMODE` | Mode SSL | `disable` |
| `DB_MAX_CONNS` | Koneksi maksimum di pool | `10` |

Berkas `.env` tidak ikut ter-commit (ada di `.gitignore`); yang ter-commit hanya `.env.example` dengan nilai kosong.

## Kontrak API

| Metode | Endpoint | Parameter | Contoh body | Status yang mungkin | Contoh respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | query: `page` (bawaan 1), `limit` (bawaan 10, maks 50), `search` (nama, ILIKE), `sort` (`id`/`name`/`nim`/`grade`/`created_at`), `order` (`asc`/`desc`), `is_active`, `grade_min`, `grade_max` | — | `200`, `500` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[...],"meta":{"page":1,"limit":10,"total":5,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | path: `id` | — | `200`, `400`, `404`, `500` | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"name":"Amba","nim":"230101","grade":90.00,"is_active":true,"created_at":"..."}}` |
| POST | `/api/v1/students` | header `Content-Type: application/json` | `{"name":"Amba","nim":"230101","grade":90}` | `201` (+`Location`), `400`, `409`, `415`, `422`, `500` | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{...}}` |
| PUT | `/api/v1/students/:id` | path: `id`; semua field wajib | `{"name":"Amba","nim":"230101","grade":95,"is_active":true}` | `200`, `400`, `404`, `409`, `415`, `422`, `500` | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya","data":{...}}` |
| PATCH | `/api/v1/students/:id` | path: `id`; hanya field yang dikirim | `{"grade":100}` | `200`, `400`, `404`, `409`, `415`, `422`, `500` | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian","data":{...}}` |
| DELETE | `/api/v1/students/:id` | path: `id` | — | `204` (tanpa body), `400`, `404`, `500` | `204`: (kosong) |
| GET | `/api/v1/health` | — | — | `200`, `503` | `200`: `{"success":true,"message":"server dan database berjalan"}` |

Asal status sekarang bisa dilacak: `404` berasal dari `pgx.ErrNoRows`/`RowsAffected() == 0` yang diterjemahkan repository jadi `ErrNotFound`, `409` dari pelanggaran UNIQUE (kode PostgreSQL 23505) yang diterjemahkan jadi `ErrDuplicate`, dan `503` dipakai `/health` saat PostgreSQL tidak bisa dihubungi — layanan masih hidup tapi tidak bisa dipakai, jadi bukan 500.

## Struktur berkas

- `main.go` — konfigurasi, koneksi database, middleware, dan route
- `handler.go` — handler tiap endpoint sebagai method dari `StudentHandler`
- `helper.go` — amplop respons dan pembacaan query string
- `app/model/student.go` — struct entitas, request, dan respons
- `app/repository/student_repository.go` — kontrak dan implementasi akses data
- `config/env.go` — pembaca variabel environment
- `database/postgres.go` — connection pool
- `migrations/001_create_students.sql` — skema tabel + data awal
