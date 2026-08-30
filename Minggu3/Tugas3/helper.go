package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"backend-golang/Minggu3/Tugas3/app/model"
)

func ok(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func okList(c fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func created(c fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func noContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{Success: false, Message: message})
}

func failValidation(c fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false, Message: "validasi gagal", Errors: errs,
	})
}

func reqCtx(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 5*time.Second)
}

func queryInt(c fiber.Ctx, key string, def int) int {
	if v, err := strconv.Atoi(c.Query(key)); err == nil {
		return v
	}
	return def
}

func queryFloat(c fiber.Ctx, key string) *float64 {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	if v, err := strconv.ParseFloat(raw, 64); err == nil {
		return &v
	}
	return nil
}

var allowedSort = map[string]bool{
	"id": true, "name": true, "nim": true, "grade": true, "created_at": true,
}

func parseListQuery(c fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:     queryInt(c, "page", 1),
		Limit:    queryInt(c, "limit", 10),
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     c.Query("sort", "id"),
		Order:    strings.ToLower(c.Query("order", "asc")),
		GradeMin: queryFloat(c, "grade_min"),
		GradeMax: queryFloat(c, "grade_max"),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	return q
}
