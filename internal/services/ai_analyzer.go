// Package services file: internal/services/ai_analyzer.go
package services

import (
	"CognitoMetrics/internal/repository"
	"encoding/json"
	"fmt"
	"sync"
)

// Prompts remains the same.
const (
	PromptSummary = `
# 角色: 顶尖教育数据科学家
# 任务: 基于年级整体统计数据，生成一份高度浓缩的核心洞察摘要。
# 核心指令:
1.  **试卷质量评估**: 首先对 ` + "`difficulty`" + ` 和 ` + "`discriminationIndex`" + ` 做出明确判断，这决定了后续分析的价值。
2.  **学情结构诊断**: 使用 ` + "`highAchieverPenetration`" + `, ` + "`academicCoreDensity`" + `, ` + "`strugglerSupportIndex`" + ` 判断学生群体是“橄榄型”、“哑铃型”还是其他分布，并解释其教学意义。
3.  **学科关联洞察**: 从 ` + "`correlationMatrix`" + ` 中找出最值得关注的一两个关联（或不关联）现象，并提出可能的解释。
4.  **语言风格**: 必须精炼、专业，直接切入要点，总字数控制在300字以内。
# 输入数据: (仅包含 groupStats 和 fullMarks)
` + "```json\n%s\n```"

	PromptComparison = `
# 角色: 资深教学策略顾问
# 任务: 基于所有班级的统计数据，进行横向对比，识别出各班的特色和关键差异。
# 核心指令:
1.  **找出领跑者和落后者**: 对比各班在总分和核心学科上的平均分、优秀率、及格率，直接点名。
2.  **分析班级画像**: 结合 ` + "`quartileCompetitiveness` (四分位竞争力)" + ` 和 ` + "`homogeneityIndex` (内部均衡度)" + `，为每个班级打上“画像标签”，例如：
    * “(1)班：高分领跑型” (高分位竞争力强，但可能内部均衡度不高)
    * “(2)班：基础扎实型” (中低分位竞争力强，及格率有保障)
    * “(3)班：整体均衡型” (各项指标接近年级平均，内部均衡度好)
3.  **聚焦差异**: 重点分析班级之间差异最大的指标，并推测可能的原因（如教学风格、班级管理等）。
# 输入数据: (包含 groupStats 和所有班级的 tableStats)
` + "```json\n%s\n```"

	PromptSingleClass = `
# 角色: 经验丰富的班主任和诊断专家
# 任务: 为指定的这一个班级，提供一份详细、可落地的深度诊断报告。
# 核心指令:
1.  **自我定位**: 首先，将本班的各项核心指标（平均分、优秀率、标准差等）与年级平均水平进行全面对比，明确本班在年级中的位置。
2.  **内部问题诊断**:
    * 分析本班的 ` + "`highAchieverPenetration`" + ` 和 ` + "`strugglerSupportIndex`" + `，判断班级的优势是在于“拔尖”还是“兜底”。
    * 解读 ` + "`stdDev`" + ` 和 ` + "`homogeneityIndex`" + `，评估班内学业分化的严重程度。
3.  **提出针对性建议**:
    * 如果分化严重，提出分层教学或“一生一策”辅导的具体建议。
    * 如果高分层薄弱，提出如何“拔尖”的策略。
    * 如果后进生问题突出，提出如何“补差”的方案。
4.  **建议必须具体、可行，直接面向班主任的日常工作。
# 输入数据: (仅包含指定班级的 tableStats 和年级的 groupStats 作为对比基准)
` + "```json\n%s\n```"
)

// AIResult 定义了最终AI分析的JSON结构
type AIResult struct {
	Summary     string            `json:"summary"`
	Comparison  string            `json:"comparison"`
	Diagnostics map[string]string `json:"diagnostics"`
}

// AIAnalyzer 负责执行AI分析
type AIAnalyzer struct {
	Repo       *repository.Repository
	LLMService *LLMService
}

// NewAIAnalyzer 创建 AIAnalyzer 实例
func NewAIAnalyzer(repo *repository.Repository, llmService *LLMService) *AIAnalyzer {
	return &AIAnalyzer{Repo: repo, LLMService: llmService}
}

