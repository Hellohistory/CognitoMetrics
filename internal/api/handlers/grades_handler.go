// file: internal/api/handlers/grades_handler.go
package handlers

import (
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/schemas"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GradeHandler struct {
	Repo *repository.Repository
}

func NewGradeHandler(repo *repository.Repository) *GradeHandler {
	return &GradeHandler{Repo: repo}
}

func (h *GradeHandler) CreateGrade(c *gin.Context) {
	var req schemas.GradeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.Repo.GetGradeByName(req.Name); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "年级名称已存在"})
		return
	}
	grade, err := h.Repo.CreateGrade(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建年级失败"})
		return
	}
	c.JSON(http.StatusCreated, grade)
}

func (h *GradeHandler) ListGrades(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	grades, err := h.Repo.ListGrades(skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取年级列表失败"})
		return
	}
	c.JSON(http.StatusOK, grades)
}

func (h *GradeHandler) UpdateGrade(c *gin.Context) {
	gradeID, _ := strconv.ParseUint(c.Param("grade_id"), 10, 32)
	var req schemas.GradeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	grade, err := h.Repo.GetGradeByID(uint(gradeID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "年级未找到"})
		return
	}
	updatedGrade, err := h.Repo.UpdateGrade(grade, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, updatedGrade)
}

func (h *GradeHandler) DeleteGrade(c *gin.Context) {
	gradeID, _ := strconv.ParseUint(c.Param("grade_id"), 10, 32)
	if err := h.Repo.DeleteGradeByID(uint(gradeID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 可能因为下面有班级而删除失败
		return
	}
	c.Status(http.StatusNoContent)
}
