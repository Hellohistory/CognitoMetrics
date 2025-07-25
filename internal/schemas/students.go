// Package schemas file: internal/schemas/students.go
package schemas

import "time"

// StudentCreate 对应 StudentCreate
type StudentCreate struct {
	Name    string `json:"name" binding:"required"`
	ClassID uint   `json:"class_id" binding:"required"`
}

// StudentCreateBatch 对应 StudentCreateBatch
type StudentCreateBatch struct {
	Students []StudentCreate `json:"students" binding:"required,gt=0,dive"` // dive会校验切片内每个元素
}

// StudentUpdate 对应 StudentUpdate
type StudentUpdate struct {
	Name    string `json:"name,omitempty"`
	ClassID *uint  `json:"class_id,omitempty"` // 使用指针表示可选
}

// StudentSchema 对应 StudentSchema
type StudentSchema struct {
	ID        uint   `json:"id"`
	StudentNo string `json:"student_no"`
	Name      string `json:"name"`
	ClassID   uint   `json:"class_id"`
	IsActive  bool   `json:"is_active"`
}

// StudentBatchClassUpdate 对应批量更新班级
type StudentBatchClassUpdate struct {
	StudentIDs    []uint `json:"student_ids" binding:"required,min=1"`
	TargetClassID uint   `json:"target_class_id" binding:"required"`
}

// StudentBatchStatusUpdate 对应批量更新状态
type StudentBatchStatusUpdate struct {
	StudentIDs []uint `json:"student_ids" binding:"required,min=1"`
	IsActive   bool   `json:"is_active"`
}

// StudentDetailSchema 对应学生详情
type StudentDetailSchema struct {
	ID             uint   `json:"id"`
	StudentNo      string `json:"student_no"`
	Name           string `json:"name"`
	ClassID        uint   `json:"class_id"`
	IsActive       bool   `json:"is_active"`
	ClassName      string `json:"class_name"`
	GradeName      string `json:"grade_name"`
	EnrollmentYear int    `json:"enrollment_year"`
}

// PerformanceRecordSchema 对应单次考试表现
type PerformanceRecordSchema struct {
	ExamID     uint      `json:"exam_id"`
	ExamName   string    `json:"exam_name"`
	ExamDate   time.Time `json:"exam_date"`
	TotalScore *float64  `json:"total_score"`
	ClassRank  *int      `json:"class_rank"`
	GradeRank  *int      `json:"grade_rank"`
}

// StudentPerformanceHistorySchema 对应学生历史表现
type StudentPerformanceHistorySchema struct {
	Records []PerformanceRecordSchema `json:"records"`
}
