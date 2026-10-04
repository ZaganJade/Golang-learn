package service

import (
	"testing"

	"backend-golang/Minggu7/Tugas7/app/model"
	"backend-golang/Minggu7/Tugas7/helper"
)

func TestValidateStudentCreate(t *testing.T) {
	errs := helper.ValidateStruct(model.CreateStudentRequest{
		Name:	"", NIM: "abc", Grade: 150,
	})
	if errs == nil {
		t.Fatal("seharusnya ada 3 field bermasalah")
	}
	if errs["name"] != "wajib diisi" {
		t.Errorf("name kosong: dapat %v", errs)
	}
	if errs["nim"] != "harus 6 digit angka" {
		t.Errorf("nim bukan 6 digit: dapat %v", errs)
	}
	if errs["grade"] != "nilai maksimal 100" {
		t.Errorf("grade 150: dapat %v", errs)
	}

	if errs := helper.ValidateStruct(model.CreateStudentRequest{
		Name:	"Amba", NIM: "230101", Grade: 90,
	}); errs != nil {
		t.Errorf("isi yang benar seharusnya lolos, dapat %v", errs)
	}
}

func TestApplyStudentPatch(t *testing.T) {
	awal := model.Student{ID: 1, Name: "Amba", NIM: "230101", Grade: 90, IsActive: true}
	grade := float64(100)

	if errs := helper.ValidateStruct(model.PatchStudentRequest{Grade: &grade}); errs != nil {
		t.Fatalf("patch satu field sah seharusnya lolos: %v", errs)
	}

	hasil := ApplyPatch(awal, model.PatchStudentRequest{Grade: &grade})
	if hasil.Grade != 100 {
		t.Errorf("grade seharusnya jadi 100, dapat %v", hasil.Grade)
	}
	if hasil.Name != "Amba" || hasil.NIM != "230101" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}

	kosong := ""
	errs := helper.ValidateStruct(model.PatchStudentRequest{Name: &kosong})
	if errs == nil {
		t.Error("nama kosong pada PATCH seharusnya ditolak")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("body tanpa field seharusnya dianggap kosong")
	}
	isActive := true
	if IsEmptyPatch(model.PatchStudentRequest{IsActive: &isActive}) {
		t.Error("body dengan satu field seharusnya tidak kosong")
	}
}
