-- SIAKAD Mini: schema + seeder awal
-- 1 admin, 20 mahasiswa (password = bcrypt(NIM)), 10 mata kuliah.

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL       PRIMARY KEY,
    email      VARCHAR(150) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa'
               CHECK (role IN ('admin', 'mahasiswa')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- email unik tanpa membedakan huruf besar/kecil
CREATE UNIQUE INDEX IF NOT EXISTS users_email_uniq_idx ON users (LOWER(email));

CREATE TABLE IF NOT EXISTS students (
    id           SERIAL       PRIMARY KEY,
    user_id      INTEGER      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim          VARCHAR(12)  NOT NULL UNIQUE,
    nama         VARCHAR(100) NOT NULL,
    prodi        VARCHAR(100) NOT NULL,
    angkatan     SMALLINT     NOT NULL,
    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (ipk_terakhir BETWEEN 0 AND 4),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS students_prodi_idx ON students (prodi);
CREATE INDEX IF NOT EXISTS students_angkatan_idx ON students (angkatan);
CREATE INDEX IF NOT EXISTS students_nama_idx ON students (LOWER(nama));

CREATE TABLE IF NOT EXISTS courses (
    id       SERIAL       PRIMARY KEY,
    kode_mk  VARCHAR(10)  NOT NULL UNIQUE,
    nama_mk  VARCHAR(100) NOT NULL,
    sks      SMALLINT     NOT NULL CHECK (sks BETWEEN 1 AND 6),
    semester SMALLINT     NOT NULL CHECK (semester BETWEEN 1 AND 8),
    kuota    SMALLINT     NOT NULL CHECK (kuota >= 0)
);

CREATE TABLE IF NOT EXISTS enrollments (
    id             SERIAL      PRIMARY KEY,
    student_id     INTEGER     NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id     INTEGER     NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS enrollments_course_idx ON enrollments (course_id);
CREATE INDEX IF NOT EXISTS enrollments_student_tahun_idx ON enrollments (student_id, tahun_akademik);

-- ===== seeder =====

INSERT INTO users (email, password, role) VALUES
    ('admin@siakad.ac.id', '$2a$10$7JTLKVC6SpKJBjPCykQ6.eZq0Qg4a8fszEBNScCsbcFQ.ZSjyKLPG', 'admin')
ON CONFLICT DO NOTHING;

INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
    ('IF101', 'Dasar Pemrograman',        4, 1, 30),
    ('IF102', 'Basis Data',               3, 2, 30),
    ('IF201', 'Struktur Data',            3, 3, 25),
    ('IF202', 'Pemrograman Web',          3, 3, 25),
    ('IF301', 'Jaringan Komputer',        3, 4, 20),
    ('IF302', 'Rekayasa Perangkat Lunak', 3, 4, 20),
    ('IF401', 'Sistem Terdistribusi',     3, 5, 15),
    ('SI301', 'Analisis dan Desain Sistem', 3, 3, 20),
    ('SI401', 'Sistem Informasi Manajemen', 3, 5, 18),
    ('SD201', 'Data Mining',              3, 4, 2)
ON CONFLICT (kode_mk) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000001@student.unair.ac.id', '$2a$10$p6Ot9sDPErlH.d9LxYyNReEP.A0zjEJMu1kLFLQKEKNsM4eo8H4je', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000001', 'Rina Putri', 'Sistem Informasi', 2022, 3.45
  FROM users WHERE email = '187221000001@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000002@student.unair.ac.id', '$2a$10$sFaco3.VX0dMxkIIELcaaedfOBd0KgLJZYr.uUY0HCsrofDO6lpq6', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000002', 'Budi Santoso', 'Teknik Informatika', 2023, 2.7
  FROM users WHERE email = '187221000002@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000003@student.unair.ac.id', '$2a$10$inQb84ieLrj0r4U13aBTGuvBzKRvSKCI.5IzZ9F4g0XAkwHmtH/PS', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000003', 'Sari Melati', 'Sains Data', 2024, 2.3
  FROM users WHERE email = '187221000003@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000004@student.unair.ac.id', '$2a$10$h1nu6bGjkxO/f9rYWKlxl.cBHP2oWz9JQM77EWilWw.7vBd2Zkaya', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000004', 'Andi Pratama', 'Sistem Informasi', 2025, 3.0
  FROM users WHERE email = '187221000004@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000005@student.unair.ac.id', '$2a$10$VxbHqukr7o5dCLo.gU1ZnOpLl0p0zgZkhYmQ8M5tj53JfflxmEcoy', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000005', 'Dewi Lestari', 'Teknik Informatika', 2022, 3.8
  FROM users WHERE email = '187221000005@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000006@student.unair.ac.id', '$2a$10$9Amf99WWjoD3xtdrakpddurrE6yAqivTAATdk5j5YA0ePLumtw2mi', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000006', 'Fajar Nugroho', 'Sains Data', 2023, 2.55
  FROM users WHERE email = '187221000006@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000007@student.unair.ac.id', '$2a$10$o7WPSFYfByjIWDEISEUf/.rDQu.S7zvFvWd8jRFR2ww05HhNVBpzu', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000007', 'Intan Permata', 'Sistem Informasi', 2024, 1.9
  FROM users WHERE email = '187221000007@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000008@student.unair.ac.id', '$2a$10$OaJGa.hpRweU9Y5JI0PWVegSjcnNy15XIuTaKxE5glytBnz3u2SsW', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000008', 'Rizky Ramadhan', 'Teknik Informatika', 2025, 2.99
  FROM users WHERE email = '187221000008@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000009@student.unair.ac.id', '$2a$10$wIRS8CIP5Ir0zsa1T08yLuAKnnCuF.oTrADLKQSQQivxeGbhC/zs6', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000009', 'Maya Anggraini', 'Sains Data', 2022, 3.1
  FROM users WHERE email = '187221000009@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000010@student.unair.ac.id', '$2a$10$0zdpYLOSpKfYHHicYhxnx.k3knhVxfnwMerGQvFeNLCLkk/gVl572', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000010', 'Bagas Saputra', 'Sistem Informasi', 2023, 2.4
  FROM users WHERE email = '187221000010@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000011@student.unair.ac.id', '$2a$10$oEswjjlCiv07jE/tF2QCtuXceWAi4Xb7WQQcEptYZds8dPLJwZml.', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000011', 'Nadia Safira', 'Teknik Informatika', 2024, 3.6
  FROM users WHERE email = '187221000011@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000012@student.unair.ac.id', '$2a$10$BalKrVOGJqI1T6QDj3KBTOgm8tuVchoTmPNv9LZ2UpBNiems9OCQG', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000012', 'Dimas Ardiansyah', 'Sains Data', 2025, 2.8
  FROM users WHERE email = '187221000012@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000013@student.unair.ac.id', '$2a$10$B8auu9qNPNRxZ./FxG8bA.WMsUzSgGHT/9A6ITU3ZmuOUp.SluQb2', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000013', 'Laras Ayuningtyas', 'Sistem Informasi', 2022, 2.2
  FROM users WHERE email = '187221000013@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000014@student.unair.ac.id', '$2a$10$sN3RWGdW6XWFUoK5Ug4JJecZgqhKqgH.Q8U4iy5xiDgc3K2KVmptu', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000014', 'Yoga Prasetyo', 'Teknik Informatika', 2023, 3.25
  FROM users WHERE email = '187221000014@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000015@student.unair.ac.id', '$2a$10$VB5UfBCNqYYIM9KNI8lChOtHXnZaMMLtjlfTaxkVnT6jKd8g63Nvu', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000015', 'Tiara Kusuma', 'Sains Data', 2024, 2.95
  FROM users WHERE email = '187221000015@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000016@student.unair.ac.id', '$2a$10$.KydtP1dZiBj238PsAxdAeaUZMZuh09sgajqCWF9AgTqM1kQob5Lu', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000016', 'Farhan Maulana', 'Sistem Informasi', 2025, 2.5
  FROM users WHERE email = '187221000016@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000017@student.unair.ac.id', '$2a$10$pA2G4IRDXWiEO9Q.qrUVVuVSpVX7kz7sWq3ITteV0lECHFtljGZ8O', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000017', 'Alya Zahra', 'Teknik Informatika', 2022, 3.75
  FROM users WHERE email = '187221000017@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000018@student.unair.ac.id', '$2a$10$CcbYAxZLuguo2iWBmipFJeA10VdRjnsIZtO5Sr2c0bmiI7WgH3rje', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000018', 'Gilang Ramadhan', 'Sains Data', 2023, 2.35
  FROM users WHERE email = '187221000018@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000019@student.unair.ac.id', '$2a$10$eR1cQO.23ZJFbFvcZy3xYOS0KpVV72OgFaaBsc8f8HSXuWphl5xwe', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000019', 'Nabila Husna', 'Sistem Informasi', 2024, 3.05
  FROM users WHERE email = '187221000019@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;

INSERT INTO users (email, password, role) VALUES
    ('187221000020@student.unair.ac.id', '$2a$10$LVa2MnwL4tG.9PFJDTGzJeo7uLzkL.vj.JETWCyjb57tCk61iwyaq', 'mahasiswa')
ON CONFLICT DO NOTHING;

INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT id, '187221000020', 'Reza Firmansyah', 'Teknik Informatika', 2025, 2.6
  FROM users WHERE email = '187221000020@student.unair.ac.id'
ON CONFLICT (nim) DO NOTHING;
