// in: internal/analyzer/stats/stats.go

package stats

import (
	"fmt"
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"

	"CognitoMetrics/internal/analyzer/types"
)

// Round 将 float64 四舍五入到指定的精度。
func Round(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Round(val*p) / p
}

// CalculateCorrelation 计算两组分数的皮尔逊相关系数。
func CalculateCorrelation(arr1, arr2 []float64) float64 {
	if len(arr1) < 2 || len(arr1) != len(arr2) {
		return 0.0
	}
	correlation := stat.Correlation(arr1, arr2, nil)
	return Round(correlation, 3)
}

// CalculateGini 计算基尼系数，衡量分布的均衡性。
func CalculateGini(arr []float64) float64 {
	if len(arr) == 0 {
		return 0.0
	}
	sortedArr := make([]float64, len(arr))
	copy(sortedArr, arr)
	sort.Float64s(sortedArr)

	var totalSum, numerator float64
	n := float64(len(sortedArr))

	for i, val := range sortedArr {
		totalSum += val
		numerator += (2*float64(i+1) - n - 1) * val
	}

	if totalSum == 0 {
		return 0.0
	}

	denominator := n * totalSum
	if denominator == 0 {
		return 0.0
	}
	return numerator / denominator
}

// CalculateSkewnessKurtosis 计算分布的偏度和峰度。
func CalculateSkewnessKurtosis(arr []float64) (skewness, kurtosis float64) {
	if len(arr) < 4 {
		return 0.0, 0.0
	}
	skewness = stat.Skew(arr, nil)

	mean := stat.Mean(arr, nil)
	stdDev := stat.StdDev(arr, nil)
	if stdDev == 0 {
		return Round(skewness, 3), 0.0
	}

	var m4 float64
	for _, val := range arr {
		m4 += math.Pow(val-mean, 4)
	}
	m4 /= float64(len(arr))
	kurtosis = (m4 / math.Pow(stdDev, 4)) - 3
	return Round(skewness, 3), Round(kurtosis, 3)
}

// CalculateFrequencyDistribution 计算分数的频率分布。
func CalculateFrequencyDistribution(scores []float64, fullMark float64, binSize int) map[string]int {
	if binSize <= 0 || len(scores) == 0 {
		return map[string]int{}
	}
	bins := make(map[string]int)
	intFullMark := int(math.Ceil(fullMark))

	for i := 0; i < intFullMark; i += binSize {
		upperBound := i + binSize
		if upperBound > intFullMark {
			upperBound = intFullMark
		}
		key := fmt.Sprintf("%d-%d", i, upperBound)
		bins[key] = 0
	}

	lastKey := fmt.Sprintf("%d-%d", (intFullMark/binSize)*binSize, intFullMark)
	if intFullMark%binSize == 0 && intFullMark > 0 {
		lastKey = fmt.Sprintf("%d-%d", intFullMark-binSize, intFullMark)
	}

	for _, score := range scores {
		if score == fullMark && lastKey != "" {
			bins[lastKey]++
			continue
		}
		if score >= fullMark {
			if lastKey != "" {
				bins[lastKey]++
			}
			continue
		}
		binIndex := int(score / float64(binSize))
		lowerBound := binIndex * binSize
		upperBound := lowerBound + binSize
		if upperBound > intFullMark {
			upperBound = intFullMark
		}
		key := fmt.Sprintf("%d-%d", lowerBound, upperBound)
		if _, ok := bins[key]; ok {
			bins[key]++
		}
	}
	return bins
}

// CalculateDiscriminationIndex 计算考试的区分度指数。
func CalculateDiscriminationIndex(scores []float64, fullMark float64) float64 {
	n := len(scores)
	if n < 10 || fullMark == 0 {
		return 0.0
	}
	sortedScores := make([]float64, n)
	copy(sortedScores, scores)
	sort.Float64s(sortedScores)
	topN := int(math.Max(1, float64(n)*0.27))

	highScores := sortedScores[n-topN:]
	lowScores := sortedScores[:topN]

	if len(highScores) == 0 || len(lowScores) == 0 {
		return 0.0
	}
	highAvg := stat.Mean(highScores, nil)
	lowAvg := stat.Mean(lowScores, nil)

	return Round((highAvg-lowAvg)/fullMark, 3)
}

// AnalyzeTrendSlope 使用最小二乘法计算历史数值的趋势斜率。
func AnalyzeTrendSlope(historicalValues []float64) float64 {
	type point struct{ x, y float64 }
	var validPoints []point
	for i, y := range historicalValues {
		validPoints = append(validPoints, point{x: float64(i + 1), y: y})
	}
	n := float64(len(validPoints))
	if n < 2 {
		return 0.0
	}
	var sumX, sumY, sumXY, sumXX float64
	for _, p := range validPoints {
		sumX += p.x
		sumY += p.y
		sumXY += p.x * p.y
		sumXX += p.x * p.x
	}
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return 0.0
	}
	numerator := n*sumXY - sumX*sumY
	slope := numerator / denominator
	return Round(slope, 3)
}

