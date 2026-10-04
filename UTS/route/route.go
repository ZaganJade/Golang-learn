package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/service"
	"backend-golang/UTS/helper"
	"backend-golang/UTS/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

// Register memetakan 10 endpoint UTS.
func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// 1. login (publik, rate limit 5x/menit)
	api.Post("/auth/login", middleware.LoginRateLimiter(), deps.AuthService.Login)

	// 2. profil pengguna yang sedang login
	api.Get("/auth/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// 3-7. students: daftar/tambah admin; detail admin atau mahasiswa sendiri
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.List)
	students.Post("/", middleware.RequireRole("admin"), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", middleware.RequireRole("admin"), deps.StudentService.Update)
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.Delete)

	// 8. courses: semua role yang login
	api.Get("/courses", middleware.RequireAuth(deps.JWT), deps.CourseService.List)

	// 9-10. enrollments: hanya mahasiswa
	enrollments := api.Group("/enrollments",
		middleware.RequireAuth(deps.JWT), middleware.RequireRole("mahasiswa"))
	enrollments.Post("/", deps.EnrollmentService.Create)
	enrollments.Delete("/:id", deps.EnrollmentService.Delete)

	api.Get("/health", healthCheck(deps.Pool))
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Internal(err)
		}
		return helper.Success(c, fiber.StatusOK, "Server dan database berjalan", nil)
	}
}
