package helper

import "testing"

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
		if got := passwordStrength(tc.password); got != tc.want {
			t.Errorf("password %q: harap %q, dapat %q", tc.password, tc.want, got)
		}
	}
}

func TestValidateStructMemakaiNamaJSON(t *testing.T) {
	type contoh struct {
		Username string `json:"username" validate:"required,min=3"`
	}

	errs := ValidateStruct(contoh{Username: "ab"})
	if errs == nil {
		t.Fatal("username 2 karakter seharusnya ditolak")
	}
	if msg, ok := errs["username"]; !ok || msg != "minimal 3 karakter" {
		t.Errorf("kunci seharusnya nama JSON \"username\", dapat %v", errs)
	}

	if errs := ValidateStruct(contoh{Username: "budi"}); errs != nil {
		t.Errorf("isi benar seharusnya lolos, dapat %v", errs)
	}
}

func TestOmitnilPointer(t *testing.T) {
	type contoh struct {
		Name *string `json:"name,omitempty" validate:"omitnil,min=3"`
	}

	if errs := ValidateStruct(contoh{Name: nil}); errs != nil {
		t.Errorf("nil seharusnya dilewati, dapat %v", errs)
	}

	kosong := ""
	errs := ValidateStruct(contoh{Name: &kosong})
	if errs == nil {
		t.Fatal("string kosong pada pointer seharusnya ditolak")
	}
	if errs["name"] != "minimal 3 karakter" {
		t.Errorf("pesan tidak sesuai, dapat %v", errs)
	}
}
