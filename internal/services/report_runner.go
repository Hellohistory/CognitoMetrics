// file: internal/services/report_runner.go
package services

import (
	"CognitoMetrics/internal/analyzer"
	"CognitoMetrics/internal/repository"
	"fmt"
	"log"
)

// ReportRunner 封装了所有后台任务的依赖
type ReportRunner struct {
	Repo       *repository.Repository
	AIAnalyzer *AIAnalyzer
}

// NewReportRunner 创建 ReportRunner 实例
func NewReportRunner(repo *repository.Repository, aiAnalyzer *AIAnalyzer) *ReportRunner {
	return &ReportRunner{Repo: repo, AIAnalyzer: aiAnalyzer}
}

// RunSingleExamAnalysisTask 在后台执行单场考试分析
func (r *ReportRunner) RunSingleExamAnalysisTask(reportID uint, examID uint, scopeLevel string, scopeIDs []uint) {
	log.Printf("后台任务启动：分析报告 ID %d", reportID)

	// 使用 defer-recover 机制确保即使发生 panic，也能更新报告状态
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("后台任务发生严重错误 (panic): %v", rec)
			r.Repo.UpdateReportStatus(reportID, "failed", fmt.Sprintf("Panic: %v", rec))
		}
	}()

	// 1. 加载数据
	analysisData, historyData, err := r.Repo.LoadAnalysisData(examID, scopeLevel, scopeIDs)
	if err != nil {
		log.Printf("后台任务失败：报告ID %d, 加载数据错误: %v", reportID, err)
		r.Repo.UpdateReportStatus(reportID, "failed", fmt.Sprintf("加载数据失败: %v", err))
		return
	}

	// 2. 执行核心分析
	// 注意：这里的 repo 传递的是 ReportRunner 内部的 repo
	reportData, err := analyzer.PerformAnalysis(analysisData, historyData, r.Repo)
	if err != nil {
		log.Printf("后台任务失败：报告ID %d, 分析过程错误: %v", reportID, err)
		r.Repo.UpdateReportStatus(reportID, "failed", fmt.Sprintf("分析过程失败: %v", err))
		return
	}

	// 3. 保存分析结果
	if err := r.Repo.SaveFullReport(reportID, reportData); err != nil {
		log.Printf("后台任务失败：报告ID %d, 保存报告错误: %v", reportID, err)
		r.Repo.UpdateReportStatus(reportID, "failed", fmt.Sprintf("保存报告失败: %v", err))
		return
	}

	// 4. 更新报告和考试状态为 "completed"
	r.Repo.UpdateReportStatus(reportID, "completed", "")
	if scopeLevel == "FULL_EXAM" {
		// 你可能需要在 repository 中添加一个 UpdateExamStatus 的方法
		// r.Repo.UpdateExamStatus(examID, "completed")
	}

	log.Printf("后台任务成功完成：报告 ID %d", reportID)
}

// RunAIAnalysisTask 在后台执行AI分析
func (r *ReportRunner) RunAIAnalysisTask(reportID uint) {
	log.Printf("后台AI分析任务启动：报告 ID %d", reportID)

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("后台AI分析任务发生严重错误 (panic): %v", rec)
			r.Repo.UpdateReportAIStatus(reportID, "failed")
		}
	}()

	if _, err := r.AIAnalyzer.getOrGenerateAIAnalysis(reportID); err != nil {
		log.Printf("后台AI分析任务失败：报告 ID %d, 错误: %v", reportID, err)
		// 错误状态已在 getOrGenerateAIAnalysis 内部处理
		r.Repo.UpdateReportAIStatus(reportID, "failed")
	} else {
		log.Printf("后台AI分析任务成功完成：报告 ID %d", reportID)
	}
}