// getOrGenerateAIAnalysis 实现了分步生成AI分析的逻辑
func (a *AIAnalyzer) getOrGenerateAIAnalysis(reportID uint) (string, error) {
	// 获取报告并进行前置检查
	report, err := a.Repo.GetReportByID(reportID)
	if err != nil {
		return "", fmt.Errorf("获取报告失败: %w", err)
	}
	if report.Status != "completed" {
		return "", fmt.Errorf("主报告尚未完成，无法进行AI分析")
	}
	if report.AIAnalysisStatus == "completed" && report.AIAnalysisCache != "" {
		return report.AIAnalysisCache, nil // 直接返回缓存
	}
	if report.FullReportData == "" {
		return "", fmt.Errorf("报告数据为空，无法分析")
	}

	var fullReportData map[string]interface{}
	if err := json.Unmarshal([]byte(report.FullReportData), &fullReportData); err != nil {
		return "", fmt.Errorf("解析报告JSON数据失败: %w", err)
	}

	// 初始化并发控制和结果存储
	var wg sync.WaitGroup
	var mu sync.Mutex
	aiResult := AIResult{Diagnostics: make(map[string]string)}

	// 提前计算任务总数，为错误channel设置合适的缓冲区大小
	tables, _ := fullReportData["tables"].([]interface{})
	taskCount := 2 + len(tables) // 1个摘要, 1个对比, N个班级诊断
	errChan := make(chan error, taskCount)

	// 并发执行AI分析任务

	// 生成年级摘要
	wg.Add(1)
	go func() {
		defer wg.Done()
		summaryData := map[string]interface{}{"groupStats": fullReportData["groupStats"], "fullMarks": fullReportData["fullMarks"]}
		summaryJSON, _ := json.MarshalIndent(summaryData, "", "  ")
		summaryPrompt := fmt.Sprintf(PromptSummary, string(summaryJSON))

		summary, err := a.LLMService.GetCompletion(summaryPrompt)
		if err != nil {
			errChan <- fmt.Errorf("生成年级摘要失败: %w", err)
			return
		}
		mu.Lock()
		aiResult.Summary = summary
		mu.Unlock()
	}()

	// 生成班级横向对比
	wg.Add(1)
	go func() {
		defer wg.Done()
		comparisonData := map[string]interface{}{"groupStats": fullReportData["groupStats"], "tables": fullReportData["tables"]}
		comparisonJSON, _ := json.MarshalIndent(comparisonData, "", "  ")
		comparisonPrompt := fmt.Sprintf(PromptComparison, string(comparisonJSON))

		comparison, err := a.LLMService.GetCompletion(comparisonPrompt)
		if err != nil {
			errChan <- fmt.Errorf("生成班级横向对比失败: %w", err)
			return
		}
		mu.Lock()
		aiResult.Comparison = comparison
		mu.Unlock()
	}()

	// 为每个班级生成深度诊断
	if tables != nil {
		for _, t := range tables {
			// 必须在循环内创建局部变量，否则goroutine会捕获到错误的t
			tableData := t.(map[string]interface{})
			wg.Add(1)
			go func(currentTable map[string]interface{}) {
				defer wg.Done()
				className, ok := currentTable["tableName"].(string)
				if !ok {
					errChan <- fmt.Errorf("班级名称格式错误")
					return
				}

				singleClassData := map[string]interface{}{
					"className":  className,
					"classStats": currentTable["tableStats"],
					"gradeStats": fullReportData["groupStats"],
				}
				singleClassJSON, _ := json.MarshalIndent(singleClassData, "", "  ")
				singleClassPrompt := fmt.Sprintf(PromptSingleClass, string(singleClassJSON))

				diagnostics, err := a.LLMService.GetCompletion(singleClassPrompt)
				if err != nil {
					errChan <- fmt.Errorf("为班级 '%s' 生成诊断报告失败: %w", className, err)
					return
				}
				mu.Lock()
				aiResult.Diagnostics[className] = diagnostics
				mu.Unlock()
			}(tableData)
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 阻塞并从channel中读取错误，只要有一个错误就立即返回
	for err := range errChan {
		if err != nil {
			// 返回第一个遇到的错误
			return "", err
		}
	}

	// 步骤 5: 将最终结果序列化并保存到数据库
	finalResultBytes, err := json.Marshal(aiResult)
	if err != nil {
		return "", fmt.Errorf("序列化最终AI结果失败: %w", err)
	}

	report.AIAnalysisCache = string(finalResultBytes)
	report.AIAnalysisStatus = "completed"
	if err := a.Repo.UpdateAnalysisReport(report); err != nil {
		return "", fmt.Errorf("更新AI分析结果到数据库失败: %w", err)
	}

	return report.AIAnalysisCache, nil
}
