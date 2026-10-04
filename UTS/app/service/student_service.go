package service

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
	"backend-golang/UTS/app/repository"
	"backend-golang/UTS/helper"
)

type StudentService struct {
	pool        *pgxpool.Pool
	repo        repository.StudentRepository
	users       repository.UserRepository
	enrollments repository.EnrollmentRepository
}

func NewStudentService(
	pool *pgxpool.Pool,
	repo repository.StudentRepository,
	users repository.UserRepository,
	enrollments repository.EnrollmentRepository,
) *StudentService {
	return &StudentService{pool: pool, repo: repo, users: users, enrollments: enrollments}
}

// ---------- GET /students (admin) ----------

func (s *StudentService) List(c fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseStudentQuery(c)
	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	lastPage := 1
	if q.PerPage > 0 {
		lastPage = (total + q.PerPage - 1) / q.PerPage
		if lastPage < 1 {
			lastPage = 1
		}
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", students, &model.Meta{
		CurrentPage: q.Page,
		PerPage:     q.PerPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

// ---------- POST /students (admin) ----------

func (s *StudentService) Create(c fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// duplikasi diperiksa eksplisit agar 422-nya menyebut field yang tepat
	errs := map[string][]string{}
	if exists, err := s.repo.ExistsByNIM(ctx, req.NIM, 0); err != nil {
		return helper.Internal(err)
	} else if exists {
		errs["nim"] = append(errs["nim"], "NIM sudah terdaftar")
	}
	if exists, err := s.repo.ExistsByEmail(ctx, req.Email, 0); err != nil {
		return helper.Internal(err)
	} else if exists {
		errs["email"] = append(errs["email"], "Email sudah terdaftar")
	}
	if len(errs) > 0 {
		return helper.Validation(errs)
	}

	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	// password awal = NIM (di-hash) sesuai ketentuan
	hashed, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Internal(err)
	}

	// users + students dalam satu transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer tx.Rollback(ctx)

	userID, err := s.users.CreateWithTx(ctx, tx, model.User{
		Email: req.Email, Password: hashed, Role: "mahasiswa",
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(helper.FieldErr("email", "Email sudah terdaftar"))
		}
		return helper.Internal(err)
	}

	student, err := s.repo.CreateWithTx(ctx, tx, model.Student{
		UserID: userID, NIM: req.NIM, Nama: req.Nama,
		Prodi: req.Prodi, Angkatan: req.Angkatan, IPKTerakhir: ipk,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(helper.FieldErr("nim", "NIM sudah terdaftar"))
		}
		return helper.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	student.Email = req.Email
	return helper.Created(c, "Mahasiswa berhasil ditambahkan", student)
}

// ---------- GET /students/:id (admin / mahasiswa sendiri) ----------

func (s *StudentService) Get(c fiber.Ctx) error {
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

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "Mahasiswa")
	}

	// mahasiswa hanya boleh melihat data sendiri
	if current.Role != "admin" {
		if current.UserID != student.UserID {
			return helper.Forbidden("Tidak berhak mengakses data mahasiswa lain")
		}
	}

	detail, err := s.enrollments.FindByStudent(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}

	totalSKS := 0
	for _, d := range detail {
		totalSKS += d.SKS
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", fiber.Map{
		"id":           student.ID,
		"nim":          student.NIM,
		"nama":         student.Nama,
		"prodi":        student.Prodi,
		"angkatan":     student.Angkatan,
		"ipk_terakhir": student.IPKTerakhir,
		"email":        student.Email,
		"mata_kuliah":  detail,
		"total_sks":    totalSKS,
		"batas_sks":    BatasSKS(student.IPKTerakhir),
	})
}

// ---------- PUT /students/:id (admin) ----------

func (s *StudentService) Update(c fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("ID harus berupa angka positif")
	}

	var req model.UpdateStudentRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "Mahasiswa")
	}

	ipk := saatIni.IPKTerakhir
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID: id, Nama: req.Nama, Prodi: req.Prodi,
		Angkatan: req.Angkatan, IPKTerakhir: ipk,
	})
	if err != nil {
		return translateError(err, "Mahasiswa")
	}
	hasil.Email = saatIni.Email

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", hasil)
}

// ---------- DELETE /students/:id (admin, soft delete) ----------

func (s *StudentService) Delete(c fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("ID harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return translateError(err, "Mahasiswa")
	}
	return helper.NoContent(c)
}

func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict(entity + " sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}
