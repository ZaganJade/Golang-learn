CREATE TABLE IF NOT EXISTS Nilai (
    id_nilai         SERIAL       PRIMARY KEY,
    id_student       INT          NOT NULL,
    nameMK       VARCHAR(100) NOT NULL,
    grade      NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    foreign key (id_student) references students(id)
);

insert into Nilai (id_student, nameMK, grade)
select * from (values
    (1, 'Matematika', 90),
    (1, 'Fisika', 85),
    (2, 'Kimia', 88),
    (2, 'Biologi', 92),
    (3, 'Sejarah', 78)
) as seed(id_student, nameMK, grade)
where not exists (select 1 from Nilai)