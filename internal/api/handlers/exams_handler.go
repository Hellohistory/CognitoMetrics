// Package handlers file: internal/api/handlers/exams_handler.go
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

type ExamHandler struct {
	Repo *repository.Repository
}

func NewExamHandler(repo *repository.Repository) *ExamHandler {
	return &ExamHandler{Repo: repo}
}

func (h *ExamHandler) CreateExam(c *gin.Context) {
	var req schemas.ExamWithSubjectsCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exam, err := h.Repo.CreateExamWithSubjects(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建考试失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exam)
}

func (h *ExamHandler) ListExams(c *gin.Context) {
	exams, err := h.Repo.ListExams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取考试列表失败"})
		return
	}
	c.JSON(http.StatusOK, exams)
}

func (h *ExamHandler) GetExamDetails(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	details, err := h.Repo.GetExamDetailsByID(uint(examID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "考试未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取考试详情失败"})
		return
	}
	c.JSON(http.StatusOK, details)
}

func (h *ExamHandler) UnlockExam(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err := h.Repo.UpdateExamStatus(uint(examID), "draft"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解锁考试失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "考试已解锁"})
}

func (h *ExamHandler) FinalizeExam(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	exam, err := h.Repo.GetExamByID(uint(examID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试未找到"})
		return
	}
	if exam.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("只有草稿状态的考试才能被定稿。当前状态: %s", exam.Status)})
		return
	}
	if err := h.Repo.UpdateExamStatus(uint(examID), "completed"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "定稿考试失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "考试已定稿"})
}

func (h *ExamHandler) DeleteExam(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err := h.Repo.DeleteExamByID(uint(examID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 可能是安全检查导致的删除失败
		return
	}
	c.Status(http.StatusNoContent)
}
