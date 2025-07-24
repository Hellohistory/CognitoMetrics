package processor

import (
	"CognitoMetrics/internal/analyzer/stats"
	"CognitoMetrics/internal/analyzer/types"
	"gonum.org/v1/gonum/stat"
	"sort"
)

func ProcessRanks(allStudents []*types.StudentInput, subjects []string, studentReportMap map[string]*types.StudentReport) {
	gradeCount := float64(len(allStudents))

	sort.SliceStable(allStudents, func(i, j int) bool { return allStudents[i].TotalScore > allStudents[j].TotalScore })
	for i, student := range allStudents {
		rank := i + 1
		if i > 0 && student.TotalScore == allStudents[i-1].TotalScore {
			rank = studentReportMap[allStudents[i-1].StudentName].GradeRank
		}
		studentReportMap[student.StudentName] = &types.StudentReport{
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

	for _, subject := range subjects {
		sort.SliceStable(allStudents, func(i, j int) bool { return allStudents[i].Scores[subject] > allStudents[j].Scores[subject] })
		for i, student := range allStudents {
			rank := i + 1
			if i > 0 && student.Scores[subject] == allStudents[i-1].Scores[subject] {
				rank = studentReportMap[allStudents[i-1].StudentName].Ranks.Subjects[subject].GradeRank
			}
			r := studentReportMap[student.StudentName].Ranks.Subjects[subject]
			r.GradeRank = rank
			r.GradePercentileRank = stats.Round((gradeCount-float64(rank)+1)/gradeCount*100, 2)
			studentReportMap[student.StudentName].Ranks.Subjects[subject] = r
		}
	}

	classStudentsMap := make(map[string][]*types.StudentInput)
	for _, s := range allStudents {
		classStudentsMap[s.TableName] = append(classStudentsMap[s.TableName], s)
	}
	for _, studentsInClass := range classStudentsMap {
		classCount := float64(len(studentsInClass))
		// 按总分对班级内学生排序
		sort.SliceStable(studentsInClass, func(i, j int) bool { return studentsInClass[i].TotalScore > studentsInClass[j].TotalScore })

		for i, student := range studentsInClass {
			rank := i + 1
			// 处理同分同名次
			if i > 0 && student.TotalScore == studentsInClass[i-1].TotalScore {
				rank = studentReportMap[studentsInClass[i-1].StudentName].ClassRank
			}

			// 更新学生报告中的班级排名信息
			studentReportMap[student.StudentName].ClassRank = rank
			r := studentReportMap[student.StudentName].Ranks.TotalScore
			r.ClassRank = rank
			r.ClassPercentileRank = stats.Round((classCount-float64(rank)+1)/classCount*100, 2)
			studentReportMap[student.StudentName].Ranks.TotalScore = r
		}
	}
}

func ProcessGroupStats(allStudents []*types.StudentInput, subjects []string, fullMarks map[string]float64, totalFullMarks float64) *types.LevelStats {
	statsBySubject := make(map[string]*types.SubjectStats)
	// Omitting goroutines here for simplicity and to ensure map safety without mutexes
	for _, subject := range subjects {
		scores := make([]float64, len(allStudents))
		for i, s := range allStudents {
			scores[i] = s.Scores[subject]
		}
		subjectStat := stats.CalculateDescriptiveStats(scores, fullMarks[subject])
		subjectStat.DiscriminationIndex = stats.CalculateDiscriminationIndex(scores, fullMarks[subject])
		stats.CalculateAdvancedGroupMetrics(scores, subjectStat)
		subjectStat.RawScores = scores
		statsBySubject[subject] = subjectStat
	}

	totalScores := make([]float64, len(allStudents))
	for i, s := range allStudents {
		totalScores[i] = s.TotalScore
	}
	totalScoreStats := stats.CalculateDescriptiveStats(totalScores, totalFullMarks)
	totalScoreStats.DiscriminationIndex = stats.CalculateDiscriminationIndex(totalScores, totalFullMarks)
	stats.CalculateAdvancedGroupMetrics(totalScores, totalScoreStats)
	totalScoreStats.RawScores = totalScores
	statsBySubject["totalScore"] = totalScoreStats

	correlationMatrix := make(map[string]map[string]float64)
	for _, s1 := range subjects {
		correlationMatrix[s1] = make(map[string]float64)
		for _, s2 := range subjects {
			if s1 == s2 {
				correlationMatrix[s1][s2] = 1.0
				continue
			}
			correlationMatrix[s1][s2] = stats.CalculateCorrelation(statsBySubject[s1].RawScores, statsBySubject[s2].RawScores)
		}
	}
	return &types.LevelStats{
		StatsBySubject:    statsBySubject,
		CorrelationMatrix: correlationMatrix,
	}
}

func ProcessSingleClass(classInput *types.ClassInputData, studentsInClass []*types.StudentInput, gradeStats *types.LevelStats, studentReportMap map[string]*types.StudentReport, subjects []string, fullMarks map[string]float64, totalFullMarks float64, historyMap map[string]*types.StudentHistory) *types.ClassReport {
	classScoresBySubject := make(map[string][]float64)
	for _, subject := range subjects {
		scores := make([]float64, len(studentsInClass))
		for i, s := range studentsInClass {
			scores[i] = s.Scores[subject]
		}
		classScoresBySubject[subject] = scores
	}

	classStats := &types.LevelStats{StatsBySubject: make(map[string]*types.SubjectStats)}
	for _, subject := range subjects {
		scores := classScoresBySubject[subject]
		subjectStats := stats.CalculateDescriptiveStats(scores, fullMarks[subject])
		stats.CalculateAdvancedGroupMetrics(scores, subjectStats)
		if gradeStats.StatsBySubject[subject].StdDev > 0 {
			subjectStats.HomogeneityIndex = stats.Round(subjectStats.StdDev/gradeStats.StatsBySubject[subject].StdDev, 3)
		} else {
			subjectStats.HomogeneityIndex = 1.0
		}

		// 计算四分位竞争力 (Quartile Competitiveness)
		gradeRawScores := gradeStats.StatsBySubject[subject].RawScores
		if len(gradeRawScores) > 0 {
			// 在调用 CDF 之前对数据进行排序
			sort.Float64s(gradeRawScores)

			subjectStats.QuartileCompetitiveness = make(map[string]float64)
			qc := subjectStats.QuartileCompetitiveness
			qc["q1"] = stats.Round(stat.CDF(subjectStats.Q1, stat.Empirical, gradeRawScores, nil)*100, 2)
			qc["median"] = stats.Round(stat.CDF(subjectStats.Median, stat.Empirical, gradeRawScores, nil)*100, 2)
			qc["q3"] = stats.Round(stat.CDF(subjectStats.Q3, stat.Empirical, gradeRawScores, nil)*100, 2)
		}

		classStats.StatsBySubject[subject] = subjectStats
	}

	var studentReportsForClass []*types.StudentReport
	for _, student := range studentsInClass {
		report := studentReportMap[student.StudentName]
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
			studentTScoreValues = append(studentTScoreValues, tScore)
		}
		totalTScore := 50.0 + 10*((report.TotalScore-gradeStats.StatsBySubject["totalScore"].Mean)/gradeStats.StatsBySubject["totalScore"].StdDev)
		report.Scores.TScores["totalScore"] = stats.Round(totalTScore, 2)
		report.Metrics.ImbalanceIndex = stats.Round(stat.StdDev(studentTScoreValues, nil), 2)

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
			report.Metrics.StrengthSubjects = []types.SubjectTScore{{Subject: tScorePairs[0].name, TScore: tScorePairs[0].tScore}}
			report.Metrics.WeaknessSubjects = []types.SubjectTScore{{Subject: tScorePairs[len(tScorePairs)-1].name, TScore: tScorePairs[len(tScorePairs)-1].tScore}}
		}

		profile := "潜力提升型" // Default profile
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

		stats.CalculateAdvancedStudentMetrics(report, classScoresBySubject)

		if historyMap != nil {
			if studentHistory, ok := historyMap[student.StudentName]; ok && len(studentHistory.AllExams) >= 2 {
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
				historicalRanks = append(historicalRanks, report.Ranks.TotalScore.GradePercentileRank)
				historicalTScores = append(historicalTScores, report.Scores.TScores["totalScore"])

				report.Metrics.History.GradePercentileRankVolatility = stats.Round(stat.StdDev(historicalRanks, nil), 2)
				report.Metrics.History.TotalTScoreVolatility = stats.Round(stat.StdDev(historicalTScores, nil), 2)
				report.Metrics.History.GradePercentileRankSlope = stats.AnalyzeTrendSlope(historicalRanks)
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
