package main

import (
	"CognitoMetrics/internal/analyzer"
	"CognitoMetrics/internal/charts"
	"CognitoMetrics/internal/models"
	"CognitoMetrics/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// setupSampleData 检查并填充模拟数据
func setupSampleData(repo *repository.Repository) (uint, error) {
	var examCount int64
	repo.DB.Model(&models.Exam{}).Count(&examCount)
	if examCount > 0 {
		log.Println("Exam data already exists.")
		return 1, nil // Assume we are analyzing the first exam
	}

	log.Println("Seeding sample data...")
	// 1. 创建班级
	class1 := models.Class{Name: "高三(1)班"}
	class2 := models.Class{Name: "高三(2)班"}
	repo.DB.Create(&class1)
	repo.DB.Create(&class2)

	// 2. 创建学生
	students := []models.Student{
		{Name: "张三", ClassID: class1.ID}, {Name: "李四", ClassID: class1.ID}, {Name: "王五", ClassID: class1.ID},
		{Name: "赵六", ClassID: class2.ID}, {Name: "孙七", ClassID: class2.ID}, {Name: "周八", ClassID: class2.ID},
	}
	repo.DB.Create(&students)

	// 3. 创建考试
	exam1 := models.Exam{Name: "2025届高三第一次模拟考", ExamDate: time.Now().AddDate(0, -1, 0)}
	repo.DB.Create(&exam1)
	subjects := []models.Subject{}
	repo.DB.Find(&subjects)
	for _, s := range subjects {
		repo.DB.Create(&models.ExamSubject{ExamID: exam1.ID, SubjectID: s.ID, FullMark: 150})
	}

	// 4. 录入成绩
	scores := []models.Score{
		{StudentID: 1, ExamID: exam1.ID, SubjectID: 1, Score: 125}, {StudentID: 1, ExamID: exam1.ID, SubjectID: 2, Score: 135}, {StudentID: 1, ExamID: exam1.ID, SubjectID: 3, Score: 140},
		{StudentID: 2, ExamID: exam1.ID, SubjectID: 1, Score: 110}, {StudentID: 2, ExamID: exam1.ID, SubjectID: 2, Score: 95}, {StudentID: 2, ExamID: exam1.ID, SubjectID: 3, Score: 125},
		{StudentID: 3, ExamID: exam1.ID, SubjectID: 1, Score: 95}, {StudentID: 3, ExamID: exam1.ID, SubjectID: 2, Score: 145}, {StudentID: 3, ExamID: exam1.ID, SubjectID: 3, Score: 105},
		{StudentID: 4, ExamID: exam1.ID, SubjectID: 1, Score: 130}, {StudentID: 4, ExamID: exam1.ID, SubjectID: 2, Score: 115}, {StudentID: 4, ExamID: exam1.ID, SubjectID: 3, Score: 135},
		{StudentID: 5, ExamID: exam1.ID, SubjectID: 1, Score: 100}, {StudentID: 5, ExamID: exam1.ID, SubjectID: 2, Score: 80}, {StudentID: 5, ExamID: exam1.ID, SubjectID: 3, Score: 90},
		{StudentID: 6, ExamID: exam1.ID, SubjectID: 1, Score: 140}, {StudentID: 6, ExamID: exam1.ID, SubjectID: 2, Score: 148}, {StudentID: 6, ExamID: exam1.ID, SubjectID: 3, Score: 142},
	}
	repo.DB.Create(&scores)

	return exam1.ID, nil
}

func main() {
	log.Println("CognitoMetrics Engine: Initializing...")

	repo, err := repository.New("cognitometrics.db")
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	log.Println("Repository initialized successfully.")

	examID, err := setupSampleData(repo)
	if err != nil {
		log.Fatalf("Failed to setup sample data: %v", err)
	}

	log.Printf("Loading analysis data for Exam ID %d...", examID)
	analysisData, historyData, err := repo.LoadAnalysisData(examID)
	if err != nil {
		log.Fatalf("Failed to load analysis data: %v", err)
	}
	log.Println("Analysis data loaded successfully.")

	log.Println("Starting analysis...")
	report, err := analyzer.PerformAnalysis(analysisData, historyData, repo)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}
	log.Println("Analysis completed! Metric write-back initiated in background.")

	log.Println("Generating chart-friendly data...")
	chartData, err := charts.GenerateChartData(report)
	if err != nil {
		log.Fatalf("Failed to generate chart data: %v", err)
	}
	log.Println("Chart data generated successfully!")

	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	chartJSON, _ := json.MarshalIndent(chartData, "", "  ")

	fmt.Println("\n========================= CognitoMetrics Analysis Report =========================")
	fmt.Println(string(reportJSON))
	fmt.Println("\n=========================== Chart-Friendly Data ===========================")
	fmt.Println(string(chartJSON))

	// 等待一秒，让后台的回写任务有机会执行和打印日志
	time.Sleep(1 * time.Second)
	log.Println("CLI execution finished.")
}
