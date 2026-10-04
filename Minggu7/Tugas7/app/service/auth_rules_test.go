package service

import (
	"testing"

	"backend-golang/Minggu7/Tugas7/app/model"
)

func TestValidateRegister(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username:	"ab", Email: "salah", Password: "password1",
	})
	if len(errs) != 3 {
		t.Fatalf("seharusnya 3 field bermasalah, dapat %d: %v", len(errs), errs)
	}
	if _, ok := errs["username"]; !ok {
		t.Error("username terlalu pendek seharusnya ditandai")
	}
	if _, ok := errs["email"]; !ok {
		t.Error("email salah seharusnya ditandai")
	}
	if errs["password"] != "password terlalu umum" {
		t.Errorf("password umum seharusnya ditandai, dapat %q", errs["password"])
	}

	if errs := ValidateRegister(model.RegisterRequest{
		Username:	"budi", Email: "budi@example.com", Password: "aman12345",
	}); len(errs) != 0 {
		t.Errorf("isi yang benar seharusnya lolos, dapat %v", errs)
	}
}

func TestValidateLogin(t *testing.T) {
	errs := ValidateLogin(model.LoginRequest{Username: " ", Password: ""})
	if len(errs) != 2 {
		t.Fatalf("seharusnya 2 field kosong ditandai, dapat %d: %v", len(errs), errs)
	}
	if errs := ValidateLogin(model.LoginRequest{Username: "sari", Password: "x"}); len(errs) != 0 {
		t.Errorf("isi lengkap seharusnya lolos, dapat %v", errs)
	}
}

func TestPasswordStrengthRules(t *testing.T) {
	cases := []struct {
		password	string
		want		string
	}{
		{"pendek1", "minimal 8 karakter"},
		{"tanpaangka", "harus memuat huruf dan angka"},
		{"12345678", "harus memuat huruf dan angka"},
		{"admin123", "password terlalu umum"},
		{"aman12345", ""},
	}
	for _, tc := range cases {
		if got := checkPasswordStrength(tc.password); got != tc.want {
			t.Errorf("password %q: harap %q, dapat %q", tc.password, tc.want, got)
		}
	}
}
