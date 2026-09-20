# API Students — Tugas 6 Praktikum Pemrograman Backend Lanjut

API mahasiswa dengan autentikasi (JWT) dari tugas 5, sekarang ditambah
authorization berbasis RBAC: setiap endpoint students dijaga permission,
dan data mahasiswa punya pemilik (`owner_id`).

## Struktur proyek

```
Tugas6/
├── app/
│   ├── model/          struct entitas Student, User, auth, dan request-response
│   ├── repository/     query tabel students, users, refresh_tokens, roles
│   └── service/        business rules murni + penerima fiber.Ctx + aturan akses
├── config/             app.go, env.go, logger.go
├── database/           connection pool PostgreSQL
├── helper/             response envelope, pembaca query string, security, jwt, auth
├── logs/               output log (tidak di-commit)
├── middleware/         RequireJSON, RequireAuth, RequirePermission, rate limiter
├── route/              pendaftaran route sekaligus peta hak akses
├── migrations/         001 students, 002 auth, 003 rbac, 004 student permissions
├── .env                konfigurasi lokal (tidak di-commit)
└── main.go             urutan perakitan + graceful shutdown
```

## Role dan Permission (RBAC)

Hak akses disimpan di database (tabel `roles`, `permissions`,
`role_permissions`) dan dimuat sekali saat aplikasi menyala ke dalam
`PermissionSet`. Pemeriksaannya fail closed: role atau permission yang
tidak dikenal selalu ditolak.

| Permission | Arti | admin | staff | user |
|---|---|---|---|---|
| `student:list` | lihat daftar seluruh mahasiswa | ✓ | ✓ | — |
| `student:read:any` | lihat data mahasiswa mana pun | ✓ | ✓ | — |
| `student:create` | tambah mahasiswa | ✓ | ✓ | — |
| `student:update:any` | ubah mahasiswa mana pun | ✓ | — | — |
| `student:delete` | hapus mahasiswa | ✓ | — | — |

Dua pemeriksaan akses berjalan berbeda tempat:

- **Middleware** (`RequirePermission`) untuk keputusan yang tidak
  perlu lihat data: list, create, delete.
- **Service** (`CanAccessStudent` di `student_auth_rules.go`) untuk
  keputusan yang bergantung pada kepemilikan: read/put/patch per id.
  Pemilik data (`owner_id` = id pemanggil) selalu boleh; selain itu
  wajib punya permission `:any`.

`owner_id` terisi otomatis dari token saat `POST /students` dan tidak
bisa diubah lewat body request.

## Logger

Setiap request tercatat satu baris JSON di `logs/app.log` berisi
`request_id`, `method`, `path`, `status`, `duration`, `ip`, dan — untuk
request yang sudah login — `user_id` serta `role`. Jadi setiap penolakan
403 kelihatan siapa yang ditolak dan role-nya apa.

## Menyiapkan basis data dari nol

```bash
psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE praktikum_backend;"
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/001_create_students.sql
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/002_auth.sql
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/003_rbac.sql
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/004_student_permissions.sql
cp .env.example .env   # lalu isi nilainya, JWT_SECRET wajib diisi
go mod tidy
go run .
```

Urutan migration penting: `002` butuh tabel users, `003` menambahkan
kolom `role` sekaligus menguncinya dengan foreign key ke tabel roles,
dan `004` menambahkan permission students beserta kolom `owner_id`.

## Variabel environment

