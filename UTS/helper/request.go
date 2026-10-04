package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
)

// RequestContext membatasi waktu query database per request.
func RequestContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 5*time.Second)
}

// ParamID membaca :id yang harus berupa angka positif.
func ParamID(c fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ParseStudentQuery membaca page, per_page, prodi, angkatan, search, sort.
func ParseStudentQuery(c fiber.Ctx) model.StudentQuery {
	q := model.StudentQuery{
		Page:    atoiDefault(c.Query("page"), 1),
		PerPage: atoiDefault(c.Query("per_page"), 10),
		Prodi:   strings.TrimSpace(c.Query("prodi")),
		Search:  strings.TrimSpace(c.Query("search")),
		Sort:    strings.TrimSpace(c.Query("sort")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 10
	}
	if q.PerPage > 50 {
		q.PerPage = 50
	}
	if raw := c.Query("angkatan"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			q.Angkatan = &v
		}
	}
	return q
}

// ParseCourseQuery membaca semester, search, available.
func ParseCourseQuery(c fiber.Ctx) model.CourseQuery {
	q := model.CourseQuery{
		Search: strings.TrimSpace(c.Query("search")),
	}
	if raw := c.Query("semester"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 1 && v <= 8 {
			q.Semester = &v
		}
	}
	// available=true: hanya yang kuotanya belum penuh
	if v, err := strconv.ParseBool(c.Query("available")); err == nil {
		q.Available = v
	}
	return q
}

func atoiDefault(raw string, def int) int {
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return def
}
