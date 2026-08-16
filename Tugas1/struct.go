package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("[%d] %s - Nilai: %.1f - Aktif: %v", s.ID, s.Name, s.Grade, s.IsActive)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	s := Student{ID: 1, Name: "Zagan Jade", Grade: 75, IsActive: false}
	fmt.Println(s.GetInfo())

	s.UpdateGrade(92.5)
	s.Activate()
	fmt.Println(s.GetInfo())
}