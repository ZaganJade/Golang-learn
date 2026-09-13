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

INSERT INTO students (nim, name, grade) VALUES
    ('230101', 'Amba', 90),
    ('230102', 'Rusdi', 85),
    ('230103', 'Mas Fuad', 88),
    ('230104', 'Azril', 92),
    ('230105', 'Mas Gatot', 78)
ON CONFLICT (nim) DO NOTHING;
