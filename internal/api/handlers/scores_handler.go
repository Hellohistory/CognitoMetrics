// Package handlers file: internal/api/handlers/scores_handler.go
package handlers

import (
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/schemas"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	Repo *repository.Repository
}

func NewScoreHandler(repo *repository.Repository) *ScoreHandler {
	return &ScoreHandler{Repo: repo}
}

func (h *ScoreHandler) RecordScoresBatch(c *gin.Context) {
	var req schemas.ScoresBatchInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam, err := h.Repo.GetExamByID(req.ExamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试未找到"})
		return
	}
	if exam.Status != "draft" {
		c.JSON(http.StatusForbidden, gin.H{"error": "考试已锁定，无法修改成绩"})
		return
	}

	count, err := h.Repo.BatchUpsertScores(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量录入成绩失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成功保存或更新了 " + strconv.Itoa(count) + " 条成绩记录。"})
}

func (h *ScoreHandler) GetScoresForClass(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	classID, _ := strconv.ParseUint(c.Param("class_id"), 10, 32)

	// 获取班级所有学生，以确保即使没成绩的学生也出现在返回结果中
	students, err := h.Repo.ListStudentsByClass(uint(classID), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取班级学生失败: " + err.Error()})
		return
	}

	// 获取该班级在该场考试中的所有成绩记录
	scores, err := h.Repo.GetScoresForClassInExam(uint(examID), uint(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取成绩失败: " + err.Error()})
		return
	}

	// 将平铺的成绩列表按学生ID分组，方便查找
	scoresByStudent := make(map[uint]map[string]*float64)
	for _, score := range scores {
		if _, ok := scoresByStudent[score.StudentID]; !ok {
			scoresByStudent[score.StudentID] = make(map[string]*float64)
		}
		// 创建一个新的变量来存储分数的地址
		scoreValue := score.Score
		scoresByStudent[score.StudentID][score.Subject.Name] = &scoreValue
	}

	// 构建最终响应结构
	response := make([]schemas.ScoreInput, len(students))
	for i, student := range students {
		studentScores, ok := scoresByStudent[student.ID]
		if !ok {
			studentScores = make(map[string]*float64) // 如果学生没有任何成绩，则为空map
		}
		response[i] = schemas.ScoreInput{
			StudentID:     student.ID,
			SubjectScores: studentScores,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ScoreHandler) RecordSingleScore(c *gin.Context) {
	var req schemas.SingleScoreUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam, err := h.Repo.GetExamByID(req.ExamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试未找到"})
		return
	}
	if exam.Status != "draft" {
		c.JSON(http.StatusForbidden, gin.H{"error": "考试已锁定，无法修改成绩"})
		return
	}

	if err := h.Repo.UpsertSingleScore(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存成绩失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成绩已保存"})
}
