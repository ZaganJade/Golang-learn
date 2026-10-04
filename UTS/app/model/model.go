package model

import "time"

// AuthUser isi token yang dipasang middleware ke locals.
type AuthUser struct {
	UserID int
	Email  string
	Role   string
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Student adalah data akademik mahasiswa, 1-1 dengan users.
type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Email       string     `json:"email"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"-"`
}

type Course struct {
	ID        int `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

type Enrollment struct {
	ID             int       `json:"id"`
	StudentID      int       `json:"student_id"`
	CourseID       int       `json:"course_id"`
	TahunAkademik  string    `json:"tahun_akademik"`
	CreatedAt      time.Time `json:"created_at"`
}

// EnrollmentDetail dipakai response detail: gabungan enrollment + mata kuliah.
type EnrollmentDetail struct {
	ID            int       `json:"id"`
	TahunAkademik string    `json:"tahun_akademik"`
	KodeMK        string    `json:"kode_mk"`
	NamaMK        string    `json:"nama_mk"`
	SKS           int       `json:"sks"`
	Semester      int       `json:"semester"`
	CreatedAt     time.Time `json:"created_at"`
}

// ---------- request ----------

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// CreateStudentRequest sekaligus membuat akun user (password awal = NIM).
type CreateStudentRequest struct {
	NIM         string   `json:"nim"          validate:"required,numeric,len=12"`
	Nama        string   `json:"nama"         validate:"required,min=3,max=100"`
	Email       string   `json:"email"        validate:"required,email,max=150"`
	Prodi       string   `json:"prodi"        validate:"required,max=100"`
	Angkatan    int      `json:"angkatan"     validate:"required,tahun"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

// UpdateStudentRequest: nim dan email tidak boleh diubah lewat PUT.
type UpdateStudentRequest struct {
	Nama        string   `json:"nama"         validate:"required,min=3,max=100"`
	Prodi       string   `json:"prodi"        validate:"required,max=100"`
	Angkatan    int      `json:"angkatan"     validate:"required,tahun"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"      validate:"required"`
	TahunAkademik string `json:"tahun_akademik" validate:"required,tahunak"`
}

// ---------- query & meta ----------

type StudentQuery struct {
	Page    int
	PerPage int
	Prodi   string
	Angkatan *int
	Search  string
	Sort    string // "nama" atau "-ipk_terakhir"
}

type CourseQuery struct {
	Semester  *int
	Search    string
	Available bool
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// ---------- response envelope ----------

type WebResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    any                    `json:"data,omitempty"`
	Meta    *Meta                  `json:"meta,omitempty"`
	Errors  map[string][]string    `json:"errors,omitempty"`
}
