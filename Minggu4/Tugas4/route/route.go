package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/Minggu4/Tugas4/app/service"
	"backend-golang/Minggu4/Tugas4/helper"
	"backend-golang/Minggu4/Tugas4/middleware"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Get("/nim/:nim/nilai", studentService.GetNilai)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
