// Package schemas file: internal/schemas/reports.go
package schemas

import (
	"encoding/json"
	"time"
)

// Scope 对应分析范围
type Scope struct {
	Level string `json:"level" binding:"required,oneof=FULL_EXAM GRADE CLASS"`
	IDs   []uint `json:"ids"`
}

// AnalysisSubmissionRequest 对应提交分析任务的请求
type AnalysisSubmissionRequest struct {
	ExamID     uint   `json:"exam_id" binding:"required"`
	ReportName string `json:"report_name" binding:"required"`
	Scope      Scope  `json:"scope" binding:"required"`
}

// ReportSubmissionResponse 对应任务提交后的响应
type ReportSubmissionResponse struct {
	Message  string `json:"message"`
	ReportID uint   `json:"report_id"`
}

// ComparisonReportRequest 对应创建对比报告的请求
type ComparisonReportRequest struct {
	ReportIDs  []uint `json:"report_ids" binding:"required,min=2"`
	ReportName string `json:"report_name,omitempty"`
}

// AnalysisReportSchema 对应返回给前端的报告详情
type AnalysisReportSchema struct {
	ID               uint            `json:"id"`
	ReportName       string          `json:"report_name"`
	ExamID           *uint           `json:"exam_id"`
	Status           string          `json:"status"`
	ReportType       string          `json:"report_type"`
	AIAnalysisStatus string          `json:"ai_analysis_status"`
	AIAnalysisCache  json.RawMessage `json:"ai_analysis_cache,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	FullReportData   json.RawMessage `json:"full_report_data,omitempty"`
	ChartData        json.RawMessage `json:"chart_data,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        *time.Time      `json:"updated_at"`
	Exam             *ExamSchema     `json:"exam,omitempty"`
}

// PaginatedAnalysisReportResponse 对应分页查询报告的响应
type PaginatedAnalysisReportResponse struct {
	Items    []AnalysisReportSchema `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}
