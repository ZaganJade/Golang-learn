package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"backend-golang/Minggu4/Tugas4/app/model"
)

func RequestContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 5*time.Second)
}

func ParamID(c fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
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

func ParseListQuery(c fiber.Ctx) model.ListQuery {
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