| Variabel | Arti | Contoh |
|---|---|---|
| `APP_PORT` | Port server | `3000` |
| `APP_NAME` | Nama aplikasi di header | `API Students` |
| `LOG_LEVEL` | Level log (debug/info/warn/error) | `info` |
| `JWT_SECRET` | Kunci tanda tangan token, minimal 32 karakter | |
| `JWT_ISSUER` | Penerbit token | `praktikum-backend` |
| `JWT_ACCESS_TTL_MINUTES` | Umur access token | `15` |
| `JWT_REFRESH_TTL_DAYS` | Umur refresh token | `7` |
| `ALLOWED_ORIGINS` | Origin yang boleh memanggil API | `http://localhost:5173` |
| `DB_HOST` | Alamat PostgreSQL | `127.0.0.1` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User basis data | `postgres` |
| `DB_PASSWORD` | Kata sandi (kosong kalau trust) | |
| `DB_NAME` | Nama database | `praktikum_backend` |
| `DB_SSLMODE` | Mode SSL | `disable` |
| `DB_MAX_CONNS` | Koneksi maksimum di pool | `10` |

## Kontrak API

Endpoint `/auth/*` publik (kecuali `/me`), seluruh endpoint `/students/*`
wajib membawa access token di header `Authorization: Bearer <token>`.

| Metode | Endpoint | Permission | Status yang mungkin |
|---|---|---|---|
| GET | `/api/v1/health` | — (publik) | `200`, `503` |
| POST | `/api/v1/auth/register` | — | `201`, `400`, `409`, `415`, `422`, `500` |
| POST | `/api/v1/auth/login` | — (rate limit 5x/menit) | `200`, `400`, `401`, `415`, `422`, `429`, `500` |
| POST | `/api/v1/auth/refresh` | — | `200`, `400`, `401`, `415`, `500` |
| POST | `/api/v1/auth/logout` | — | `200`, `400`, `415` |
| GET | `/api/v1/auth/me` | login | `200`, `401`, `404`, `500` |
| GET | `/api/v1/students` | `student:list` | `200`, `401`, `403`, `500` |
| GET | `/api/v1/students/:id` | pemilik atau `student:read:any` | `200`, `400`, `401`, `403`, `404`, `500` |
| POST | `/api/v1/students` | `student:create` | `201` (+`Location`), `400`, `401`, `403`, `409`, `415`, `422`, `500` |
| PUT | `/api/v1/students/:id` | pemilik atau `student:update:any` | `200`, `400`, `401`, `403`, `404`, `409`, `415`, `422`, `500` |
| PATCH | `/api/v1/students/:id` | pemilik atau `student:update:any` | `200`, `400`, `401`, `403`, `404`, `409`, `415`, `422`, `500` |
| DELETE | `/api/v1/students/:id` | `student:delete` | `204`, `400`, `401`, `403`, `404`, `500` |

Perbedaan 401 dan 403: 401 berarti identitas belum diketahui (belum
login, token kadaluarsa, atau tidak sah), 403 berarti identitas sudah
dikenal tetapi role-nya tidak punya permission. Salah satu pengecualian:
`PATCH /students/:id/role` atas diri sendiri tetap ditolak 422 karena
melanggar aturan bisnis, bukan karena hak.

Perubahan role tidak berlaku seketika pada token lama (JWT bersifat
stateless) — access token baru mengenali role terbaru setelah login
ulang atau refresh. Masa berlaku 15 menit menjadi batas keterlambatan
terbesarnya.

## Matriks hak akses

| Request | admin | staff | user (pemilik) | user (bukan pemilik) |
|---|---|---|---|---|
| GET `/students` | 200 | 200 | 403 | 403 |
| POST `/students` | 201 | 201 | 403 | 403 |
| GET `/students/:id` milik sendiri | 200 | 200 | 200 | — |
| GET `/students/:id` milik orang | 200 | 200 | — | 403 |
| PUT `/students/:id` milik sendiri | 200 | 200 | 200 | — |
| PUT `/students/:id` milik orang | 200 | 403 | — | 403 |
| PATCH `/students/:id` milik sendiri | 200 | 200 | 200 | — |
| PATCH `/students/:id` milik orang | 200 | 403 | — | 403 |
| DELETE `/students/:id` | 204 | 403 | 403 | 403 |
| Tanpa `Authorization` | 401 | 401 | 401 | 401 |
