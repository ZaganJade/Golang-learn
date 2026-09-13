package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/Minggu5/Tugas5/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindNilaiByNIM(ctx context.Context, nim string) ([]model.StudentNilai, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

var kolomUrut = map[string]string{
	"id":         "id",
	"name":       "name",
	"nim":        "nim",
	"grade":      "grade",
	"created_at": "created_at",
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func buildFilter(q model.ListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}
	if q.GradeMin != nil {
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, *q.GradeMin)
	}
	if q.GradeMax != nil {
		where += fmt.Sprintf(" AND grade <= $%d", len(args)+1)
		args = append(args, *q.GradeMax)
	}

	return where, args
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at
		   FROM students%s
		  ORDER BY %s %s
		  LIMIT $%d OFFSET $%d`,
		where, kolomUrut[q.Sort], arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade,
			&s.IsActive, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, total, nil
}

func (r *studentPostgresRepository) FindNilaiByNIM(
	ctx context.Context, nim string,
) ([]model.StudentNilai, error) {
	var idStudent int
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM students WHERE nim = $1`, nim).Scan(&idStudent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("mencari mahasiswa dengan nim %s: %w", nim, err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id_nilai, nameMK, id_student, grade, is_active, created_at
		   FROM Nilai WHERE id_student = $1`, idStudent)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar nilai: %w", err)
	}
	defer rows.Close()

	hasil := []model.StudentNilai{}
	for rows.Next() {
		var s model.StudentNilai
		if err := rows.Scan(&s.ID_Nilai, &s.NameMK, &s.ID_Student,
			&s.Grade, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris nilai: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, nil
}

func (r *studentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	var s model.Student

	err := r.pool.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active, created_at
		   FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}

	return s, nil
}

func (r *studentPostgresRepository) Create(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}

	return s, nil
}

func (r *studentPostgresRepository) Update(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students SET nim = $1, name = $2, grade = $3, is_active = $4
		  WHERE id = $5
		  RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.ID,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}

	return s, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
