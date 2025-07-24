package analyzer

import (
	"CognitoMetrics/internal/analyzer/processor"
	"CognitoMetrics/internal/analyzer/types"
	"CognitoMetrics/internal/repository"
	"fmt"
	"sort"
	"sync"
)

func PerformAnalysis(input *types.AnalysisInput, historyMap map[string]*types.StudentHistory, repo *repository.Repository) (*types.AnalysisReport, error) {
	if input == nil || len(input.Tables) == 0 {
		return nil, fmt.Errorf("input data is nil or empty")
	}

	subjects := make([]string, 0, len(input.FullMarks))
	var totalFullMarks float64
	for subject, mark := range input.FullMarks {
		subjects = append(subjects, subject)
		totalFullMarks += mark
	}

	report := &types.AnalysisReport{
		GroupName: input.GroupName,
		FullMarks: input.FullMarks,
		Tables:    []*types.ClassReport{},
	}

	var allStudents []*types.StudentInput
	for _, classTable := range input.Tables {
		for _, student := range classTable.Students {
			student.TableName = classTable.TableName
			var studentTotalScore float64
			for _, subjectName := range subjects {
				studentTotalScore += student.Scores[subjectName]
			}
			student.TotalScore = studentTotalScore
			allStudents = append(allStudents, student)
		}
	}
	if len(allStudents) == 0 {
		report.Error = "No student data found for analysis."
		return report, nil
	}

	studentReportMap := make(map[string]*types.StudentReport, len(allStudents))
	processor.ProcessRanks(allStudents, subjects, studentReportMap)

	gradeStats := processor.ProcessGroupStats(allStudents, subjects, input.FullMarks, totalFullMarks)
	report.GroupStats = gradeStats

	var wg sync.WaitGroup
	var mu sync.Mutex
	classStudentsMap := make(map[string][]*types.StudentInput)
	for _, s := range allStudents {
		classStudentsMap[s.TableName] = append(classStudentsMap[s.TableName], s)
	}

	for _, classInput := range input.Tables {
		wg.Add(1)
		go func(ci *types.ClassInputData) {
			defer wg.Done()
			classReport := processor.ProcessSingleClass(ci, classStudentsMap[ci.TableName], gradeStats, studentReportMap, subjects, input.FullMarks, totalFullMarks, historyMap)
			mu.Lock()
			report.Tables = append(report.Tables, classReport)
			mu.Unlock()
		}(classInput)
	}
	wg.Wait()

	sort.Slice(report.Tables, func(i, j int) bool { return report.Tables[i].TableName < report.Tables[j].TableName })

	// 异步回写结果，不阻塞主流程
	go repo.UpdateScoresWithMetrics(report, input.ExamID)

	return report, nil
}
