// in: internal/charts/types.go

package charts

// ChartData 是所有图表数据的根容器
type ChartData struct {
	GradeLevelCharts      GradeLevelCharts      `json:"grade_level_charts"`
	ClassComparisonCharts ClassComparisonCharts `json:"class_comparison_charts"`
	StudentLevelCharts    StudentLevelCharts    `json:"student_level_charts"`
}

// GradeLevelCharts 年级层级图表结构
type GradeLevelCharts struct {
	ScoreDistributionHistogram map[string]HistogramData `json:"score_distribution_histogram"`
	SubjectCorrelationHeatmap  HeatmapData              `json:"subject_correlation_heatmap"`
	SubjectDifficultyScatter   ScatterPlotData          `json:"subject_difficulty_discrimination_scatter"`
}

// ClassComparisonCharts 班级对比图表结构
type ClassComparisonCharts struct {
	MetricsBarChart          map[string]map[string]BarChartData `json:"metrics_bar_chart"`
	ScoreDistributionBoxplot map[string]BoxplotData             `json:"score_distribution_boxplot"`
	ClassProfileRadar        map[string]RadarChartData          `json:"class_profile_radar"`
}

// StudentLevelCharts 学生个体层级图表结构
type StudentLevelCharts struct {
	SubjectVsSubjectScatter map[string]ScatterPlotData `json:"subject_vs_subject_scatter"`
}

// 通用图表组件结构

type HistogramData struct {
	Categories []string `json:"categories"`
	SeriesData []int    `json:"series_data"`
	SeriesName string   `json:"series_name"`
}

type HeatmapData struct {
	XAxisLabels []string        `json:"x_axis_labels"`
	YAxisLabels []string        `json:"y_axis_labels"`
	Data        [][]interface{} `json:"data"` // 使用 interface{} 来容纳 [x, y, value]
	Title       string          `json:"title"`
}

type ScatterPlotData struct {
	Data      [][]interface{} `json:"data"` // 使用 interface{} 来容纳 [x, y, name, category, ...]
	XAxisName string          `json:"x_axis_name"`
	YAxisName string          `json:"y_axis_name"`
	Title     string          `json:"title"`
}

type BarChartData struct {
	Categories []string  `json:"categories"`
	SeriesData []float64 `json:"series_data"`
	SeriesName string    `json:"series_name"`
}

type BoxplotData struct {
	Categories []string    `json:"categories"`
	Data       [][]float64 `json:"data"` // [min, q1, median, q3, max]
	Title      string      `json:"title"`
}

type RadarChartData struct {
	Indicator []RadarIndicator `json:"indicator"`
	Series    []RadarSeries    `json:"series"`
	Title     string           `json:"title"`
}

type RadarIndicator struct {
	Name string  `json:"name"`
	Max  float64 `json:"max"`
}

type RadarSeries struct {
	Name  string    `json:"name"`
	Value []float64 `json:"value"`
}
