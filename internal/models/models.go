package models

import (
	"gorm.io/gorm"
	"time"
)

// Exam 考试信息
type Exam struct {
	gorm.Model
	Name     string
	ExamDate time.Time
	Subjects []*Subject `gorm:"many2many:exam_subjects;"`
	Scores   []Score
	Reports  []AnalysisReport
}

// Subject 学科信息
type Subject struct {
	gorm.Model
	Name             string `gorm:"unique"`
	ExamAssociations []ExamSubject
	Scores           []Score
}

// ExamSubject 是 Exam 和 Subject 的连接表，并包含该科目在该次考试的满分
type ExamSubject struct {
	ExamID    uint `gorm:"primaryKey"`
	SubjectID uint `gorm:"primaryKey"`
	FullMark  float64
}

// Class 班级信息
type Class struct {
	gorm.Model
	Name     string `gorm:"unique"`
	Students []Student
}

// Student 学生信息
type Student struct {
	gorm.Model
	Name    string
	ClassID uint
	Class   Class
	Scores  []Score
}

// Score 单条成绩记录
type Score struct {
	gorm.Model
	Score     float64
	StudentID uint
	Student   Student
	ExamID    uint
	Exam      Exam
	SubjectID uint
	Subject   Subject

	// 新增：用于持久化分析结果的字段
	TScore              float64 `gorm:"index"`
	GradePercentileRank float64 `gorm:"index"`
}

// AnalysisReport 分析报告表
type AnalysisReport struct {
	gorm.Model
	ReportName     string
	ExamID         uint
	Status         string
	FullReportData string // Store as JSON string
	ChartData      string // Store as JSON string
	ErrorMessage   string
	Exam           Exam
}
