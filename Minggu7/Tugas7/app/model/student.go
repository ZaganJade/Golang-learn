package model

import "time"

type Student struct {
	ID		int		`json:"id"`
	Name		string		`json:"name"`
	NIM		string		`json:"nim"`
	Grade		float64		`json:"grade"`
	IsActive	bool		`json:"is_active"`
	OwnerID		*int		`json:"owner_id"`
	CreatedAt	time.Time	`json:"created_at"`
}

type CreateStudentRequest struct {
	Name	string	`json:"name"  validate:"required,min=1,max=100"`
	NIM	string	`json:"nim"   validate:"required,nim"`
	Grade	float64	`json:"grade" validate:"gte=0,lte=100"`
}

type StudentNilai struct {
	ID_Nilai	int		`json:"id"`
	NameMK		string		`json:"name"`
	ID_Student	int		`json:"id_student"`
	Grade		float64		`json:"grade"`
	IsActive	bool		`json:"is_active"`
	CreatedAt	time.Time	`json:"created_at"`
}

type ReplaceStudentRequest struct {
	Name		string	`json:"name"      validate:"required,min=1,max=100"`
	NIM		string	`json:"nim"       validate:"required,nim"`
	Grade		float64	`json:"grade"     validate:"gte=0,lte=100"`
	IsActive	bool	`json:"is_active"`
}

type PatchStudentRequest struct {
	Name		*string		`json:"name,omitempty"       validate:"omitnil,min=1,max=100"`
	NIM		*string		`json:"nim,omitempty"        validate:"omitnil,nim"`
	Grade		*float64	`json:"grade,omitempty"     validate:"omitnil,gte=0,lte=100"`
	IsActive	*bool		`json:"is_active,omitempty"`
}

type WebResponse struct {
	Success	bool	`json:"success"`
	Message	string	`json:"message"`
	Data	any	`json:"data,omitempty"`

	Meta	any	`json:"meta,omitempty"`
	Errors	any	`json:"errors,omitempty"`
}

type Meta struct {
	Page		int	`json:"page"`
	Limit		int	`json:"limit"`
	Total		int	`json:"total"`
	TotalPages	int	`json:"total_pages"`
}

type ListQuery struct {
	Page		int
	Limit		int
	Search		string
	Sort		string
	Order		string
	IsActive	*bool
	GradeMin	*float64
	GradeMax	*float64
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
