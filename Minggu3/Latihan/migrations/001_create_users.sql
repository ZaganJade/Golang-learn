CREATE TABLE IF NOT EXISTS users (
    id         SERIAL      PRIMARY KEY,
    username   VARCHAR(50) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Keunikan username tanpa membedakan huruf besar dan kecil.
-- Inilah yang menggantikan pemeriksaan manual di pertemuan 2.
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key
    ON users (LOWER(username));

CREATE INDEX IF NOT EXISTS users_email_lower_idx
    ON users (LOWER(email));

INSERT INTO users (username, email, password) VALUES
    ('bu tejo', 'butejo@example.com', 'rahasia123'),
    ('bang toyib', 'bangtoyib@example.com', 'rahasia123'),
    ('mbak mumun', 'mbakmumun@example.com', 'rahasia123'),
    ('mas gugum', 'masgugum@example.com', 'rahasia123'),
    ('pak kumis', 'pakkumis@example.com', 'rahasia123')
ON CONFLICT DO NOTHING;
