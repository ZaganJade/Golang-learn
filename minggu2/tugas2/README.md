# API Students — Tugas 2 Praktikum Pemrograman Backend Lanjut

REST API CRUD mahasiswa pakai Fiber v3. Data disimpan di memori (hilang saat server mati).

## Cara jalanin

```bash
cd minggu2/tugas2
go run .
# server jalan di http://localhost:3000
```

## Kontrak API

| Metode | Endpoint | Parameter | Contoh body | Status yang mungkin | Contoh respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | query: `page` (bawaan 1), `limit` (bawaan 10, maks 50), `search` (cari nama, tidak peduli huruf besar/kecil), `sort` (`id`/`name`/`nim`/`grade`/`created_at`, bawaan `id`), `order` (`asc`/`desc`, bawaan `asc`), `is_active` (`true`/`false`), `grade_min`, `grade_max` | — | `200` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[...],"meta":{"page":1,"limit":10,"total":3,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | path: `id` (angka positif) | — | `200`, `400` (id bukan angka), `404` | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"name":"Zagan Jade","nim":"230101","grade":92.5,"is_active":true,"created_at":"..."}}` |
| POST | `/api/v1/students` | header: `Content-Type: application/json` (wajib) | `{"name":"Zagan Jade","nim":"230101","grade":92.5}` | `201` (+ header `Location`), `400` (body bukan JSON valid), `409` (NIM sudah dipakai), `415` (Content-Type bukan JSON), `422` (validasi gagal) | `201`: `{"success":true,"message":"mahasiswa berhasil dibuat","data":{...}}` + header `Location: /api/v1/students/1` |
| PUT | `/api/v1/students/:id` | path: `id`; header: `Content-Type: application/json` | `{"name":"Zagan Baru","nim":"230999","grade":95,"is_active":false}` — **semua field wajib** karena PUT mengganti seluruh isi | `200`, `400`, `404`, `409`, `415`, `422` | `200`: `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya","data":{...}}` |
| PATCH | `/api/v1/students/:id` | path: `id`; header: `Content-Type: application/json` | `{"grade":100}` — hanya field yang dikirim yang berubah | `200`, `400` (tidak ada field), `404`, `409`, `415`, `422` | `200`: `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian","data":{...}}` |
| DELETE | `/api/v1/students/:id` | path: `id` | — | `204` (tanpa body), `400`, `404` | `204`: (kosong) |
| GET | `/api/v1/health` | — | — | `200` | `{"success":true,"message":"server berjalan"}` |

## Bentuk respons

- Berhasil: `{"success":true,"message":"...","data":{...}}` (daftar menambah `"meta"`)
- Gagal biasa: `{"success":false,"message":"..."}`
- Gagal validasi (422): `{"success":false,"message":"validasi gagal","errors":{"grade":"harus di antara 0 sampai 100"}}` — error menyebut field yang bermasalah.

## Contoh pengujian

```bash
# buat mahasiswa (201 + Location)
curl -i -X POST localhost:3000/api/v1/students -H "Content-Type: application/json" \
  -d '{"name":"Zagan Jade","nim":"230101","grade":92.5}'

# daftar dengan saring, urut, dan paginasi
curl "localhost:3000/api/v1/students?search=zagan&sort=grade&order=desc&grade_min=80&page=1&limit=10"

# PUT — harus kirim semua field
curl -X PUT localhost:3000/api/v1/students/1 -H "Content-Type: application/json" \
  -d '{"name":"Zagan Baru","nim":"230999","grade":95,"is_active":false}'

# PATCH — cukup field yang mau diubah
curl -X PATCH localhost:3000/api/v1/students/1 -H "Content-Type: application/json" \
  -d '{"grade":100}'

# hapus (204, tanpa body)
curl -i -X DELETE localhost:3000/api/v1/students/1
```

## Struktur berkas

- `main.go` — konfigurasi aplikasi, middleware, dan route
- `model.go` — struct entitas, request, dan respons
- `helper.go` — amplop respons dan pembacaan query string
- `handler.go` — fungsi penangan tiap endpoint
