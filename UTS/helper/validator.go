package helper

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// validate dibuat SEKALI untuk seluruh aplikasi (cache refleksi).
var validate = newValidator()

var reTahunAkademik = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

func newValidator() *validator.Validate {
	v := validator.New()

	// pesan error memakai nama JSON, bukan nama field Go
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	// angkatan: 4 digit, tidak boleh melebihi tahun berjalan
	_ = v.RegisterValidation("tahun", func(fl validator.FieldLevel) bool {
		t := fl.Field().Int()
		return t >= 1000 && t <= int64(time.Now().Year())
	})

	// tahun akademik: 2026/2027-Ganjil
	_ = v.RegisterValidation("tahunak", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if !reTahunAkademik.MatchString(s) {
			return false
		}
		parts := strings.SplitN(s, "/", 2)
		a, err1 := strconv.Atoi(parts[0])
		b, err2 := strconv.Atoi(strings.SplitN(parts[1], "-", 2)[0])
		if err1 != nil || err2 != nil {
			return false
		}
		return b == a+1
	})
	return v
}

// ValidateStruct menjalankan tag; balasan map[string][]string sesuai format UTS.
func ValidateStruct(s any) map[string][]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string][]string{"_": {"objek yang divalidasi tidak sah"}}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string][]string{"_": {"Validasi gagal"}}
	}

	result := map[string][]string{}
	for _, fe := range fieldErrors {
		result[fe.Field()] = append(result[fe.Field()], messageFor(fe))
	}
	return result
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "nilai maksimal " + fe.Param()
	case "numeric":
		return "harus berupa angka"
	case "len":
		return "harus tepat " + fe.Param() + " digit"
	case "gte":
		return "nilai minimal " + fe.Param()
	case "lte":
		return "nilai maksimal " + fe.Param()
	case "tahun":
		return "angkatan tidak valid (4 digit, tidak melebihi tahun berjalan)"
	case "tahunak":
		return "format tahun akademik tidak valid (contoh: 2026/2027-Ganjil)"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}