// CalculateDescriptiveStats 为一组分数计算一整套描述性统计量。
func CalculateDescriptiveStats(scores []float64, fullMark float64) *types.SubjectStats {
	count := len(scores)
	if count == 0 {
		return &types.SubjectStats{}
	}
	stats := &types.SubjectStats{Count: count}
	stats.Mean = stat.Mean(scores, nil)
	stats.StdDev = stat.StdDev(scores, nil)
	stats.Variance = stat.Variance(scores, nil)

	sortedScores := make([]float64, count)
	copy(sortedScores, scores)
	sort.Float64s(sortedScores)

	stats.Min = sortedScores[0]
	stats.Max = sortedScores[count-1]
	stats.Range = stats.Max - stats.Min
	stats.Q1 = stat.Quantile(0.25, stat.Empirical, sortedScores, nil)
	stats.Median = stat.Quantile(0.50, stat.Empirical, sortedScores, nil)
	stats.Q3 = stat.Quantile(0.75, stat.Empirical, sortedScores, nil)

	var passCount, goodCount, excellentCount, fullMarkCount, zeroMarkCount int
	if fullMark > 0 {
		passThreshold := fullMark * 0.60
		goodThreshold := fullMark * 0.70
		excellentThreshold := fullMark * 0.85
		for _, s := range scores {
			if s >= passThreshold {
				passCount++
			}
			if s >= goodThreshold && s < excellentThreshold {
				goodCount++
			}
			if s >= excellentThreshold {
				excellentCount++
			}
			if s == fullMark {
				fullMarkCount++
			}
			if s == 0 {
				zeroMarkCount++
			}
		}
		stats.Difficulty = Round(stats.Mean/fullMark, 3)
		stats.PassRate = Round(float64(passCount)/float64(count), 3)
		stats.GoodRate = Round(float64(goodCount)/float64(count), 3)
		stats.ExcellentRate = Round(float64(excellentCount)/float64(count), 3)
		stats.LowScoreRate = Round(float64(count-passCount)/float64(count), 3)
	}
	stats.FullMarkCount = fullMarkCount
	stats.ZeroMarkCount = zeroMarkCount

	skew, kurt := CalculateSkewnessKurtosis(scores)
	stats.Skewness = skew
	stats.Kurtosis = kurt
	stats.BoxPlotData = map[string]float64{"min": stats.Min, "q1": stats.Q1, "median": stats.Median, "q3": stats.Q3, "max": stats.Max}
	stats.FrequencyDistribution = CalculateFrequencyDistribution(scores, fullMark, 10)
	return stats
}

// CalculateAdvancedGroupMetrics 计算群体的结构性指标。
func CalculateAdvancedGroupMetrics(scores []float64, stats *types.SubjectStats) {
	n := len(scores)
	if n < 10 {
		return
	}
	sortedScores := make([]float64, n)
	copy(sortedScores, scores)
	sort.Float64s(sortedScores)
	topN := int(math.Max(1, float64(n)*0.27))
	bottomN := int(math.Max(1, float64(n)*0.27))
	stats.HighAchieverPenetration = Round(stat.Mean(sortedScores[n-topN:], nil), 2)
	stats.StrugglerSupportIndex = Round(stat.Mean(sortedScores[:bottomN], nil), 2)
	if stats.StdDev > 0 {
		coreCount := 0
		lowerBound := stats.Mean - 0.5*stats.StdDev
		upperBound := stats.Mean + 0.5*stats.StdDev
		for _, s := range sortedScores {
			if s >= lowerBound && s <= upperBound {
				coreCount++
			}
		}
		stats.AcademicCoreDensity = Round(float64(coreCount)/float64(n), 3)
	} else {
		stats.AcademicCoreDensity = 1.0
	}
}

// CalculateAdvancedStudentMetrics 计算学生的个体高级指标。
func CalculateAdvancedStudentMetrics(report *types.StudentReport, classScoresBySubject map[string][]float64) {
	contribution := make(map[string]float64)
	for subject, scoresInClass := range classScoresBySubject {
		studentScore, ok := report.Scores.RawScores[subject]
		if !ok || len(scoresInClass) < 2 {
			contribution[subject] = 0
			continue
		}
		var classTotal float64
		for _, s := range scoresInClass {
			classTotal += s
		}
		othersMean := (classTotal - studentScore) / float64(len(scoresInClass)-1)
		contribution[subject] = Round(studentScore-othersMean, 2)
	}
	report.Metrics.ContributionScore = contribution

	tScoresList := make([]float64, 0, len(report.Scores.TScores))
	for subject, tScore := range report.Scores.TScores {
		if subject != "totalScore" {
			tScoresList = append(tScoresList, tScore)
		}
	}
	if len(tScoresList) > 1 {
		report.Metrics.SpecializationIndex = Round(CalculateGini(tScoresList), 3)
	} else {
		report.Metrics.SpecializationIndex = 0.0
	}
}
