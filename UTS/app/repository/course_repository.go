package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, q model.CourseQuery) ([]model.Course, error)
}

type courseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &courseRepository{pool: pool}
}

// terisi/sisa_kuota dihitung langsung dari tabel enrollments lewat subquery,
// sehingga tidak perlu query tambahan per mata kuliah.
func (r *courseRepository) FindAll(
	ctx context.Context, q model.CourseQuery,
) ([]model.Course, error) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Semester != nil {
		args = append(args, *q.Semester)
		where += fmt.Sprintf(" AND c.semester = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)",
			len(args)-1, len(args))
	}
	if q.Available {
		where += " AND c.kuota > (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)"
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) AS terisi
		  FROM courses c%s
		 ORDER BY c.semester, c.kode_mk`, where),
		args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar course: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS,
			&c.Semester, &c.Kuota, &c.Terisi); err != nil {
			return nil, fmt.Errorf("membaca baris course: %w", err)
		}
		c.SisaKuota = c.Kuota - c.Terisi
		hasil = append(hasil, c)
	}
	return hasil, nil
}
