// Package schemas file: internal/schemas/exams_and_scores.go
package schemas

import "time"

// SubjectInExamCreate 对应创建考试时附带的学科信息
type SubjectInExamCreate struct {
	Name     string  `json:"name" binding:"required"`
	FullMark float64 `json:"full_mark" binding:"required,gt=0"`
}

// ExamWithSubjectsCreate 对应创建一场新考试的请求体
type ExamWithSubjectsCreate struct {
	Name     string                `json:"name" binding:"required"`
	ExamDate string                `json:"exam_date" binding:"required"`
	Subjects []SubjectInExamCreate `json:"subjects" binding:"required,min=1,dive"`
}

func (e ExamWithSubjectsCreate) ParsedExamDate() (time.Time, error) {
	if t, err := time.Parse("2006-01-02", e.ExamDate); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, e.ExamDate)
}

// ExamSchema 对应考试的基本信息
type ExamSchema struct {
	ID       uint      `json:"id"`
	Name     string    `json:"name"`
	ExamDate time.Time `json:"exam_date"`
	Status   string    `json:"status"`
}

// ExamSubjectDetailSchema 对应考试详情中的科目
type ExamSubjectDetailSchema struct {
	Name     string  `json:"name"`
	FullMark float64 `json:"full_mark"`
}

// ExamDetailSchema 对应考试的详细信息
type ExamDetailSchema struct {
	ExamSchema
	Subjects []ExamSubjectDetailSchema `json:"subjects"`
}

// ScoreInput 对应批量录入成绩时的单条记录
type ScoreInput struct {
	StudentID     uint                `json:"student_id" binding:"required"`
	SubjectScores map[string]*float64 `json:"subject_scores"` // 使用指针表示成绩可以为 null (缺考)
}

// ScoresBatchInput 对应批量录入成绩的请求体
type ScoresBatchInput struct {
	ExamID uint         `json:"exam_id" binding:"required"`
	Scores []ScoreInput `json:"scores" binding:"required,min=1,dive"`
}

// SingleScoreUpdate 对应单条成绩更新
type SingleScoreUpdate struct {
	ExamID      uint     `json:"exam_id" binding:"required"`
	StudentID   uint     `json:"student_id" binding:"required"`
	SubjectName string   `json:"subject_name" binding:"required"`
	Score       *float64 `json:"score"` // 使用指针表示成绩可以为 null
}
