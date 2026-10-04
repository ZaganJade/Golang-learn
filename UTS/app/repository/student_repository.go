package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	ExistsByNIM(ctx context.Context, nim string, ignoreID int) (bool, error)
	ExistsByEmail(ctx context.Context, email string, ignoreID int) (bool, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentRepository{pool: pool}
}

const studentColumns = `s.id, s.user_id, u.email, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir`

func scanStudent(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.Email, &s.NIM, &s.Nama,
		&s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	return s, err
}

func (r *studentRepository) buildFilter(q model.StudentQuery) (string, []any) {
	where := " WHERE s.deleted_at IS NULL"
	args := []any{}

	if q.Prodi != "" {
		args = append(args, q.Prodi)
		where += fmt.Sprintf(" AND s.prodi = $%d", len(args))
	}
	if q.Angkatan != nil {
		args = append(args, *q.Angkatan)
		where += fmt.Sprintf(" AND s.angkatan = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (s.nim ILIKE $%d OR s.nama ILIKE $%d)",
			len(args)-1, len(args))
	}
	return where, args
}

// urutan kolom pakai daftar putih karena ORDER BY tidak bisa memakai parameter.
var kolomSort = map[string]string{
	"nama":           "s.nama ASC, s.id ASC",
	"-ipk_terakhir":  "s.ipk_terakhir DESC, s.id ASC",
}

func (r *studentRepository) FindAll(
	ctx context.Context, q model.StudentQuery,
) ([]model.Student, int, error) {
	where, args := r.buildFilter(q)

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM students s"+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	sort, ok := kolomSort[q.Sort]
	if !ok {
		sort = kolomSort["nama"]
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM students s JOIN users u ON u.id = s.user_id%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		studentColumns, where, sort, len(args)+1, len(args)+2,
	)
	args = append(args, q.PerPage, (q.Page-1)*q.PerPage)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.Email, &s.NIM, &s.Nama,
			&s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, s)
	}
	return hasil, total, nil
}

func (r *studentRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	s, err := scanStudent(r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM students s JOIN users u ON u.id = s.user_id
		 WHERE s.id = $1 AND s.deleted_at IS NULL`, studentColumns),
		id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return s, nil
}

func (r *studentRepository) FindByUserID(
	ctx context.Context, userID int,
) (model.Student, error) {
	s, err := scanStudent(r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM students s JOIN users u ON u.id = s.user_id
		 WHERE s.user_id = $1 AND s.deleted_at IS NULL`, studentColumns),
		userID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student by user: %w", err)
	}
	return s, nil
}

func (r *studentRepository) ExistsByNIM(
	ctx context.Context, nim string, ignoreID int,
) (bool, error) {
	// ignoreID = id student yang sedang diubah (0 saat create).
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM students
		               WHERE nim = $1 AND deleted_at IS NULL AND ($2 = 0 OR id <> $2))`,
		nim, ignoreID,
	).Scan(&exists)
	return exists, err
}

func (r *studentRepository) ExistsByEmail(
	ctx context.Context, email string, ignoreID int,
) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users u
		                 JOIN students s ON s.user_id = u.id
		                WHERE LOWER(u.email) = LOWER($1)
		                  AND s.deleted_at IS NULL
		                  AND ($2 = 0 OR s.id <> $2))`,
		email, ignoreID,
	).Scan(&exists)
	return exists, err
}

func (r *studentRepository) CreateWithTx(
	ctx context.Context, tx pgx.Tx, s model.Student,
) (model.Student, error) {
	err := tx.QueryRow(ctx, `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	).Scan(&s.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}
	return s, nil
}

func (r *studentRepository) Update(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE students SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir`,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}
	return s, nil
}

func (r *studentRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW()
		  WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
