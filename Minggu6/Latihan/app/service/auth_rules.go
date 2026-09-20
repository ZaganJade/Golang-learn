package service

import (
	"strings"
	"unicode"

	"backend-golang/Minggu6/Latihan/app/model"
	"backend-golang/Minggu6/Latihan/helper"
)

const minPasswordLength = 8

func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}
	username := strings.TrimSpace(req.Username)
	switch {
	case username == "":
		errs["username"] = "wajib diisi"
	case len(username) < 3:
		errs["username"] = "minimal 3 karakter"
	case !isValidUsername(username):
		errs["username"] = "hanya boleh huruf, angka, titik, dan garis bawah"
	}
	if !isValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}
	if msg := checkPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}
	return errs
}

// ValidateLogin hanya memeriksa kelengkapan, BUKAN kekuatan password.
// Aturan kekuatan tidak diberlakukan di sini karena password lama
// mungkin dibuat sebelum aturannya berubah.
func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}
	return errs
}

func checkPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	// Daftar ini sengaja sangat pendek. Sistem sungguhan memakai daftar
	// berisi jutaan password yang pernah bocor.
	weak := map[string]bool{
		"password1": true, "12345678": true, "qwerty123": true,
		"admin123": true, "password123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}
	return ""
}

func isValidUsername(username string) bool {
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
			return false
		}
	}
	return true
}

// CanAccessUser memutuskan apakah seseorang boleh menyentuh data user lain.
//
// Dua jalur yang diizinkan:
//  1. Kepemilikan (ownership) — data itu miliknya sendiri.
//  2. Permission — role-nya memang berhak atas data siapa pun.
//
// Urutannya disengaja: pemeriksaan kepemilikan didahulukan karena paling
// murah dan paling sering benar. Bila keduanya gagal, jawabannya false.
func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

// ValidateAssignRole memeriksa permintaan pergantian role.
//
// Perhatikan aturan terakhir: seseorang tidak boleh mengubah role dirinya
// sendiri. Tanpa aturan itu, satu-satunya admin dapat menurunkan dirinya
// sendiri menjadi user biasa dan sistem kehilangan admin selamanya.
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}

	if !perms.IsKnownRole(role) {
		errs["role"] = "role tidak dikenal, pilih salah satu dari: " +
			strings.Join(perms.KnownRoles(), ", ")
	}

	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role diri sendiri"
	}

	return errs
}
