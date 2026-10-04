package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
)

type EnrollmentRepository interface {
	// FindByStudent: daftar mata kuliah yang diambil mahasiswa (untuk detail KRS).
	FindByStudent(ctx context.Context, studentID int) ([]model.EnrollmentDetail, error)
	// FindByIDForOwner: satu enrollment + student_id pemiliknya (cek akses).
	FindByIDForOwner(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int) error

	// ---------- di dalam transaction ----------

	// LockCourse mengunci baris course (FOR UPDATE) dan mengembalikan
	// kuota + jumlah terisi. Row locking mencegah dua mahasiswa mengambil
	// kuota terakhir secara bersamaan.
	LockCourse(ctx context.Context, tx pgx.Tx, courseID int) (model.Course, error)
	// ExistsForStudent: cek duplikasi (student, course, tahun_akademik).
	ExistsForStudent(ctx context.Context, tx pgx.Tx, e model.Enrollment) (bool, error)
	// SumSKS: total SKS mahasiswa pada satu tahun akademik.
	SumSKS(ctx context.Context, tx pgx.Tx, studentID int, tahun string) (int, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error)
}

type enrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepository{pool: pool}
}

func (r *enrollmentRepository) FindByStudent(
	ctx context.Context, studentID int,
) ([]model.EnrollmentDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.tahun_akademik, c.kode_mk, c.nama_mk, c.sks, c.semester, e.created_at
		  FROM enrollments e
		  JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.created_at DESC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("mengambil KRS mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.EnrollmentDetail{}
	for rows.Next() {
		var d model.EnrollmentDetail
		if err := rows.Scan(&d.ID, &d.TahunAkademik, &d.KodeMK,
			&d.NamaMK, &d.SKS, &d.Semester, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris KRS: %w", err)
		}
		hasil = append(hasil, d)
	}
	return hasil, nil
}

func (r *enrollmentRepository) FindByIDForOwner(
	ctx context.Context, id int,
) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, course_id, tahun_akademik, created_at
		  FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *enrollmentRepository) LockCourse(
	ctx context.Context, tx pgx.Tx, courseID int,
) (model.Course, error) {
	var c model.Course
	err := tx.QueryRow(ctx, `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)
		  FROM courses c WHERE c.id = $1 FOR UPDATE`, courseID,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengunci course: %w", err)
	}
	c.SisaKuota = c.Kuota - c.Terisi
	return c, nil
}

func (r *enrollmentRepository) ExistsForStudent(
	ctx context.Context, tx pgx.Tx, e model.Enrollment,
) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM enrollments
		               WHERE student_id = $1 AND course_id = $2
		                 AND tahun_akademik = $3)`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&exists)
	return exists, err
}

func (r *enrollmentRepository) SumSKS(
	ctx context.Context, tx pgx.Tx, studentID int, tahun string,
) (int, error) {
	var total int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.sks), 0)
		  FROM enrollments e JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahun,
	).Scan(&total)
	return total, err
}

func (r *enrollmentRepository) CreateWithTx(
	ctx context.Context, tx pgx.Tx, e model.Enrollment,
) (model.Enrollment, error) {
	err := tx.QueryRow(ctx, `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3) RETURNING id, created_at`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("menyimpan enrollment: %w", err)
	}
	return e, nil
}
