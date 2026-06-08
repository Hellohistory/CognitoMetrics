// in: internal/analyzer/types/types.go

package types

// AnalysisInput 是整个分析流程的输入
type AnalysisInput struct {
	GroupName      string
	FullMarks      map[string]float64
	Tables         []*ClassInputData
	ExamID         uint
	PersistMetrics bool
}

// ClassInputData 代表一个班级的原始输入数据
type ClassInputData struct {
	TableName string
	Students  []*StudentInput
}

// StudentInput 代表一个学生的原始分数
type StudentInput struct {
	ID          uint // 学生在数据库中的ID
	StudentName string
	Scores      map[string]float64
	TableName   string
	TotalScore  float64
}

// AnalysisReport 是最终生成的完整分析报告
type AnalysisReport struct {
	GroupName  string             `json:"groupName"`
	FullMarks  map[string]float64 `json:"fullMarks"`
	GroupStats *LevelStats        `json:"groupStats"`
	Tables     []*ClassReport     `json:"tables"`
	Error      string             `json:"error,omitempty"`
}

// LevelStats 可用于年级 (Group) 或班级 (Table) 层面的统计
type LevelStats struct {
	StatsBySubject    map[string]*SubjectStats      `json:"statsBySubject"`
	CorrelationMatrix map[string]map[string]float64 `json:"correlationMatrix,omitempty"`
}

// ClassReport 包含一个班级的完整分析结果
type ClassReport struct {
	TableName  string           `json:"tableName"`
	TableStats *LevelStats      `json:"tableStats"`
	Students   []*StudentReport `json:"students"`
}

// StudentReport 包含一个学生的完整画像数据
type StudentReport struct {
	ID          uint           `json:"id"` // 学生ID
	StudentName string         `json:"studentName"`
	TableName   string         `json:"tableName"`
	TotalScore  float64        `json:"totalScore"`
	ClassRank   int            `json:"classRank"`
	GradeRank   int            `json:"gradeRank"`
	Profile     string         `json:"profile"`
	Ranks       StudentRanks   `json:"ranks"`
	Scores      StudentScores  `json:"scores"`
	Metrics     StudentMetrics `json:"metrics"`
}

// SubjectStats 包含单科或总分的所有描述性统计量
type SubjectStats struct {
	Count                   int                `json:"count"`
	Mean                    float64            `json:"mean"`
	StdDev                  float64            `json:"stdDev"`
	Variance                float64            `json:"variance"`
	Min                     float64            `json:"min"`
	Q1                      float64            `json:"q1"`
	Median                  float64            `json:"median"`
	Q3                      float64            `json:"q3"`
	Max                     float64            `json:"max"`
	Range                   float64            `json:"range"`
	ExcellentRate           float64            `json:"excellentRate"`
	GoodRate                float64            `json:"goodRate"`
	PassRate                float64            `json:"passRate"`
	LowScoreRate            float64            `json:"lowScoreRate"`
	Difficulty              float64            `json:"difficulty"`
	Skewness                float64            `json:"skewness"`
	Kurtosis                float64            `json:"kurtosis"`
	FullMarkCount           int                `json:"fullMarkCount"`
	ZeroMarkCount           int                `json:"zeroMarkCount"`
	BoxPlotData             map[string]float64 `json:"boxPlotData"`
	FrequencyDistribution   map[string]int     `json:"frequencyDistribution"`
	DiscriminationIndex     float64            `json:"discriminationIndex"`
	HighAchieverPenetration float64            `json:"highAchieverPenetration"`
	StrugglerSupportIndex   float64            `json:"strugglerSupportIndex"`
	AcademicCoreDensity     float64            `json:"academicCoreDensity"`
	HomogeneityIndex        float64            `json:"homogeneityIndex,omitempty"`
	QuartileCompetitiveness map[string]float64 `json:"quartileCompetitiveness,omitempty"`
	RawScores               []float64          `json:"-"`
}

// StudentScores 包含学生的各类分数
type StudentScores struct {
	RawScores  map[string]float64 `json:"rawScores"`
	ZScores    map[string]float64 `json:"zScores"`
	TScores    map[string]float64 `json:"tScores"`
	ScoreRates map[string]float64 `json:"scoreRates"`
}

// StudentMetrics 包含学生的各类高级指标
type StudentMetrics struct {
	ImbalanceIndex      float64            `json:"imbalanceIndex"`
	StrengthSubjects    []SubjectTScore    `json:"strengthSubjects"`
	WeaknessSubjects    []SubjectTScore    `json:"weaknessSubjects"`
	ContributionScore   map[string]float64 `json:"contributionScore"`
	SpecializationIndex float64            `json:"specializationIndex"`
	History             *HistoricalMetrics `json:"history,omitempty"`
	// NEW: 新增缺失的指标
	PointsToPass      float64 `json:"pointsToPass,omitempty"`
	PointsToExcellent float64 `json:"pointsToExcellent,omitempty"`
}

// StudentRanks 包含学生的所有排名信息
type StudentRanks struct {
	TotalScore RankInfo            `json:"totalScore"`
	Subjects   map[string]RankInfo `json:"subjects"`
}

// RankInfo 存储单项排名
type RankInfo struct {
	ClassRank           int     `json:"classRank"`
	GradeRank           int     `json:"gradeRank"`
	ClassPercentileRank float64 `json:"classPercentileRank"`
	GradePercentileRank float64 `json:"gradePercentileRank"`
}

// SubjectTScore 用于强弱科分析
type SubjectTScore struct {
	Subject string  `json:"subject"`
	TScore  float64 `json:"tScore"`
}

// HistoricalMetrics 存储历史数据分析结果
type HistoricalMetrics struct {
	Trend                         map[string]interface{} `json:"trend"`
	Stability                     map[string]interface{} `json:"stability"`
	GradePercentileRankSlope      float64                `json:"gradePercentileRankSlope"`
	TotalTScoreVolatility         float64                `json:"totalTScoreVolatility"`
	GradePercentileRankVolatility float64                `json:"gradePercentileRankVolatility"`
}

// HistoricalExam 代表一次过去考试的成绩记录
type HistoricalExam struct {
	ExamName            string
	ExamDate            string
	Scores              map[string]float64
	TotalScore          float64
	GradePercentileRank float64
	TotalTScore         float64
}

// StudentHistory 包含一个学生的所有历史考试记录
type StudentHistory struct {
	AllExams []*HistoricalExam
	LastExam *HistoricalExam
}
