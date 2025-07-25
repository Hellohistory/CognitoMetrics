// file: internal/api/handlers/reports_handler.go
package handlers

import (
	"CognitoMetrics/internal/analyzer/types"
	"CognitoMetrics/internal/charts"
	"CognitoMetrics/internal/models"
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/schemas"
	"CognitoMetrics/internal/services"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReportHandler 封装了报告相关的处理器及其依赖
type ReportHandler struct {
	Repo         *repository.Repository
	ReportRunner *services.ReportRunner
}

// NewReportHandler 创建一个新的 ReportHandler 实例
func NewReportHandler(repo *repository.Repository, runner *services.ReportRunner) *ReportHandler {
	return &ReportHandler{Repo: repo, ReportRunner: runner}
}

// SubmitAnalysis 处理提交分析任务 (POST /api/analysis/submit)
func (h *ReportHandler) SubmitAnalysis(c *gin.Context) {
	var req schemas.AnalysisSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	exam, err := h.Repo.GetExamByID(req.ExamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试未找到"})
		return
	}

	sourceDesc := fmt.Sprintf("Scope: %s, IDs: %v", req.Scope.Level, req.Scope.IDs)
	newReport, err := h.Repo.CreateAnalysisReport(req.ReportName, req.ExamID, sourceDesc, "single")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建报告记录失败: " + err.Error()})
		return
	}

	// 异步执行分析任务
	go h.ReportRunner.RunSingleExamAnalysisTask(newReport.ID, exam.ID, req.Scope.Level, req.Scope.IDs)

	c.JSON(http.StatusAccepted, schemas.ReportSubmissionResponse{
		Message:  "分析任务已成功提交，正在后台处理。",
		ReportID: newReport.ID,
	})
}

// --- 补全实现 START ---
// ListReports 获取分析报告列表 (GET /api/analysis/reports)
func (h *ReportHandler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	query := c.Query("query")
	status := c.Query("status")
	reportType := c.Query("report_type")

	reports, total, err := h.Repo.ListReports(page, pageSize, query, status, reportType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取报告列表失败: " + err.Error()})
		return
	}

	// 将数据库模型转换为API响应的Schema
	reportSchemas := make([]schemas.AnalysisReportSchema, len(reports))
	for i, report := range reports {
		var examSchema *schemas.ExamSchema
		// 检查关联的Exam是否存在
		if report.Exam.ID != 0 {
			examSchema = &schemas.ExamSchema{
				ID:       report.Exam.ID,
				Name:     report.Exam.Name,
				ExamDate: report.Exam.ExamDate,
				Status:   report.Exam.Status,
			}
		}

		reportSchemas[i] = schemas.AnalysisReportSchema{
			ID:               report.ID,
			ReportName:       report.ReportName,
			ExamID:           &report.ExamID,
			Status:           report.Status,
			ReportType:       report.ReportType,
			AIAnalysisStatus: report.AIAnalysisStatus,
			AIAnalysisCache:  json.RawMessage(report.AIAnalysisCache),
			ErrorMessage:     report.ErrorMessage,
			// 列表视图通常不需要完整的报告数据，保持为空以提高性能
			FullReportData: nil,
			ChartData:      nil,
			CreatedAt:      report.CreatedAt,
			UpdatedAt:      &report.UpdatedAt,
			Exam:           examSchema,
		}
	}

	c.JSON(http.StatusOK, schemas.PaginatedAnalysisReportResponse{
		Items:    reportSchemas,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// --- 补全实现 END ---

// GetReportDetails 获取单个报告详情
func (h *ReportHandler) GetReportDetails(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	report, err := h.Repo.GetReportByID(uint(reportID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "报告未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取报告失败"})
		}
		return
	}

	// 转换模型为 Schema
	var examSchema *schemas.ExamSchema
	if report.Exam.ID != 0 {
		examSchema = &schemas.ExamSchema{
			ID:       report.Exam.ID,
			Name:     report.Exam.Name,
			ExamDate: report.Exam.ExamDate,
			Status:   report.Exam.Status,
		}
	}

	resp := schemas.AnalysisReportSchema{
		ID:               report.ID,
		ReportName:       report.ReportName,
		ExamID:           &report.ExamID,
		Status:           report.Status,
		ReportType:       report.ReportType,
		AIAnalysisStatus: report.AIAnalysisStatus,
		AIAnalysisCache:  json.RawMessage(report.AIAnalysisCache),
		ErrorMessage:     report.ErrorMessage,
		FullReportData:   json.RawMessage(report.FullReportData),
		CreatedAt:        report.CreatedAt,
		UpdatedAt:        &report.UpdatedAt,
		Exam:             examSchema,
	}

	c.JSON(http.StatusOK, resp)
}

// SubmitAIAnalysis ... (此函数保持不变)
func (h *ReportHandler) SubmitAIAnalysis(c *gin.Context) {
	reportID, _ := strconv.ParseUint(c.Param("report_id"), 10, 32)

	report, err := h.Repo.GetReportByID(uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告未找到"})
		return
	}
	if report.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "主报告尚未完成，无法进行AI分析。"})
		return
	}
	if report.AIAnalysisStatus == "processing" {
		c.JSON(http.StatusConflict, gin.H{"error": "AI分析任务已在处理中，请勿重复提交。"})
		return
	}
	if report.AIAnalysisStatus == "completed" && report.AIAnalysisCache != "" {
		c.JSON(http.StatusOK, gin.H{
			"message":            "AI分析已完成，直接从缓存返回。",
			"report_id":          report.ID,
			"ai_analysis_status": report.AIAnalysisStatus,
			"analysis":           json.RawMessage(report.AIAnalysisCache),
		})
		return
	}

	h.Repo.UpdateReportAIStatus(report.ID, "processing")
	go h.ReportRunner.RunAIAnalysisTask(report.ID)

	c.JSON(http.StatusAccepted, gin.H{
		"message":            "AI分析任务已成功提交，正在后台处理。",
		"report_id":          report.ID,
		"ai_analysis_status": "processing",
	})
}

