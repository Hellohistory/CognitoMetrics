// Package api file: internal/api/router.go
package api

import (
	"CognitoMetrics/internal/api/handlers"
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/services"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有API路由
func SetupRouter(repo *repository.Repository, runner *services.ReportRunner) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	reportHandler := handlers.NewReportHandler(repo, runner)
	gradeHandler := handlers.NewGradeHandler(repo)
	classHandler := handlers.NewClassHandler(repo)
	studentHandler := handlers.NewStudentHandler(repo)
	examHandler := handlers.NewExamHandler(repo)
	scoreHandler := handlers.NewScoreHandler(repo)

	api := router.Group("/api")
	{
		api.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "欢迎使用由 Hellohistory 开发设计的分析系统 API"})
		})

		// 学情分析路由
		analysis := api.Group("/analysis")
		{
			analysis.POST("/submit", reportHandler.SubmitAnalysis)
			analysis.POST("/compare", reportHandler.CompareReports)
			analysis.GET("/reports", reportHandler.ListReports)
			analysis.GET("/reports/:report_id", reportHandler.GetReportDetails)
			analysis.DELETE("/reports/:report_id", reportHandler.DeleteReport)
			analysis.POST("/reports/:report_id/retry", reportHandler.RetryAnalysis)
			analysis.GET("/reports/:report_id/group-stats", reportHandler.GetReportGroupStats)
			analysis.GET("/reports/:report_id/class/:class_name", reportHandler.GetReportClassDetails)
			analysis.GET("/reports/:report_id/student/:student_name", reportHandler.GetReportStudentDetails)
			analysis.GET("/reports/:report_id/charts", reportHandler.GetReportChartData)
			analysis.POST("/reports/:report_id/ai-analysis", reportHandler.SubmitAIAnalysis)
		}

		// 学生路由
		students := api.Group("/students")
		{
			students.POST("", studentHandler.CreateStudent)
			students.POST("/batch", studentHandler.CreateStudentsBatch)
			students.GET("/by_class/:class_id", studentHandler.ListStudentsByClass)
			students.PUT("/:student_id", studentHandler.UpdateStudent)
			students.PUT("/:student_id/activate", studentHandler.ActivateStudent)
			students.PUT("/:student_id/deactivate", studentHandler.DeactivateStudent)
			students.POST("/batch-update-status", studentHandler.BatchUpdateStatus)
			students.POST("/batch-update-class", studentHandler.BatchUpdateClass)
			students.GET("/:student_id/details", studentHandler.GetStudentDetails)
			students.GET("/:student_id/performance", studentHandler.GetStudentPerformanceHistory)
		}

		// 班级路由
		classes := api.Group("/classes")
		{
			classes.POST("", classHandler.CreateClass)
			classes.GET("", classHandler.ListClasses)
			classes.GET("/tree", classHandler.GetClassTree)
			classes.GET("/:class_id", classHandler.GetClassByID)
			classes.PUT("/:class_id", classHandler.UpdateClass)
			classes.DELETE("/:class_id", classHandler.DeleteClass)
		}

		// 年级路由
		grades := api.Group("/grades")
		{
			grades.POST("", gradeHandler.CreateGrade)
			grades.GET("", gradeHandler.ListGrades)
			grades.PUT("/:grade_id", gradeHandler.UpdateGrade)
			grades.DELETE("/:grade_id", gradeHandler.DeleteGrade)
		}

		// 考试路由
		exams := api.Group("/exams")
		{
			exams.POST("", examHandler.CreateExam)
			exams.GET("", examHandler.ListExams)
			exams.GET("/:exam_id", examHandler.GetExamDetails)
			exams.PUT("/:exam_id/unlock", examHandler.UnlockExam)
			exams.PUT("/:exam_id/finalize", examHandler.FinalizeExam)
			exams.DELETE("/:exam_id", examHandler.DeleteExam)
		}

		// 成绩路由
		scores := api.Group("/scores")
		{
			scores.POST("/batch", scoreHandler.RecordScoresBatch)
			scores.PUT("/single", scoreHandler.RecordSingleScore)
			scores.GET("/exam/:exam_id/class/:class_id", scoreHandler.GetScoresForClass)
		}
	}

	return router
}
