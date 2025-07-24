// in: internal/charts/generator.go

package charts

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"CognitoMetrics/internal/analyzer/types"
)

// GenerateChartData 将详细报告数据二次处理，转换为适合前端图表库使用的格式
func GenerateChartData(report *types.AnalysisReport) (*ChartData, error) {
	chartData := &ChartData{}

	// --- 初始化 ---
	groupStats := report.GroupStats
	tables := report.Tables
	if groupStats == nil || len(tables) == 0 {
		return chartData, nil // 如果缺少核心数据，直接返回空结构
	}

	// 提取基础信息
	var subjects []string
	for sub := range groupStats.CorrelationMatrix {
		subjects = append(subjects, sub)
	}
	sort.Strings(subjects) // 保证顺序

	var classNames []string
	for _, t := range tables {
		classNames = append(classNames, t.TableName)
	}

	// --- 年级整体图表生成 ---
	gradeCharts := &chartData.GradeLevelCharts
	gradeCharts.ScoreDistributionHistogram = make(map[string]HistogramData)

	// 分数分布直方图
	allSubjectsAndTotal := append(subjects, "totalScore")
	for _, subject := range allSubjectsAndTotal {
		if stats, ok := groupStats.StatsBySubject[subject]; ok {
			freqDist := stats.FrequencyDistribution
			keys := make([]string, 0, len(freqDist))
			for k := range freqDist {
				keys = append(keys, k)
			}
			// 重要：对分数段进行数值排序
			sort.Slice(keys, func(i, j int) bool {
				numA, _ := strconv.Atoi(strings.Split(keys[i], "-")[0])
				numB, _ := strconv.Atoi(strings.Split(keys[j], "-")[0])
				return numA < numB
			})

			seriesData := make([]int, len(keys))
			for i, k := range keys {
				seriesData[i] = freqDist[k]
			}

			gradeCharts.ScoreDistributionHistogram[subject] = HistogramData{
				Categories: keys,
				SeriesData: seriesData,
				SeriesName: fmt.Sprintf("年级%s分数分布", subject),
			}
		}
	}

	// 相关性热力图
	heatmapData := make([][]interface{}, 0)
	for i, s1 := range subjects {
		for j, s2 := range subjects {
			if corr, ok := groupStats.CorrelationMatrix[s1][s2]; ok {
				heatmapData = append(heatmapData, []interface{}{j, i, corr})
			}
		}
	}
	gradeCharts.SubjectCorrelationHeatmap = HeatmapData{
		XAxisLabels: subjects,
		YAxisLabels: subjects,
		Data:        heatmapData,
		Title:       "学科成绩相关性热力图",
	}

	// 难度-区分度散点图
	diffScatterData := make([][]interface{}, 0)
	for _, subject := range subjects {
		if stats, ok := groupStats.StatsBySubject[subject]; ok {
			diffScatterData = append(diffScatterData, []interface{}{stats.Difficulty, stats.DiscriminationIndex, subject})
		}
	}
	gradeCharts.SubjectDifficultyScatter = ScatterPlotData{
		Data:      diffScatterData,
		XAxisName: "难度",
		YAxisName: "区分度",
		Title:     "学科难度-区分度分析",
	}

	// --- 班级对比图表生成 ---
	classCharts := &chartData.ClassComparisonCharts
	classCharts.MetricsBarChart = make(map[string]map[string]BarChartData)
	metricsToCompare := []string{"mean", "passRate", "excellentRate", "highAchieverPenetration", "academicCoreDensity"}

	// 指标对比柱状图
	for _, metric := range metricsToCompare {
		classCharts.MetricsBarChart[metric] = make(map[string]BarChartData)
		for _, subject := range allSubjectsAndTotal {
			seriesData := make([]float64, len(tables))
			for i, table := range tables {
				if stats, ok := table.TableStats.StatsBySubject[subject]; ok {
					switch metric {
					case "mean":
						seriesData[i] = stats.Mean
					case "passRate":
						seriesData[i] = stats.PassRate
					case "excellentRate":
						seriesData[i] = stats.ExcellentRate
					case "highAchieverPenetration":
						seriesData[i] = stats.HighAchieverPenetration
					case "academicCoreDensity":
						seriesData[i] = stats.AcademicCoreDensity
					}
				}
			}
			// 添加年级平均
			seriesDataWithGrade := seriesData
			if stats, ok := groupStats.StatsBySubject[subject]; ok {
				var gradeMetricValue float64
				switch metric {
				case "mean":
					gradeMetricValue = stats.Mean
				case "passRate":
					gradeMetricValue = stats.PassRate
				case "excellentRate":
					gradeMetricValue = stats.ExcellentRate
				case "highAchieverPenetration":
					gradeMetricValue = stats.HighAchieverPenetration
				case "academicCoreDensity":
					gradeMetricValue = stats.AcademicCoreDensity
				}
				seriesDataWithGrade = append(seriesDataWithGrade, gradeMetricValue)
			}

			classCharts.MetricsBarChart[metric][subject] = BarChartData{
				Categories: append(classNames, "年级平均"),
				SeriesData: seriesDataWithGrade,
				SeriesName: fmt.Sprintf("%s - %s", subject, metric),
			}
		}
	}

	// 分数分布箱线图
	classCharts.ScoreDistributionBoxplot = make(map[string]BoxplotData)
	for _, subject := range allSubjectsAndTotal {
		seriesData := make([][]float64, len(tables))
		for i, table := range tables {
			if stats, ok := table.TableStats.StatsBySubject[subject]; ok {
				bp := stats.BoxPlotData
				seriesData[i] = []float64{bp["min"], bp["q1"], bp["median"], bp["q3"], bp["max"]}
			}
		}
		if stats, ok := groupStats.StatsBySubject[subject]; ok {
			bp := stats.BoxPlotData
			seriesData = append(seriesData, []float64{bp["min"], bp["q1"], bp["median"], bp["q3"], bp["max"]})
		}

		classCharts.ScoreDistributionBoxplot[subject] = BoxplotData{
			Categories: append(classNames, "年级整体"),
			Data:       seriesData,
			Title:      fmt.Sprintf("%s 成绩分布箱线图", subject),
		}
	}

	// 班级画像雷达图
	classCharts.ClassProfileRadar = make(map[string]RadarChartData)
	radarIndicator := make([]RadarIndicator, len(subjects))
	for i, s := range subjects {
		radarIndicator[i] = RadarIndicator{Name: s, Max: report.FullMarks[s]}
	}
	gradeMeanSeries := RadarSeries{Name: "年级平均"}
	for _, s := range subjects {
		gradeMeanSeries.Value = append(gradeMeanSeries.Value, groupStats.StatsBySubject[s].Mean)
	}

	for _, table := range tables {
		classMeanSeries := RadarSeries{Name: table.TableName}
		for _, s := range subjects {
			classMeanSeries.Value = append(classMeanSeries.Value, table.TableStats.StatsBySubject[s].Mean)
		}
		classCharts.ClassProfileRadar[table.TableName] = RadarChartData{
			Indicator: radarIndicator,
			Series:    []RadarSeries{classMeanSeries, gradeMeanSeries},
			Title:     fmt.Sprintf("%s 学科平均分画像", table.TableName),
		}
	}

	// --- 学生个体层级图表生成 ---
	studentCharts := &chartData.StudentLevelCharts
	studentCharts.SubjectVsSubjectScatter = make(map[string]ScatterPlotData)

	allStudentsFlat := make([]*types.StudentReport, 0)
	for _, table := range tables {
		allStudentsFlat = append(allStudentsFlat, table.Students...)
	}

	// 学科vs学科散点图 (itertools.combinations的Go实现)
	for i := 0; i < len(subjects); i++ {
		for j := i + 1; j < len(subjects); j++ {
			sub1, sub2 := subjects[i], subjects[j]
			scatterData := make([][]interface{}, len(allStudentsFlat))
			for k, s := range allStudentsFlat {
				scatterData[k] = []interface{}{
					s.Scores.RawScores[sub1],
					s.Scores.RawScores[sub2],
					s.StudentName,
					s.TableName,
				}
			}
			key := fmt.Sprintf("%s_vs_%s", sub1, sub2)
			studentCharts.SubjectVsSubjectScatter[key] = ScatterPlotData{
				Data:      scatterData,
				XAxisName: sub1,
				YAxisName: sub2,
				Title:     fmt.Sprintf("%s vs %s 成绩散点图", sub1, sub2),
			}
		}
	}

	return chartData, nil
}
