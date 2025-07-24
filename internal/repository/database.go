package repository

import (
	"CognitoMetrics/internal/analyzer/types"
	"CognitoMetrics/internal/models"
	"errors"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"sort"
	"time"
)

type Repository struct {
	DB              *gorm.DB
	subjectNameToID map[string]uint
	subjectIDToName map[uint]string
}

func New(dbPath string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Exam{}, &models.Subject{}, &models.ExamSubject{},
		&models.Class{}, &models.Student{}, &models.Score{}, &models.AnalysisReport{},
	)
	if err != nil {
		return nil, err
	}

	repo := &Repository{DB: db}
	if err := repo.seedAndLoadSubjectCache(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repository) seedAndLoadSubjectCache() error {
	var count int64
	r.DB.Model(&models.Subject{}).Count(&count)
	if count == 0 {
		log.Println("Seeding subjects...")
		subjects := []*models.Subject{{Name: "语文"}, {Name: "数学"}, {Name: "英语"}}
		if err := r.DB.Create(&subjects).Error; err != nil {
			return err
		}
	}

	var subjects []models.Subject
	r.DB.Find(&subjects)
	r.subjectIDToName = make(map[uint]string)
	r.subjectNameToID = make(map[string]uint)
	for _, s := range subjects {
		r.subjectIDToName[s.ID] = s.Name
		r.subjectNameToID[s.Name] = s.ID
	}
	return nil
}

func (r *Repository) LoadAnalysisData(examID uint) (*types.AnalysisInput, map[string]*types.StudentHistory, error) {
	var exam models.Exam
	if err := r.DB.First(&exam, examID).Error; err != nil {
		return nil, nil, err
	}

	var examSubjects []models.ExamSubject
	r.DB.Where("exam_id = ?", exam.ID).Find(&examSubjects)
	fullMarks := make(map[string]float64, len(examSubjects))
	for _, es := range examSubjects {
		fullMarks[r.subjectIDToName[es.SubjectID]] = es.FullMark
	}

	var scores []models.Score
	r.DB.Where("exam_id = ?", examID).Preload("Student.Class").Find(&scores)

	studentIDsInScope := make(map[uint]bool)
	tablesMap := make(map[string]*types.ClassInputData)
	for _, score := range scores {
		studentIDsInScope[score.StudentID] = true
		className := score.Student.Class.Name
		if _, ok := tablesMap[className]; !ok {
			tablesMap[className] = &types.ClassInputData{TableName: className}
		}

		var studentInput *types.StudentInput
		for _, s := range tablesMap[className].Students {
			if s.ID == score.StudentID {
				studentInput = s
				break
			}
		}
		if studentInput == nil {
			studentInput = &types.StudentInput{
				ID:          score.StudentID,
				StudentName: score.Student.Name,
				Scores:      make(map[string]float64),
			}
			tablesMap[className].Students = append(tablesMap[className].Students, studentInput)
		}
		studentInput.Scores[r.subjectIDToName[score.SubjectID]] = score.Score
	}

	analysisInput := &types.AnalysisInput{
		GroupName: exam.Name,
		FullMarks: fullMarks,
		ExamID:    examID,
	}
	for _, table := range tablesMap {
		analysisInput.Tables = append(analysisInput.Tables, table)
	}

	historyMap, err := r.loadHistoricalData(studentIDsInScope, exam.ExamDate)
	if err != nil {
		return nil, nil, err
	}

	return analysisInput, historyMap, nil
}

