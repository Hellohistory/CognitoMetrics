// file: internal/analyzer/processor/processor.go
package processor

import (
	"CognitoMetrics/internal/analyzer/stats"
	"CognitoMetrics/internal/analyzer/types"
	"sort"

	"gonum.org/v1/gonum/stat"
)

func ProcessRanks(allStudents []*types.StudentInput, subjects []string, studentReportMap map[uint]*types.StudentReport) {
	gradeCount := float64(len(allStudents))

	// --- 年级总分排名 ---
	sort.SliceStable(allStudents, func(i, j int) bool { return allStudents[i].TotalScore > allStudents[j].TotalScore })
	for i, student := range allStudents {
		rank := i + 1
		if i > 0 && student.TotalScore == allStudents[i-1].TotalScore {
			rank = studentReportMap[allStudents[i-1].ID].GradeRank
		}
		studentReportMap[student.ID] = &types.StudentReport{
			ID:          student.ID,
			StudentName: student.StudentName,
			TableName:   student.TableName,
			TotalScore:  stats.Round(student.TotalScore, 2),
			GradeRank:   rank,
			Ranks: types.StudentRanks{
				TotalScore: types.RankInfo{GradeRank: rank, GradePercentileRank: stats.Round((gradeCount-float64(rank)+1)/gradeCount*100, 2)},
				Subjects:   make(map[string]types.RankInfo),
			},
			Scores:  types.StudentScores{RawScores: student.Scores, ZScores: make(map[string]float64), TScores: make(map[string]float64), ScoreRates: make(map[string]float64)},
			Metrics: types.StudentMetrics{StrengthSubjects: []types.SubjectTScore{}, WeaknessSubjects: []types.SubjectTScore{}, ContributionScore: make(map[string]float64)},
		}
	}

	// --- 年级单科排名 ---
	for _, subject := range subjects {
		sort.SliceStable(allStudents, func(i, j int) bool { return allStudents[i].Scores[subject] > allStudents[j].Scores[subject] })
		for i, student := range allStudents {
			rank := i + 1
			if i > 0 && student.Scores[subject] == allStudents[i-1].Scores[subject] {
				rank = studentReportMap[allStudents[i-1].ID].Ranks.Subjects[subject].GradeRank
			}
			r := studentReportMap[student.ID].Ranks.Subjects[subject]
			r.GradeRank = rank
			r.GradePercentileRank = stats.Round((gradeCount-float64(rank)+1)/gradeCount*100, 2)
			studentReportMap[student.ID].Ranks.Subjects[subject] = r
		}
	}

	// 按班级对学生进行分组
	classStudentsMap := make(map[string][]*types.StudentInput)
	for _, s := range allStudents {
		classStudentsMap[s.TableName] = append(classStudentsMap[s.TableName], s)
	}

	// --- 班级总分排名 ---
	for _, studentsInClass := range classStudentsMap {
		classCount := float64(len(studentsInClass))
		// 按总分对班级内学生排序
		sort.SliceStable(studentsInClass, func(i, j int) bool { return studentsInClass[i].TotalScore > studentsInClass[j].TotalScore })

		for i, student := range studentsInClass {
			rank := i + 1
			// 处理同分同名次
			if i > 0 && student.TotalScore == studentsInClass[i-1].TotalScore {
				rank = studentReportMap[studentsInClass[i-1].ID].ClassRank
			}

			// 更新学生报告中的班级排名信息
			studentReportMap[student.ID].ClassRank = rank
			r := studentReportMap[student.ID].Ranks.TotalScore
			r.ClassRank = rank
			r.ClassPercentileRank = stats.Round((classCount-float64(rank)+1)/classCount*100, 2)
			studentReportMap[student.ID].Ranks.TotalScore = r
		}
	}

	// --- 班级内单科排名 ---
	for _, subject := range subjects {
		for _, studentsInClass := range classStudentsMap {
			classCount := float64(len(studentsInClass))
			// 按当前学科分数对班级内学生排序
			sort.SliceStable(studentsInClass, func(i, j int) bool {
				return studentsInClass[i].Scores[subject] > studentsInClass[j].Scores[subject]
			})

			for i, student := range studentsInClass {
				rank := i + 1
				// 处理同分同名次
				if i > 0 && student.Scores[subject] == studentsInClass[i-1].Scores[subject] {
					rank = studentReportMap[studentsInClass[i-1].ID].Ranks.Subjects[subject].ClassRank
				}
				// 更新学生报告中的班级学科排名信息
				r := studentReportMap[student.ID].Ranks.Subjects[subject]
				r.ClassRank = rank
				r.ClassPercentileRank = stats.Round((classCount-float64(rank)+1)/classCount*100, 2)
				studentReportMap[student.ID].Ranks.Subjects[subject] = r
			}
		}
	}
}

