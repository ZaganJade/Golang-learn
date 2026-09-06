package service

import (
	"testing"

	"backend-golang/Minggu4/Latihan/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.
func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
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

func TestApplyPatch(t *testing.T) {
	initial := model.User{ID: 1, Username: "sari", Email: "sari@mail.com", IsActive: true}
	inactive := false

	result, errs := ApplyPatch(initial, model.PatchUserRequest{IsActive: &inactive})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Username != "sari" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestValidateCreate(t *testing.T) {
	errs := ValidateCreate(model.CreateUserRequest{
		Username: "  ", Email: "bukan-email", Password: "pendek",
	})
	if len(errs) != 3 {
		t.Fatalf("seharusnya ada 3 field bermasalah, dapat %d: %v", len(errs), errs)
	}
	if _, ok := errs["username"]; !ok {
		t.Error("username kosong seharusnya ditandai")
	}
	if _, ok := errs["email"]; !ok {
		t.Error("email salah seharusnya ditandai")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("password pendek seharusnya ditandai")
	}

	ok := ValidateCreate(model.CreateUserRequest{
		Username: "sari", Email: "sari@mail.com", Password: "rahasia123",
	})
	if len(ok) != 0 {
		t.Errorf("isi yang benar seharusnya lolos, dapat %v", ok)
	}
}
