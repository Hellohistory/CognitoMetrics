// file: internal/models/models.go
package models

import (
	"gorm.io/gorm"
	"time"
)

// --- NEW: 新增 Grade 模型 ---
// Grade 年级信息 (例如: 初一, 高三)
type Grade struct {
	gorm.Model
	Name    string  `gorm:"unique"`
	Classes []Class // 一个年级包含多个班级
}

// Exam 考试信息
type Exam struct {
	gorm.Model
	Name     string
	ExamDate time.Time
	Status   string     // 新增: 用于追踪考试状态 (draft, submitted, completed)
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

// --- UPDATED: 更新 Class 模型 ---
// Class 班级信息
type Class struct {
	gorm.Model
	Name           string `gorm:"index:idx_grade_class_name,unique"` // 复合唯一索引
	EnrollmentYear int
	GradeID        uint      `gorm:"index:idx_grade_class_name,unique"` // 外键，并加入复合索引
	Grade          Grade     // 班级所属的年级
	Students       []Student // 班级包含的学生
}

// Student 学生信息
type Student struct {
	gorm.Model
	Name      string
	StudentNo string `gorm:"uniqueIndex"`  // 学号唯一
	IsActive  bool   `gorm:"default:true"` // 是否在读
	ClassID   uint
	Class     Class
	Scores    []Score
}

// Score 单条成绩记录
type Score struct {
	gorm.Model
	Score     float64
	StudentID uint `gorm:"index:idx_student_exam_subject,unique"` // 复合唯一索引
	ExamID    uint `gorm:"index:idx_student_exam_subject,unique"` // 复合唯一索引
	SubjectID uint `gorm:"index:idx_student_exam_subject,unique"` // 复合唯一索引
	Student   Student
	Exam      Exam
	Subject   Subject

	// 新增：用于持久化分析结果的字段
	TScore              float64 `gorm:"index"`
	GradePercentileRank float64 `gorm:"index"`
}

// AnalysisReport 分析报告表
type AnalysisReport struct {
	gorm.Model
	ReportName        string
	ExamID            uint
	ReportType        string // 新增: 'single' 或 'comparison'
	SourceDescription string // 新增: 用于记录分析范围，便于重试
	Status            string // processing, completed, failed
	ErrorMessage      string
	FullReportData    string // Store as JSON string
	ChartData         string // Store as JSON string
	AIAnalysisStatus  string `gorm:"default:'not_started'"` // 新增: AI分析状态
	AIAnalysisCache   string // 新增: AI分析结果缓存
	Exam              Exam
}
