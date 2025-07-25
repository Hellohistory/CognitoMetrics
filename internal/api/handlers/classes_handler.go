// Package handlers file: internal/api/handlers/classes_handler.go
package handlers

import (
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/schemas"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	Repo *repository.Repository
}

func NewClassHandler(repo *repository.Repository) *ClassHandler {
	return &ClassHandler{Repo: repo}
}

func (h *ClassHandler) CreateClass(c *gin.Context) {
	var req schemas.ClassCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 检查年级是否存在
	if _, err := h.Repo.GetGradeByID(req.GradeID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "所属年级不存在"})
		return
	}
	// 检查同一年的班级名是否重复
	if _, err := h.Repo.GetClassByGradeAndName(req.GradeID, req.Name); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "该年级下已存在同名班级"})
		return
	}

	class, err := h.Repo.CreateClass(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建班级失败"})
		return
	}
	c.JSON(http.StatusCreated, class)
}

func (h *ClassHandler) GetClassTree(c *gin.Context) {
	grades, err := h.Repo.GetClassTreeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取数据失败"})
		return
	}

	// 将 models 转换为 schemas，并计算学生数量
	result := make([]schemas.GradeForTree, len(grades))
	for i, grade := range grades {
		gradeData := schemas.GradeForTree{
			ID:      grade.ID,
			Name:    grade.Name,
			Classes: make([]schemas.ClassForTree, len(grade.Classes)),
		}
		// 按班级名称排序
		sort.Slice(grade.Classes, func(i, j int) bool {
			return grade.Classes[i].Name < grade.Classes[j].Name
		})

		for j, cls := range grade.Classes {
			activeStudentCount := 0
			for _, s := range cls.Students {
				if s.IsActive {
					activeStudentCount++
				}
			}
			gradeData.Classes[j] = schemas.ClassForTree{
				ID:             cls.ID,
				Name:           cls.Name,
				StudentCount:   activeStudentCount,
				EnrollmentYear: cls.EnrollmentYear,
			}
		}
		result[i] = gradeData
	}
	c.JSON(http.StatusOK, result)
}

func (h *ClassHandler) ListClasses(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	classes, err := h.Repo.ListClasses(skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取班级列表失败: " + err.Error()})
		return
	}

	// 虽然前端可能不需要总数，但返回总数是良好实践
	total, err := h.Repo.CountClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取班级总数失败: " + err.Error()})
		return
	}

	// 将 model 转换为 schema
	classSchemas := make([]schemas.ClassSchema, len(classes))
	for i, class := range classes {
		classSchemas[i] = schemas.ClassSchema{
			ID:             class.ID,
			Name:           class.Name,
			EnrollmentYear: class.EnrollmentYear,
			GradeID:        class.GradeID,
		}
	}

	// 返回包含总数的分页结构
	c.JSON(http.StatusOK, gin.H{
		"items": classSchemas,
		"total": total,
	})
}

func (h *ClassHandler) GetClassByID(c *gin.Context) {
	classID, _ := strconv.ParseUint(c.Param("class_id"), 10, 32)
	class, err := h.Repo.GetClassByID(uint(classID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "班级未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, class)
}

func (h *ClassHandler) UpdateClass(c *gin.Context) {
	classID, _ := strconv.ParseUint(c.Param("class_id"), 10, 32)
	var req schemas.ClassUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	class, err := h.Repo.GetClassByID(uint(classID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "班级未找到"})
		return
	}
	// 检查名称冲突
	if req.Name != "" && req.Name != class.Name {
		if _, err := h.Repo.GetClassByGradeAndName(class.GradeID, req.Name); err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "该年级下已存在同名班级"})
			return
		}
	}
	updatedClass, err := h.Repo.UpdateClass(class, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, updatedClass)
}

func (h *ClassHandler) DeleteClass(c *gin.Context) {
	classID, _ := strconv.ParseUint(c.Param("class_id"), 10, 32)
	if err := h.Repo.DeleteClassByID(uint(classID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