// --- 新增函数 START ---
// ProcessGroupStats 计算年级/群体的整体统计数据
func ProcessGroupStats(allStudents []*types.StudentInput, subjects []string, fullMarks map[string]float64, totalFullMarks float64) *types.LevelStats {
	gradeStats := &types.LevelStats{
		StatsBySubject:    make(map[string]*types.SubjectStats),
		CorrelationMatrix: make(map[string]map[string]float64),
	}
	groupScoresBySubject := make(map[string][]float64)

	// 1. 提取所有学科和总分的原始分数列表
	for _, subject := range subjects {
		scores := make([]float64, len(allStudents))
		for i, s := range allStudents {
			scores[i] = s.Scores[subject]
		}
		groupScoresBySubject[subject] = scores
	}
	totalScores := make([]float64, len(allStudents))
	for i, s := range allStudents {
		totalScores[i] = s.TotalScore
	}
	groupScoresBySubject["totalScore"] = totalScores

	// 2. 计算各科和总分的描述性统计量
	allSubjectsAndTotal := append(subjects, "totalScore")
	for _, subject := range allSubjectsAndTotal {
		scores := groupScoresBySubject[subject]
		currentFullMark := fullMarks[subject]
		if subject == "totalScore" {
			currentFullMark = totalFullMarks
		}

		subjectStats := stats.CalculateDescriptiveStats(scores, currentFullMark)
		subjectStats.DiscriminationIndex = stats.CalculateDiscriminationIndex(scores, currentFullMark)
		stats.CalculateAdvancedGroupMetrics(scores, subjectStats)
		subjectStats.RawScores = scores // 缓存原始分数，供后续班级竞争力计算使用
		gradeStats.StatsBySubject[subject] = subjectStats
	}

	// 3. 计算学科间的相关性矩阵
	for _, s1 := range subjects {
		gradeStats.CorrelationMatrix[s1] = make(map[string]float64)
		for _, s2 := range subjects {
			if s1 == s2 {
				gradeStats.CorrelationMatrix[s1][s2] = 1.0
			} else {
				scores1 := groupScoresBySubject[s1]
				scores2 := groupScoresBySubject[s2]
				correlation := stats.CalculateCorrelation(scores1, scores2)
				gradeStats.CorrelationMatrix[s1][s2] = correlation
			}
		}
	}

	return gradeStats
}

// --- 新增函数 END ---

