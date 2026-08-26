package main

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

var students []Student
var nextID = 1

func seedData() {
	awal := []struct {
		nama  string
		nim   string
		grade float64
	}{
		{"Amba", "230101", 90},
		{"Rusdi", "230102", 85},
		{"Mas Fuad", "230103", 88},
		{"Azril", "230104", 92},
		{"Mas Gatot", "230105", 78},
	}
	for _, s := range awal {
		students = append(students, Student{
			ID:        nextID,
			Name:      s.nama,
			NIM:       s.nim,
			Grade:     s.grade,
			IsActive:  true,
			CreatedAt: time.Now(),
		})
		nextID++
	}
}

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func nimDipakai(nim string, excludeID int) bool {
	for _, s := range students {
		if s.ID != excludeID && strings.EqualFold(s.NIM, nim) {
			return true
		}
	}
	return false
}

func cocokPencarian(s Student, kata string) bool {
	return strings.Contains(strings.ToLower(s.Name), strings.ToLower(kata))
}

func paramID(c fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func validasiNama(nama string) bool {
	return strings.TrimSpace(nama) != ""
}

func validasiNIM(nim string) bool {
	return strings.TrimSpace(nim) != ""
}

func validasiGrade(g float64) bool {
	return g >= 0 && g <= 100
}

func listStudents(c fiber.Ctx) error {
	q := parseListQuery(c)

	hasil := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.GradeMin != nil && s.Grade < *q.GradeMin {
			continue
		}
		if q.GradeMax != nil && s.Grade > *q.GradeMax {
			continue
		}
		if q.Search != "" && !cocokPencarian(s, q.Search) {
			continue
		}
		hasil = append(hasil, s)
	}

	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "name":
			lebihKecil = hasil[i].Name < hasil[j].Name
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		case "created_at":
			lebihKecil = hasil[i].CreatedAt.Before(hasil[j].CreatedAt)
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}
	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func getStudent(c fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	return ok(c, "mahasiswa ditemukan", students[i])
}

func createStudent(c fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)
	if !validasiNama(req.Name) {
		errs["name"] = "wajib diisi"
	}
	if !validasiNIM(req.NIM) {
		errs["nim"] = "wajib diisi"
	}
	if !validasiGrade(req.Grade) {
		errs["grade"] = "harus di antara 0 sampai 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}
	if nimDipakai(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}

	baru := Student{
		ID:        nextID,
		Name:      req.Name,
		NIM:       req.NIM,
		Grade:     req.Grade,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	students = append(students, baru)
	nextID++
	return created(c, "mahasiswa berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func replaceStudent(c fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)
	if !validasiNama(req.Name) {
		errs["name"] = "wajib diisi pada PUT"
	}
	if !validasiNIM(req.NIM) {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if !validasiGrade(req.Grade) {
		errs["grade"] = "harus di antara 0 sampai 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	if nimDipakai(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}

	students[i].Name = req.Name
	students[i].NIM = req.NIM
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive
	return ok(c, "mahasiswa berhasil diganti seluruhnya", students[i])
}

func patchStudent(c fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.Name == nil && req.NIM == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}
	if req.Name != nil {
		if !validasiNama(*req.Name) {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		students[i].Name = *req.Name
	}
	if req.NIM != nil {
		if !validasiNIM(*req.NIM) {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		if nimDipakai(*req.NIM, id) {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
		}
		students[i].NIM = *req.NIM
	}
	if req.Grade != nil {
		if !validasiGrade(*req.Grade) {
			return failValidation(c, map[string]string{"grade": "harus di antara 0 sampai 100"})
		}
		students[i].Grade = *req.Grade
	}
	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}
	return ok(c, "mahasiswa berhasil diperbarui sebagian", students[i])
}

func deleteStudent(c fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)
	return noContent(c)
}
