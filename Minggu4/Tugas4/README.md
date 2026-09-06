# API Students — Tugas 4 Praktikum Pemrograman Backend Lanjut

API mahasiswa dari tugas 3, direstrukturisasi ke struktur baku Clean Architecture mata kuliah. Perilaku HTTP-nya tidak berubah; yang berubah hanya letak kodenya.

## Struktur proyek

```
Tugas4/
├── app/
│   ├── model/          struct entitas Student dan request-response
│   ├── repository/     query ke tabel students + sentinel error
│   └── service/        business rules murni + penerima fiber.Ctx
├── config/             app.go, env.go, logger.go
├── database/           connection pool PostgreSQL
├── helper/             response envelope dan pembaca query string
├── logs/               output log (tidak di-commit)
├── middleware/         middleware global dan RequireJSON
├── route/              pendaftaran route
├── migrations/         skema tabel + data awal
├── .env                konfigurasi lokal (tidak di-commit)
└── main.go             urutan perakitan + graceful shutdown
```

Pemetaan ke empat layer Clean Architecture:

| Layer | Folder |
|---|---|
| 1. Entities | `app/model/` |
| 2. Use Cases | `app/service/student_rules.go` (business rules murni) |
| 3. Interface Adapters | `app/repository/` (gateway), `helper/` (presenter), `app/service/student_service.go` (controller) |
| 4. Frameworks & Drivers | `config/`, `database/`, `middleware/`, `route/`, `main.go` |

Business rules dipisah sebagai fungsi biasa (`ValidateCreate`, `ValidateReplace`, `ApplyPatch`) yang tidak menyentuh fiber.Ctx sama sekali, sehingga bisa diuji dengan `go test` tanpa menyalakan server maupun database: `go test ./app/service/...`.

## Logger

Setiap request tercatat satu baris JSON di `logs/app.log` (sekalian ke layar), berisi `request_id`, `method`, `path`, `status`, `duration`, dan `ip`. File dirotasi otomatis oleh lumberjack (10 MB, simpan 5 file lama). Folder `logs/` ada di `.gitignore`.

## Menyiapkan basis data dari nol

```bash
psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE praktikum_backend;"
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/001_create_students.sql
cp .env.example .env   # lalu isi nilainya
go mod tidy
go run .
```

## Variabel environment

| Variabel | Arti | Contoh |
|---|---|---|
| `APP_PORT` | Port server | `3000` |
| `APP_NAME` | Nama aplikasi di header | `API Students` |
| `LOG_LEVEL` | Level log (debug/info/warn/error) | `info` |
| `DB_HOST` | Alamat PostgreSQL | `127.0.0.1` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User basis data | `postgres` |
| `DB_PASSWORD` | Kata sandi (kosong kalau trust) | |
| `DB_NAME` | Nama database | `praktikum_backend` |
| `DB_SSLMODE` | Mode SSL | `disable` |
| `DB_MAX_CONNS` | Koneksi maksimum di pool | `10` |

## Kontrak API

Sama persis dengan tugas 3 — tidak ada endpoint yang berubah:

| Metode | Endpoint | Status yang mungkin |
|---|---|---|
| GET | `/api/v1/health` | `200`, `503` |
| GET | `/api/v1/students` (`page`, `limit` maks 50, `search`, `sort`, `order`, `is_active`, `grade_min`, `grade_max`) | `200`, `500` |
| GET | `/api/v1/students/:id` | `200`, `400`, `404`, `500` |
| POST | `/api/v1/students` | `201` (+`Location`), `400`, `409`, `415`, `422`, `500` |
| PUT | `/api/v1/students/:id` | `200`, `400`, `404`, `409`, `415`, `422`, `500` |
| PATCH | `/api/v1/students/:id` | `200`, `400`, `404`, `409`, `415`, `422`, `500` |
| DELETE | `/api/v1/students/:id` | `204`, `400`, `404`, `500` |

Bentuk respons amplop (`success`/`message`/`data`/`meta`/`errors`) juga sama seperti tugas 3.
