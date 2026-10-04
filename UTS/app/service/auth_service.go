package service

import (
	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
	"backend-golang/UTS/app/repository"
	"backend-golang/UTS/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	students repository.StudentRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{users: users, students: students, jwt: jwtManager}
}

func (s *AuthService) Login(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, deletedAt, err := s.users.FindForLogin(ctx, req.Email)
	if err != nil {
		// verifikasi dummy agar waktu respons tidak membedakan email
		// yang terdaftar dari yang tidak
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("Email atau password salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("Email atau password salah")
	}

	// mahasiswa yang di-soft delete tidak boleh login
	if user.Role == "mahasiswa" && deletedAt != nil {
		return helper.Unauthorized("Email atau password salah")
	}

	token, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.jwt.AccessTTL().Seconds()),
		"user": fiber.Map{
			"id": user.ID, "email": user.Email, "role": user.Role,
		},
	})
}

func (s *AuthService) Me(c fiber.Ctx) error {
	auth, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByID(ctx, auth.UserID)
	if err != nil {
		return helper.Unauthorized("Token tidak valid")
	}

	data := fiber.Map{"id": user.ID, "email": user.Email, "role": user.Role}

	// untuk mahasiswa, sertakan data akademiknya
	if user.Role == "mahasiswa" {
		if student, err := s.students.FindByUserID(ctx, user.ID); err == nil {
			data["student"] = fiber.Map{
				"id": student.ID, "nim": student.NIM, "nama": student.Nama,
				"prodi": student.Prodi, "angkatan": student.Angkatan,
				"ipk_terakhir": student.IPKTerakhir,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil berhasil diambil", data)
}
