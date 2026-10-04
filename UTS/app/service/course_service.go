package service

import (
	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
	"backend-golang/UTS/app/repository"
	"backend-golang/UTS/helper"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

// GET /courses: semua role yang sudah login.
func (s *CourseService) List(c fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseCourseQuery(c)
	courses, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.SuccessList(c, "Data mata kuliah berhasil diambil", courses, &model.Meta{
		CurrentPage: 1, PerPage: len(courses), Total: len(courses), LastPage: 1,
	})
}

// BatasSKS sesuai business rule berdasarkan IPK terakhir.
func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}
