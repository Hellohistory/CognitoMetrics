package handlers

import (
	"CognitoMetrics/internal/models"
	"CognitoMetrics/internal/schemas"
)

func toGradeSchema(grade *models.Grade) schemas.GradeSchema {
	return schemas.GradeSchema{
		ID:   grade.ID,
		Name: grade.Name,
	}
}

func toClassSchema(class *models.Class) schemas.ClassSchema {
	return schemas.ClassSchema{
		ID:             class.ID,
		GradeID:        class.GradeID,
		Name:           class.Name,
		EnrollmentYear: class.EnrollmentYear,
	}
}

func toStudentSchema(student *models.Student) schemas.StudentSchema {
	return schemas.StudentSchema{
		ID:        student.ID,
		StudentNo: student.StudentNo,
		Name:      student.Name,
		ClassID:   student.ClassID,
		IsActive:  student.IsActive,
	}
}

func toExamSchema(exam *models.Exam) schemas.ExamSchema {
	return schemas.ExamSchema{
		ID:       exam.ID,
		Name:     exam.Name,
		ExamDate: exam.ExamDate,
		Status:   exam.Status,
	}
}

func toGradeSchemas(grades []models.Grade) []schemas.GradeSchema {
	result := make([]schemas.GradeSchema, len(grades))
	for i := range grades {
		result[i] = toGradeSchema(&grades[i])
	}
	return result
}

func toStudentSchemas(students []models.Student) []schemas.StudentSchema {
	result := make([]schemas.StudentSchema, len(students))
	for i := range students {
		result[i] = toStudentSchema(&students[i])
	}
	return result
}

func toExamSchemas(exams []models.Exam) []schemas.ExamSchema {
	result := make([]schemas.ExamSchema, len(exams))
	for i := range exams {
		result[i] = toExamSchema(&exams[i])
	}
	return result
}
