package service

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
	"backend-golang/UTS/app/repository"
	"backend-golang/UTS/helper"
)

type EnrollmentService struct {
	pool     *pgxpool.Pool
	students repository.StudentRepository
	courses  repository.CourseRepository
	enroll   repository.EnrollmentRepository
}

func NewEnrollmentService(
	pool *pgxpool.Pool,
	students repository.StudentRepository,
	courses repository.CourseRepository,
	enroll repository.EnrollmentRepository,
) *EnrollmentService {
	return &EnrollmentService{pool: pool, students: students, courses: courses, enroll: enroll}
}

// POST /enrollments (mahasiswa): mengambil mata kuliah ke KRS.
//
// Seluruh pemeriksaan (duplikasi, kuota, batas SKS) berjalan dalam SATU
// transaction, dan baris course dikunci dengan FOR UPDATE supaya dua
// mahasiswa tidak bisa merebut kuota terakhir secara bersamaan.
func (s *EnrollmentService) Create(c fiber.Ctx) error {
	var req model.CreateEnrollmentRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.students.FindByUserID(ctx, current.UserID)
	if err != nil {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer tx.Rollback(ctx)

	// 1) kunci baris course; 404 bila tidak ada
	course, err := s.enroll.LockCourse(ctx, tx, req.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Validation(helper.FieldErr("course_id", "Mata kuliah tidak ditemukan"))
		}
		return helper.Internal(err)
	}

	// 2) tidak boleh mengambil mata kuliah yang sama pada tahun akademik sama
	exists, err := s.enroll.ExistsForStudent(ctx, tx, model.Enrollment{
		StudentID: student.ID, CourseID: course.ID, TahunAkademik: req.TahunAkademik,
	})
	if err != nil {
		return helper.Internal(err)
	}
	if exists {
		return helper.Conflict(fmt.Sprintf(
			"Mata kuliah %s sudah pernah diambil pada tahun akademik %s",
			course.KodeMK, req.TahunAkademik))
	}

	// 3) kuota penuh tidak dapat diambil
	if course.Terisi >= course.Kuota {
		return helper.Validation(helper.FieldErr("course_id",
			fmt.Sprintf("Kuota mata kuliah %s penuh", course.KodeMK)))
	}

	// 4) batas SKS sesuai IPK; pesan menyebut sisa SKS
	batas := BatasSKS(student.IPKTerakhir)
	terpakai, err := s.enroll.SumSKS(ctx, tx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}
	if terpakai+course.SKS > batas {
		return helper.Validation(helper.FieldErr("course_id",
			fmt.Sprintf("Total SKS melebihi batas: sisa SKS Anda %d, mata kuliah ini %d SKS",
				batas-terpakai, course.SKS)))
	}

	// 5) simpan
	hasil, err := s.enroll.CreateWithTx(ctx, tx, model.Enrollment{
		StudentID: student.ID, CourseID: course.ID, TahunAkademik: req.TahunAkademik,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("Mata kuliah sudah pernah diambil pada tahun akademik ini")
		}
		return helper.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "Mata kuliah berhasil diambil", fiber.Map{
		"id":             hasil.ID,
		"student_id":     hasil.StudentID,
		"course_id":      course.ID,
		"kode_mk":        course.KodeMK,
		"nama_mk":        course.NamaMK,
		"sks":            course.SKS,
		"tahun_akademik": hasil.TahunAkademik,
		"created_at":     hasil.CreatedAt,
	})
}

// DELETE /enrollments/:id (mahasiswa, milik sendiri).
func (s *EnrollmentService) Delete(c fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("ID harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	enrollment, err := s.enroll.FindByIDForOwner(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	student, err := s.students.FindByUserID(ctx, current.UserID)
	if err != nil || student.ID != enrollment.StudentID {
		return helper.Forbidden("Enrollment ini bukan milik Anda")
	}

	if err := s.enroll.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}
	return helper.NoContent(c)
}