// 辅助函数：获取一个已完成的报告，并处理通用错误
func (h *ReportHandler) getCompletedReport(c *gin.Context) (*models.AnalysisReport, *types.AnalysisReport) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return nil, nil
	}

	report, err := h.Repo.GetReportByID(uint(reportID))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "报告未找到"})
		return nil, nil
	}

	if report.Status != "completed" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("报告状态为 '%s'，尚未分析完成。", report.Status)})
		return nil, nil
	}
	if report.FullReportData == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "报告数据为空。"})
		return nil, nil
	}

	var reportData types.AnalysisReport
	if err := json.Unmarshal([]byte(report.FullReportData), &reportData); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "解析报告数据失败"})
		return nil, nil
	}

	return report, &reportData
}

// CompareReports 创建对比分析报告 (POST /api/analysis/compare)
func (h *ReportHandler) CompareReports(c *gin.Context) {
	var req schemas.ComparisonReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	reportName := req.ReportName
	if reportName == "" {
		reportName = fmt.Sprintf("对 %d 场考试的对比分析", len(req.ReportIDs))
	}

	sourceDesc := fmt.Sprintf("Comparing reports: %v", req.ReportIDs)
	newReport, err := h.Repo.CreateAnalysisReport(reportName, 0, sourceDesc, "comparison")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对比报告失败"})
		return
	}

	c.JSON(http.StatusAccepted, schemas.ReportSubmissionResponse{
		Message:  "对比分析任务已创建。",
		ReportID: newReport.ID,
	})
}

// GetReportGroupStats 获取报告的整体统计数据
func (h *ReportHandler) GetReportGroupStats(c *gin.Context) {
	_, reportData := h.getCompletedReport(c)
	if reportData == nil {
		return
	}
	c.JSON(http.StatusOK, reportData.GroupStats)
}

// GetReportClassDetails 获取报告中指定班级的详情
func (h *ReportHandler) GetReportClassDetails(c *gin.Context) {
	className := c.Param("class_name")
	_, reportData := h.getCompletedReport(c)
	if reportData == nil {
		return
	}

	for _, table := range reportData.Tables {
		if table.TableName == className {
			c.JSON(http.StatusOK, table)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("在报告中未找到班级 '%s'", className)})
}

// GetReportStudentDetails 获取报告中指定学生的详情
func (h *ReportHandler) GetReportStudentDetails(c *gin.Context) {
	studentName := c.Param("student_name")
	_, reportData := h.getCompletedReport(c)
	if reportData == nil {
		return
	}

	for _, table := range reportData.Tables {
		for _, student := range table.Students {
			if student.StudentName == studentName {
				c.JSON(http.StatusOK, student)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("在报告中未找到学生 '%s'", studentName)})
}

// GetReportChartData 获取报告的图表优化数据
func (h *ReportHandler) GetReportChartData(c *gin.Context) {
	_, reportData := h.getCompletedReport(c)
	if reportData == nil {
		return
	}

	chartData, err := charts.GenerateChartData(reportData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成图表数据失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, chartData)
}

// DeleteReport 删除分析报告
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	err = h.Repo.DeleteReportByID(uint(reportID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "报告未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除报告失败"})
		return
	}

	c.Status(http.StatusNoContent)
}

// RetryAnalysis 重试失败的分析任务
func (h *ReportHandler) RetryAnalysis(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	report, err := h.Repo.GetReportByID(uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告未找到"})
		return
	}

	if report.Status != "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有失败的任务才能重试"})
		return
	}
	if report.SourceDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "报告缺少源描述，无法重试"})
		return
	}

	reLevel := regexp.MustCompile(`Scope: (\w+)`)
	reIDs := regexp.MustCompile(`IDs: \[(.*?)]`)

	levelMatch := reLevel.FindStringSubmatch(report.SourceDescription)
	idsMatch := reIDs.FindStringSubmatch(report.SourceDescription)

	if len(levelMatch) < 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析重试参数失败: 无法找到 scope_level"})
		return
	}
	scopeLevel := levelMatch[1]

	var scopeIDs []uint
	if len(idsMatch) >= 2 && idsMatch[1] != "" {
		reNum := regexp.MustCompile(`\d+`)
		idStrs := reNum.FindAllString(idsMatch[1], -1)
		for _, s := range idStrs {
			id, _ := strconv.ParseUint(s, 10, 32)
			scopeIDs = append(scopeIDs, uint(id))
		}
	}

	if err := h.Repo.UpdateReportStatus(report.ID, "processing", ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新报告状态失败"})
		return
	}

	go h.ReportRunner.RunSingleExamAnalysisTask(report.ID, report.ExamID, scopeLevel, scopeIDs)

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "任务已重新提交，正在后台处理。",
		"report_id": report.ID,
	})
}
