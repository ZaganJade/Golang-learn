package service

import (
	"testing"

	"backend-golang/Minggu7/Tugas7/app/model"
)

func TestValidateCreate(t *testing.T) {
	errs := ValidateCreate(model.CreateStudentRequest{Name: " ", NIM: "", Grade: 150})
	if len(errs) != 3 {
		t.Fatalf("seharusnya 3 field bermasalah, dapat %d: %v", len(errs), errs)
	}
	if _, ok := errs["name"]; !ok {
		t.Error("nama kosong seharusnya ditandai")
	}
	if _, ok := errs["nim"]; !ok {
		t.Error("nim kosong seharusnya ditandai")
	}
	if _, ok := errs["grade"]; !ok {
		t.Error("grade di luar 0-100 seharusnya ditandai")
	}

	if errs := ValidateCreate(model.CreateStudentRequest{Name: "Amba", NIM: "230101", Grade: 90}); len(errs) != 0 {
		t.Errorf("isi yang benar seharusnya lolos, dapat %v", errs)
	}
}

func TestApplyPatch(t *testing.T) {
	awal := model.Student{ID: 1, Name: "Amba", NIM: "230101", Grade: 90, IsActive: true}
	grade := float64(100)

	hasil, errs := ApplyPatch(awal, model.PatchStudentRequest{Grade: &grade})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if hasil.Grade != 100 {
		t.Errorf("grade seharusnya jadi 100, dapat %v", hasil.Grade)
	}
	if hasil.Name != "Amba" || hasil.NIM != "230101" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}

	kosong := ""
	if _, errs := ApplyPatch(awal, model.PatchStudentRequest{Name: &kosong}); len(errs) != 1 {
		t.Errorf("nama kosong seharusnya ditolak, dapat %v", errs)
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{5, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}