func (r *Repository) loadHistoricalData(studentIDs map[uint]bool, currentExamDate time.Time) (map[string]*types.StudentHistory, error) {
	if len(studentIDs) == 0 {
		return make(map[string]*types.StudentHistory), nil
	}
	var sids []uint
	for id := range studentIDs {
		sids = append(sids, id)
	}

	var historicalScores []models.Score
	err := r.DB.Joins("Exam", r.DB.Where("exam_date < ?", currentExamDate)).
		Where("student_id IN ?", sids).
		Preload("Student").Preload("Exam").
		Order("exam_date asc").
		Find(&historicalScores).Error
	if err != nil {
		return nil, err
	}

	tempHistoryMap := make(map[string]map[uint]*types.HistoricalExam)
	studentIDToName := make(map[uint]string)

	for _, score := range historicalScores {
		studentName := score.Student.Name
		studentIDToName[score.StudentID] = studentName
		examID := score.ExamID

		if _, ok := tempHistoryMap[studentName]; !ok {
			tempHistoryMap[studentName] = make(map[uint]*types.HistoricalExam)
		}
		if _, ok := tempHistoryMap[studentName][examID]; !ok {
			tempHistoryMap[studentName][examID] = &types.HistoricalExam{
				ExamName: score.Exam.Name,
				ExamDate: score.Exam.ExamDate.Format("2006-01-02"),
				Scores:   make(map[string]float64),
			}
		}
		subjectName := r.subjectIDToName[score.SubjectID]
		tempHistoryMap[studentName][examID].Scores[subjectName] = score.Score

		// For total score T-score and percentile, we assume they are stored on a special "total" score record,
		// or calculated. For simplicity, we'll aggregate them here.
		// A more robust solution might have total scores pre-calculated and stored.
		if score.TScore > 0 {
			tempHistoryMap[studentName][examID].TotalTScore += score.TScore // Simplified aggregation
		}
		if score.GradePercentileRank > 0 {
			tempHistoryMap[studentName][examID].GradePercentileRank += score.GradePercentileRank // Simplified aggregation
		}
	}

	finalHistoryMap := make(map[string]*types.StudentHistory)
	for studentName, examsMap := range tempHistoryMap {
		studentHistory := &types.StudentHistory{AllExams: []*types.HistoricalExam{}}
		for _, examData := range examsMap {
			var totalScore float64
			for _, s := range examData.Scores {
				totalScore += s
			}
			examData.TotalScore = totalScore
			examData.TotalTScore /= float64(len(examData.Scores))         // Averaging
			examData.GradePercentileRank /= float64(len(examData.Scores)) // Averaging
			studentHistory.AllExams = append(studentHistory.AllExams, examData)
		}
		sort.Slice(studentHistory.AllExams, func(i, j int) bool {
			return studentHistory.AllExams[i].ExamDate < studentHistory.AllExams[j].ExamDate
		})
		if len(studentHistory.AllExams) > 0 {
			studentHistory.LastExam = studentHistory.AllExams[len(studentHistory.AllExams)-1]
		}
		finalHistoryMap[studentName] = studentHistory
	}

	return finalHistoryMap, nil
}

func (r *Repository) UpdateScoresWithMetrics(report *types.AnalysisReport, examID uint) error {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	log.Printf("Starting metric write-back for Exam ID: %d...", examID)

	for _, table := range report.Tables {
		for _, student := range table.Students {
			for subjectName, rawScore := range student.Scores.RawScores {
				subjectID, ok := r.subjectNameToID[subjectName]
				if !ok {
					continue
				}

				tScore := student.Scores.TScores[subjectName]
				percentileRank := student.Ranks.Subjects[subjectName].GradePercentileRank

				res := tx.Model(&models.Score{}).
					Where("student_id = ? AND exam_id = ? AND subject_id = ?", student.ID, examID, subjectID).
					Where("score = ?", rawScore). // Add score to where clause to be more specific
					Updates(map[string]interface{}{
						"t_score":               tScore,
						"grade_percentile_rank": percentileRank,
					})

				if res.Error != nil {
					tx.Rollback()
					log.Printf("ERROR: Failed to write back for Student %d, Subject %d. Rolling back. Error: %v", student.ID, subjectID, res.Error)
					return res.Error
				}
			}
		}
	}

	log.Println("Metric write-back successful. Committing transaction.")
	return tx.Commit().Error
}

func (r *Repository) ImportExamData(input *types.AnalysisInput) (uint, error) {
	// (This function is from your provided code, can be included for completeness)
	return 0, errors.New("import function not fully shown, but can be added here")
}
