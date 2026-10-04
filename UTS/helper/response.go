package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
)

// AppError satu-satunya bentuk kegagalan; response ditulis ErrorHandler terpusat.
type AppError struct {
	Status  int                 // status HTTP
	Message string              // pesan untuk manusia
	Errors  map[string][]string // detail per-field (khusus 422)
	cause   error               // error asli, hanya untuk log
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%d: %s: %v", e.Status, e.Message, e.cause)
	}
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

// Cause membuka akses baca untuk log; tidak pernah masuk response.
func (e *AppError) Cause() error { return e.cause }

func Unauthorized(message string) *AppError {
	return &AppError{Status: fiber.StatusUnauthorized, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: fiber.StatusForbidden, Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Status: fiber.StatusNotFound, Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Status: fiber.StatusConflict, Message: message}
}

func BadRequest(message string) *AppError {
	return &AppError{Status: fiber.StatusBadRequest, Message: message}
}

func TooManyRequests(message string) *AppError {
	return &AppError{Status: fiber.StatusTooManyRequests, Message: message}
}

// Validation: 422 dengan errors per-field sesuai format UTS.
func Validation(errors map[string][]string) *AppError {
	return &AppError{
		Status: fiber.StatusUnprocessableEntity,
		Message: "Validasi gagal",
		Errors: errors,
	}
}

// FieldErr menyusun satu entry errors: map[string][]string.
func FieldErr(field, message string) map[string][]string {
	return map[string][]string{field: {message}}
}

// Internal: pesan seragam, detail hanya ke log.
func Internal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Message: "Terjadi kesalahan pada server",
		cause:   cause,
	}
}

// ---------- response sukses ----------

func Success(c fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func SuccessList(c fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func Created(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
