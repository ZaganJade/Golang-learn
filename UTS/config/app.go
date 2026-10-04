package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
	"backend-golang/UTS/helper"
	"backend-golang/UTS/middleware"
	"backend-golang/UTS/route"
)

func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "SIAKAD Mini - UTS",
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	route.Register(app, deps)

	app.Use(func(c fiber.Ctx) error {
		return helper.NotFound("Endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler satu-satunya tempat error menjadi response HTTP.
// Bentuk sesuai spesifikasi UTS: {success, message, errors?}.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):

		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status: fiberErr.Code, Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			attrs := []any{
				slog.String("path", c.Path()),
				slog.Int("status", appErr.Status),
			}
			if cause := appErr.Cause(); cause != nil {
				attrs = append(attrs, slog.String("error", cause.Error()))
			}
			logger.Error("request_failed", attrs...)
		}

		return c.Status(appErr.Status).JSON(model.WebResponse{
			Success: false,
			Message: appErr.Message,
			Errors:  appErr.Errors,
		})
	}
}
