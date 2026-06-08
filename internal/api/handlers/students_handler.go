// Package handlers file: internal/api/handlers/students_handler.go
package handlers

import (
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/schemas"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudentHandler struct {
	Repo *repository.Repository
}

func NewStudentHandler(repo *repository.Repository) *StudentHandler {
	return &StudentHandler{Repo: repo}
}

// CreateStudent 新增单个学生
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var req schemas.StudentCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentNos, err := h.Repo.GenerateNewStudentNumbers(req.ClassID, 1)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 可能是班级未找到
		return
	}

	students, err := h.Repo.BatchCreateStudents([]schemas.StudentCreate{req}, studentNos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toStudentSchema(&students[0]))
}

// CreateStudentsBatch 批量新增学生
func (h *StudentHandler) CreateStudentsBatch(c *gin.Context) {
	var req schemas.StudentCreateBatch
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Students) == 0 {
		c.JSON(http.StatusOK, []schemas.StudentSchema{})
		return
	}

	classID := req.Students[0].ClassID
	// 校验所有学生是否属于同一个班级
	for _, s := range req.Students {
		if s.ClassID != classID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "批量创建的学生必须属于同一个班级"})
			return
		}
	}

	studentNos, err := h.Repo.GenerateNewStudentNumbers(classID, len(req.Students))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	newStudents, err := h.Repo.BatchCreateStudents(req.Students, studentNos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量创建学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toStudentSchemas(newStudents))
}

// ListStudentsByClass 获取班级下的学生列表
func (h *StudentHandler) ListStudentsByClass(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("class_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的班级ID"})
		return
	}
	includeInactive := c.Query("include_inactive") == "true"

	students, err := h.Repo.ListStudentsByClass(uint(classID), includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
		return
	}
	c.JSON(http.StatusOK, toStudentSchemas(students))
}

// UpdateStudent 更新学生信息（如改名、换班）
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	// 1. 从URL获取学生ID
	studentID, err := strconv.ParseUint(c.Param("student_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的学生ID"})
		return
	}

	// 2. 绑定请求体
	var req schemas.StudentUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. 获取要更新的学生实体
	student, err := h.Repo.GetStudentByID(uint(studentID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	// 4. 调用仓库层的方法执行更新
	updatedStudent, err := h.Repo.UpdateStudent(student, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败: " + err.Error()})
		return
	}

	// 5. 返回更新后的学生信息
	c.JSON(http.StatusOK, toStudentSchema(updatedStudent))
}

// ActivateStudent 激活学生
func (h *StudentHandler) ActivateStudent(c *gin.Context) {
	studentID, _ := strconv.ParseUint(c.Param("student_id"), 10, 32)
	student, err := h.Repo.UpdateStudentStatus(uint(studentID), true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, toStudentSchema(student))
}

// DeactivateStudent 停用学生
func (h *StudentHandler) DeactivateStudent(c *gin.Context) {
	studentID, _ := strconv.ParseUint(c.Param("student_id"), 10, 32)
	student, err := h.Repo.UpdateStudentStatus(uint(studentID), false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, toStudentSchema(student))
}

// BatchUpdateStatus 批量更新学生状态
func (h *StudentHandler) BatchUpdateStatus(c *gin.Context) {
	var req schemas.StudentBatchStatusUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.BatchUpdateStudentsStatus(req.StudentIDs, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量更新状态失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功为 %d 名学生更新了在读状态。", len(req.StudentIDs))})
}

// BatchUpdateClass 批量更新学生班级
func (h *StudentHandler) BatchUpdateClass(c *gin.Context) {
	var req schemas.StudentBatchClassUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.BatchUpdateStudentsClass(req.StudentIDs, req.TargetClassID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量更新班级失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功为 %d 名学生更新了班级。", len(req.StudentIDs))})
}

// GetStudentDetails 获取单个学生详细信息
func (h *StudentHandler) GetStudentDetails(c *gin.Context) {
	studentID, _ := strconv.ParseUint(c.Param("student_id"), 10, 32)
	details, err := h.Repo.GetStudentDetailsByID(uint(studentID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取详情失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// GetStudentPerformanceHistory 获取学生个人表现历史
func (h *StudentHandler) GetStudentPerformanceHistory(c *gin.Context) {
	studentID, _ := strconv.ParseUint(c.Param("student_id"), 10, 32)
	records, err := h.Repo.GetStudentPerformanceHistory(uint(studentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史表现失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, schemas.StudentPerformanceHistorySchema{Records: records})
}
