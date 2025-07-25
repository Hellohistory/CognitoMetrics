// Package schemas file: internal/schemas/classes.go
package schemas

type GradeBase struct {
	Name string `json:"name" binding:"required"`
}

type GradeCreate struct {
	GradeBase
}

type GradeUpdate struct {
	Name string `json:"name,omitempty"`
}

type GradeSchema struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ClassBase struct {
	Name           string `json:"name" binding:"required"`
	EnrollmentYear int    `json:"enrollment_year" binding:"required"`
}

type ClassCreate struct {
	ClassBase
	GradeID uint `json:"grade_id" binding:"required"`
}

type ClassUpdate struct {
	Name           string `json:"name,omitempty"`
	EnrollmentYear *int   `json:"enrollment_year,omitempty"`
}

type ClassSchema struct {
	ID             uint   `json:"id"`
	GradeID        uint   `json:"grade_id"`
	Name           string `json:"name"`
	EnrollmentYear int    `json:"enrollment_year"`
}

type ClassForTree struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	StudentCount   int    `json:"student_count"`
	EnrollmentYear int    `json:"enrollment_year"`
}

type GradeForTree struct {
	ID      uint           `json:"id"`
	Name    string         `json:"name"`
	Classes []ClassForTree `json:"classes"`
}
