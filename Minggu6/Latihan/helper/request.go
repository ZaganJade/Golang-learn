package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"backend-golang/Minggu6/Latihan/app/model"
)

// RequestContext memberi timeout untuk setiap operasi database.
func RequestContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur dan memastikan bentuknya benar.
func ParamID(c fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// queryInt membaca query string bertipe angka dengan nilai bawaan.
func queryInt(c fiber.Ctx, key string, def int) int {
	if v, err := strconv.Atoi(c.Query(key)); err == nil {
		return v
	}
	return def
}

var allowedSort = map[string]bool{
	"id": true, "username": true, "email": true, "created_at": true,
}

// ParseListQuery membaca query string dan memberi nilai bawaan yang aman.
func ParseListQuery(c fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   queryInt(c, "page", 1),
		Limit:  queryInt(c, "limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
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