func ProcessSingleClass(classInput *types.ClassInputData, studentsInClass []*types.StudentInput, gradeStats *types.LevelStats, studentReportMap map[uint]*types.StudentReport, subjects []string, fullMarks map[string]float64, totalFullMarks float64, historyMap map[uint]*types.StudentHistory) *types.ClassReport {
	classScoresBySubject := make(map[string][]float64)
	for _, subject := range subjects {
		scores := make([]float64, len(studentsInClass))
		for i, s := range studentsInClass {
			scores[i] = s.Scores[subject]
		}
		classScoresBySubject[subject] = scores
	}

	classStats := &types.LevelStats{StatsBySubject: make(map[string]*types.SubjectStats)}
	allSubjectsAndTotal := append(subjects, "totalScore")

	// 计算班级总分列表
	classTotalScores := make([]float64, len(studentsInClass))
	for i, s := range studentsInClass {
		classTotalScores[i] = s.TotalScore
	}
	classScoresBySubject["totalScore"] = classTotalScores

	// 计算班级各科统计指标
	for _, subject := range allSubjectsAndTotal {
		scores := classScoresBySubject[subject]
		currentFullMark := fullMarks[subject]
		if subject == "totalScore" {
			currentFullMark = totalFullMarks
		}

		subjectStats := stats.CalculateDescriptiveStats(scores, currentFullMark)
		stats.CalculateAdvancedGroupMetrics(scores, subjectStats)

		// 计算班级 vs 年级的对比性指标
		if gradeStats.StatsBySubject[subject] != nil && gradeStats.StatsBySubject[subject].StdDev > 0 {
			subjectStats.HomogeneityIndex = stats.Round(subjectStats.StdDev/gradeStats.StatsBySubject[subject].StdDev, 3)
		} else {
			subjectStats.HomogeneityIndex = 1.0
		}
		gradeRawScores := gradeStats.StatsBySubject[subject].RawScores
		if len(gradeRawScores) > 0 {
			// 这里需要对年级分数进行排序，以用于计算CDF
			// 注意：为了不修改原始的gradeStats，我们复制一份
			sortedGradeRawScores := make([]float64, len(gradeRawScores))
			copy(sortedGradeRawScores, gradeRawScores)
			sort.Float64s(sortedGradeRawScores)

			subjectStats.QuartileCompetitiveness = make(map[string]float64)
			qc := subjectStats.QuartileCompetitiveness
			qc["q1"] = stats.Round(stat.CDF(subjectStats.Q1, stat.Empirical, sortedGradeRawScores, nil)*100, 2)
			qc["median"] = stats.Round(stat.CDF(subjectStats.Median, stat.Empirical, sortedGradeRawScores, nil)*100, 2)
			qc["q3"] = stats.Round(stat.CDF(subjectStats.Q3, stat.Empirical, sortedGradeRawScores, nil)*100, 2)
		}

		classStats.StatsBySubject[subject] = subjectStats
	}

	var studentReportsForClass []*types.StudentReport
	const PASS_THRESHOLD = 0.60
	const EXCELLENT_THRESHOLD = 0.85
	passScoreLine := totalFullMarks * PASS_THRESHOLD
	excellentScoreLine := totalFullMarks * EXCELLENT_THRESHOLD

	for _, student := range studentsInClass {
		report := studentReportMap[student.ID]
		var studentTScoreValues []float64
		for _, subj := range subjects {
			gradeMean := gradeStats.StatsBySubject[subj].Mean
			gradeStdDev := gradeStats.StatsBySubject[subj].StdDev
			var zScore, tScore float64
			if gradeStdDev != 0 {
				zScore = (report.Scores.RawScores[subj] - gradeMean) / gradeStdDev
				tScore = 50.0 + 10*zScore
			} else {
				zScore, tScore = 0.0, 50.0
			}
			report.Scores.ZScores[subj] = stats.Round(zScore, 3)
			report.Scores.TScores[subj] = stats.Round(tScore, 2)
			if fullMarks[subj] > 0 {
				report.Scores.ScoreRates[subj] = stats.Round(report.Scores.RawScores[subj]/fullMarks[subj], 3)
			}
			studentTScoreValues = append(studentTScoreValues, tScore)
		}
		if totalFullMarks > 0 {
			report.Scores.ScoreRates["totalScore"] = stats.Round(report.TotalScore/totalFullMarks, 3)
		}

		// 计算总分T-Score
		if gradeStats.StatsBySubject["totalScore"] != nil && gradeStats.StatsBySubject["totalScore"].StdDev != 0 {
			totalTScore := 50.0 + 10*((report.TotalScore-gradeStats.StatsBySubject["totalScore"].Mean)/gradeStats.StatsBySubject["totalScore"].StdDev)
			report.Scores.TScores["totalScore"] = stats.Round(totalTScore, 2)
		} else {
			report.Scores.TScores["totalScore"] = 50.0
		}

		totalTScore := report.Scores.TScores["totalScore"]

		if len(studentTScoreValues) > 1 {
			report.Metrics.ImbalanceIndex = stats.Round(stat.StdDev(studentTScoreValues, nil), 2)
		}

		type subjectTScorePair struct {
			name   string
			tScore float64
		}
		tScorePairs := make([]subjectTScorePair, 0, len(subjects))
		for _, subj := range subjects {
			tScorePairs = append(tScorePairs, subjectTScorePair{name: subj, tScore: report.Scores.TScores[subj]})
		}
		sort.Slice(tScorePairs, func(i, j int) bool { return tScorePairs[i].tScore > tScorePairs[j].tScore })
		if len(tScorePairs) > 0 {
			report.Metrics.StrengthSubjects = []types.SubjectTScore{{Subject: tScorePairs[0].name, TScore: stats.Round(tScorePairs[0].tScore, 2)}}
			report.Metrics.WeaknessSubjects = []types.SubjectTScore{{Subject: tScorePairs[len(tScorePairs)-1].name, TScore: stats.Round(tScorePairs[len(tScorePairs)-1].tScore, 2)}}
		}

		profile := "潜力提升型"
		if totalTScore >= 62 {
			profile = "拔尖均衡型"
			if report.Metrics.ImbalanceIndex >= 12 {
				profile = "拔尖偏科型"
			}
		} else if totalTScore >= 55 {
			profile = "稳健发展型"
		} else if totalTScore >= 45 {
			profile = "中坚力量型"
		} else {
			profile = "基础薄弱型"
			if report.Metrics.ImbalanceIndex >= 12 {
				profile = "短板制约型"
			}
		}
		report.Profile = profile

		if report.TotalScore < passScoreLine {
			report.Metrics.PointsToPass = stats.Round(passScoreLine-report.TotalScore, 2)
		}
		if report.TotalScore < excellentScoreLine {
			report.Metrics.PointsToExcellent = stats.Round(excellentScoreLine-report.TotalScore, 2)
		}

		stats.CalculateAdvancedStudentMetrics(report, classScoresBySubject)

		if historyMap != nil {
			if studentHistory, ok := historyMap[student.ID]; ok && len(studentHistory.AllExams) >= 2 {
				report.Metrics.History = &types.HistoricalMetrics{Trend: make(map[string]interface{}), Stability: make(map[string]interface{})}
				if studentHistory.LastExam != nil {
					report.Metrics.History.Trend["totalScore"] = stats.Round(report.TotalScore-studentHistory.LastExam.TotalScore, 2)
				}

				historicalRanks := make([]float64, len(studentHistory.AllExams))
				historicalTScores := make([]float64, len(studentHistory.AllExams))
				for i, exam := range studentHistory.AllExams {
					historicalRanks[i] = exam.GradePercentileRank
					historicalTScores[i] = exam.TotalTScore
				}
				currentPercentileRank, _ := report.Ranks.TotalScore.GradePercentileRank, 0.0
				historicalRanks = append(historicalRanks, currentPercentileRank)
				historicalTScores = append(historicalTScores, report.Scores.TScores["totalScore"])

				if len(historicalRanks) > 1 {
					report.Metrics.History.GradePercentileRankVolatility = stats.Round(stat.StdDev(historicalRanks, nil), 2)
					report.Metrics.History.GradePercentileRankSlope = stats.AnalyzeTrendSlope(historicalRanks)
				}
				if len(historicalTScores) > 1 {
					report.Metrics.History.TotalTScoreVolatility = stats.Round(stat.StdDev(historicalTScores, nil), 2)
				}
			}
		}

		studentReportsForClass = append(studentReportsForClass, report)
	}
	sort.Slice(studentReportsForClass, func(i, j int) bool { return studentReportsForClass[i].ClassRank < studentReportsForClass[j].ClassRank })

	return &types.ClassReport{
		TableName:  classInput.TableName,
		TableStats: classStats,
		Students:   studentReportsForClass,
	}
}
