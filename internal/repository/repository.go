// file: internal/repository/repository.go
package repository

import (
	"CognitoMetrics/internal/analyzer/types"
	"CognitoMetrics/internal/models"
	"CognitoMetrics/internal/schemas"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Repository 结构体封装了所有数据库操作
type Repository struct {
	DB              *gorm.DB
	subjectNameToID map[string]uint
	subjectIDToName map[uint]string
	subjectMu       sync.RWMutex
}

// New 创建一个新的 Repository 实例，并进行数据库迁移和缓存预热
func New(dbPath string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Grade{}, &models.Class{}, &models.Student{},
		&models.Exam{}, &models.Subject{}, &models.ExamSubject{},
		&models.Score{}, &models.AnalysisReport{},
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

// seedAndLoadSubjectCache 初始化种子数据并加载学科缓存
func (r *Repository) seedAndLoadSubjectCache() error {
	var count int64
	r.DB.Model(&models.Subject{}).Count(&count)
	if count == 0 {
		log.Println("数据库为空，正在植入默认学科...")
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

func (r *Repository) cacheSubject(subject models.Subject) {
	r.subjectMu.Lock()
	defer r.subjectMu.Unlock()
	r.subjectIDToName[subject.ID] = subject.Name
	r.subjectNameToID[subject.Name] = subject.ID
}

func (r *Repository) getSubjectIDByName(name string) (uint, bool) {
	r.subjectMu.RLock()
	defer r.subjectMu.RUnlock()
	id, ok := r.subjectNameToID[name]
	return id, ok
}

func (r *Repository) getSubjectNameByID(id uint) (string, bool) {
	r.subjectMu.RLock()
	defer r.subjectMu.RUnlock()
	name, ok := r.subjectIDToName[id]
	return name, ok
}

func reportTableName(class models.Class) string {
	if class.Grade.Name == "" {
		return class.Name
	}
	return fmt.Sprintf("%s-%s", class.Grade.Name, class.Name)
}

// paginate 是一个 GORM Scope，用于分页
func paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//==============================================================================
//--- 年级 (Grade) 相关方法 ---
//==============================================================================

// CreateGrade 创建新年级
func (r *Repository) CreateGrade(gradeIn schemas.GradeCreate) (*models.Grade, error) {
	grade := &models.Grade{Name: gradeIn.Name}
	if err := r.DB.Create(grade).Error; err != nil {
		return nil, err
	}
	return grade, nil
}

// GetGradeByID 根据ID获取年级
func (r *Repository) GetGradeByID(id uint) (*models.Grade, error) {
	var grade models.Grade
	if err := r.DB.First(&grade, id).Error; err != nil {
		return nil, err
	}
	return &grade, nil
}

// GetGradeByName 根据名称获取年级
func (r *Repository) GetGradeByName(name string) (*models.Grade, error) {
	var grade models.Grade
	if err := r.DB.Where("name = ?", name).First(&grade).Error; err != nil {
		return nil, err
	}
	return &grade, nil
}

// ListGrades 获取所有年级列表
func (r *Repository) ListGrades(skip, limit int) ([]models.Grade, error) {
	var grades []models.Grade
	err := r.DB.Offset(skip).Limit(limit).Find(&grades).Error
	return grades, err
}

// UpdateGrade 更新年级信息
func (r *Repository) UpdateGrade(grade *models.Grade, gradeIn schemas.GradeUpdate) (*models.Grade, error) {
	if gradeIn.Name != "" {
		grade.Name = gradeIn.Name
	}
	err := r.DB.Save(grade).Error
	return grade, err
}

// DeleteGradeByID 删除年级，会检查其下是否有班级
func (r *Repository) DeleteGradeByID(id uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var grade models.Grade
		if err := tx.Preload("Classes").First(&grade, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if len(grade.Classes) > 0 {
			return fmt.Errorf("无法删除年级 '%s'，因为它下面仍有班级", grade.Name)
		}
		return tx.Delete(&grade).Error
	})
}

//==============================================================================
//--- 班级 (Class) 相关方法 ---
//==============================================================================

// CreateClass 创建新班级
func (r *Repository) CreateClass(classIn schemas.ClassCreate) (*models.Class, error) {
	class := &models.Class{
		Name:           classIn.Name,
		EnrollmentYear: classIn.EnrollmentYear,
		GradeID:        classIn.GradeID,
	}
	err := r.DB.Create(class).Error
	return class, err
}

// ListClasses 分页获取班级列表
func (r *Repository) ListClasses(skip, limit int) ([]models.Class, error) {
	var classes []models.Class
	err := r.DB.Order("id asc").Offset(skip).Limit(limit).Find(&classes).Error
	return classes, err
}

// CountClasses 获取班级总数
func (r *Repository) CountClasses() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Class{}).Count(&count).Error
	return count, err
}

// GetClassByID 根据ID获取班级
func (r *Repository) GetClassByID(id uint) (*models.Class, error) {
	var class models.Class
	if err := r.DB.First(&class, id).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

// GetClassByGradeAndName 根据年级ID和班级名获取班级
func (r *Repository) GetClassByGradeAndName(gradeID uint, name string) (*models.Class, error) {
	var class models.Class
	err := r.DB.Where("grade_id = ? AND name = ?", gradeID, name).First(&class).Error
	return &class, err
}

// GetClassTreeData 获取年级-班级树状结构数据
func (r *Repository) GetClassTreeData() ([]models.Grade, error) {
	var grades []models.Grade
	err := r.DB.Preload("Classes.Students").Order("name asc").Find(&grades).Error
	return grades, err
}

// UpdateClass 更新班级信息
func (r *Repository) UpdateClass(class *models.Class, classIn schemas.ClassUpdate) (*models.Class, error) {
	if classIn.Name != "" {
		class.Name = classIn.Name
	}
	if classIn.EnrollmentYear != nil {
		class.EnrollmentYear = *classIn.EnrollmentYear
	}
	err := r.DB.Save(class).Error
	return class, err
}

// DeleteClassByID 删除班级，会检查其下是否有学生
func (r *Repository) DeleteClassByID(id uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var class models.Class
		if err := tx.Preload("Students").First(&class, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if len(class.Students) > 0 {
			return fmt.Errorf("无法删除班级 '%s'，因为它下面仍有学生", class.Name)
		}
		return tx.Delete(&class).Error
	})
}

//==============================================================================
//--- 学生 (Student) 相关方法 ---
//==============================================================================

// GenerateNewStudentNumbers 为指定班级生成一批学号
func (r *Repository) GenerateNewStudentNumbers(classID uint, count int) ([]string, error) {
	var targetClass models.Class
	if err := r.DB.First(&targetClass, classID).Error; err != nil {
		return nil, fmt.Errorf("班级ID %d 未找到", classID)
	}

	yearPrefix := strconv.Itoa(targetClass.EnrollmentYear)
	var latestStudentNo string
	r.DB.Model(&models.Student{}).Where("student_no LIKE ?", yearPrefix+"%").Select("MAX(student_no)").Row().Scan(&latestStudentNo)

	startSequence := 1
	if latestStudentNo != "" && len(latestStudentNo) > 4 {
		seq, err := strconv.Atoi(latestStudentNo[4:])
		if err == nil {
			startSequence = seq + 1
		}
	}

	newNos := make([]string, count)
	for i := 0; i < count; i++ {
		newNos[i] = fmt.Sprintf("%s%04d", yearPrefix, startSequence+i)
	}
	return newNos, nil
}

// BatchCreateStudents 在一个事务中批量创建学生
func (r *Repository) BatchCreateStudents(studentsIn []schemas.StudentCreate, studentNos []string) ([]models.Student, error) {
	newStudents := make([]models.Student, len(studentsIn))
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		for i, sIn := range studentsIn {
			newStudents[i] = models.Student{
				Name:      sIn.Name,
				ClassID:   sIn.ClassID,
				StudentNo: studentNos[i],
			}
		}
		return tx.Create(&newStudents).Error
	})
	return newStudents, err
}

// ListStudentsByClass 获取班级下的学生列表
func (r *Repository) ListStudentsByClass(classID uint, includeInactive bool) ([]models.Student, error) {
	var students []models.Student
	db := r.DB.Where("class_id = ?", classID)
	if !includeInactive {
		db = db.Where("is_active = ?", true)
	}
	err := db.Order("student_no asc").Find(&students).Error
	return students, err
}

// UpdateStudentStatus 更新单个学生状态
func (r *Repository) UpdateStudentStatus(id uint, isActive bool) (*models.Student, error) {
	var student models.Student
	if err := r.DB.First(&student, id).Error; err != nil {
		return nil, err
	}
	student.IsActive = isActive
	err := r.DB.Save(&student).Error
	return &student, err
}

// BatchUpdateStudentsStatus 批量更新学生状态
func (r *Repository) BatchUpdateStudentsStatus(ids []uint, isActive bool) error {
	return r.DB.Model(&models.Student{}).Where("id IN ?", ids).Update("is_active", isActive).Error
}

// BatchUpdateStudentsClass 批量更新学生班级
func (r *Repository) BatchUpdateStudentsClass(ids []uint, targetClassID uint) error {
	var count int64
	r.DB.Model(&models.Class{}).Where("id = ?", targetClassID).Count(&count)
	if count == 0 {
		return fmt.Errorf("目标班级ID %d 不存在", targetClassID)
	}
	return r.DB.Model(&models.Student{}).Where("id IN ?", ids).Update("class_id", targetClassID).Error
}

// GetStudentPerformanceHistory 获取学生历次考试表现 (精确版)
func (r *Repository) GetStudentPerformanceHistory(studentID uint) ([]schemas.PerformanceRecordSchema, error) {
	// 1. 查找所有已完成的、包含该学生的单场考试报告
	var reports []models.AnalysisReport
	err := r.DB.
		Joins("Exam"). // 关联考试以按日期排序
		Where("status = ? AND report_type = ? AND full_report_data != ''", "completed", "single").
		Order("`exams`.`exam_date` ASC").
		Find(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询分析报告失败: %w", err)
	}

	var performanceRecords []schemas.PerformanceRecordSchema

	// 2. 遍历每一份报告，解析JSON并查找学生数据
	for _, report := range reports {
		if report.FullReportData == "" {
			continue
		}

		var reportData types.AnalysisReport
		if err := json.Unmarshal([]byte(report.FullReportData), &reportData); err != nil {
			log.Printf("警告: 解析报告ID %d 的JSON数据失败: %v", report.ID, err)
			continue
		}

		studentFound := false
		for _, classTable := range reportData.Tables {
			for _, studentReport := range classTable.Students {
				if studentReport.ID == studentID {
					// 找到了学生，提取所需信息
					totalScore := studentReport.TotalScore
					classRank := studentReport.ClassRank
					gradeRank := studentReport.GradeRank
					record := schemas.PerformanceRecordSchema{
						ExamID:     report.ExamID,
						ExamName:   report.Exam.Name,
						ExamDate:   report.Exam.ExamDate,
						TotalScore: &totalScore,
						ClassRank:  &classRank,
						GradeRank:  &gradeRank,
					}
					performanceRecords = append(performanceRecords, record)
					studentFound = true
					break
				}
			}
			if studentFound {
				break
			}
		}
	}

	return performanceRecords, nil
}

// GetStudentByID 获取单个学生模型
func (r *Repository) GetStudentByID(id uint) (*models.Student, error) {
	var student models.Student
	if err := r.DB.First(&student, id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// UpdateStudent 更新学生信息
func (r *Repository) UpdateStudent(student *models.Student, studentIn schemas.StudentUpdate) (*models.Student, error) {
	if studentIn.Name != "" {
		student.Name = studentIn.Name
	}
	if studentIn.ClassID != nil {
		var count int64
		r.DB.Model(&models.Class{}).Where("id = ?", *studentIn.ClassID).Count(&count)
		if count == 0 {
			return nil, fmt.Errorf("目标班级ID %d 不存在", *studentIn.ClassID)
		}
		student.ClassID = *studentIn.ClassID
	}
	err := r.DB.Save(student).Error
	return student, err
}

// GetStudentDetailsByID 获取单个学生的详细信息，包括班级和年级名
func (r *Repository) GetStudentDetailsByID(id uint) (*schemas.StudentDetailSchema, error) {
	var student models.Student
	if err := r.DB.Preload("Class.Grade").First(&student, id).Error; err != nil {
		return nil, err
	}
	if student.Class.ID == 0 || student.Class.Grade.ID == 0 {
		return nil, errors.New("学生关联的班级或年级信息不完整")
	}
	return &schemas.StudentDetailSchema{
		ID:             student.ID,
		StudentNo:      student.StudentNo,
		Name:           student.Name,
		ClassID:        student.ClassID,
		IsActive:       student.IsActive,
		ClassName:      student.Class.Name,
		GradeName:      student.Class.Grade.Name,
		EnrollmentYear: student.Class.EnrollmentYear,
	}, nil
}

//==============================================================================
//--- 成绩 (Score) 相关方法 ---
//==============================================================================

// UpsertSingleScore 更新或插入单条成绩
func (r *Repository) UpsertSingleScore(scoreIn schemas.SingleScoreUpdate) error {
	subjectID, ok := r.getSubjectIDByName(scoreIn.SubjectName)
	if !ok {
		return fmt.Errorf("学科 '%s' 未找到", scoreIn.SubjectName)
	}
	if scoreIn.Score == nil {
		return r.DB.Unscoped().
			Where("student_id = ? AND exam_id = ? AND subject_id = ?", scoreIn.StudentID, scoreIn.ExamID, subjectID).
			Delete(&models.Score{}).Error
	}
	score := models.Score{
		StudentID: scoreIn.StudentID,
		ExamID:    scoreIn.ExamID,
		SubjectID: subjectID,
		Score:     *scoreIn.Score,
	}
	return r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "student_id"}, {Name: "exam_id"}, {Name: "subject_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"score"}),
	}).Create(&score).Error
}

// BatchUpsertScores 批量更新或插入成绩
func (r *Repository) BatchUpsertScores(batchIn schemas.ScoresBatchInput) (int, error) {
	var scoresToUpsert []models.Score
	var scoresToDelete []models.Score
	for _, scoreInput := range batchIn.Scores {
		for subjName, scoreVal := range scoreInput.SubjectScores {
			subjectID, ok := r.getSubjectIDByName(subjName)
			if !ok {
				return 0, fmt.Errorf("学科 '%s' 未找到", subjName)
			}
			if scoreVal == nil {
				scoresToDelete = append(scoresToDelete, models.Score{
					StudentID: scoreInput.StudentID,
					ExamID:    batchIn.ExamID,
					SubjectID: subjectID,
				})
				continue
			}
			scoresToUpsert = append(scoresToUpsert, models.Score{
				StudentID: scoreInput.StudentID,
				ExamID:    batchIn.ExamID,
				SubjectID: subjectID,
				Score:     *scoreVal,
			})
		}
	}
	if len(scoresToUpsert) == 0 && len(scoresToDelete) == 0 {
		return 0, errors.New("没有有效的成绩数据可以录入")
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		for _, score := range scoresToDelete {
			if err := tx.Unscoped().
				Where("student_id = ? AND exam_id = ? AND subject_id = ?", score.StudentID, score.ExamID, score.SubjectID).
				Delete(&models.Score{}).Error; err != nil {
				return err
			}
		}
		if len(scoresToUpsert) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "student_id"}, {Name: "exam_id"}, {Name: "subject_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"score"}),
		}).Create(&scoresToUpsert).Error
	})

	return len(scoresToUpsert) + len(scoresToDelete), err
}

// GetScoresForClassInExam 获取指定班级在某场考试中的所有成绩记录
func (r *Repository) GetScoresForClassInExam(examID, classID uint) ([]models.Score, error) {
	var studentIDs []uint
	if err := r.DB.Model(&models.Student{}).Where("class_id = ?", classID).Pluck("id", &studentIDs).Error; err != nil {
		return nil, err
	}
	if len(studentIDs) == 0 {
		return []models.Score{}, nil
	}
	var scores []models.Score
	err := r.DB.Preload("Subject").Where("exam_id = ? AND student_id IN ?", examID, studentIDs).Find(&scores).Error
	return scores, err
}

//==============================================================================
//--- 报告 (Report) 相关方法 ---
//==============================================================================

// CreateAnalysisReport 在数据库中创建一条新的分析报告记录
func (r *Repository) CreateAnalysisReport(name string, examID uint, desc string, reportType string) (*models.AnalysisReport, error) {
	report := &models.AnalysisReport{
		ReportName:        name,
		ExamID:            examID,
		SourceDescription: desc,
		ReportType:        reportType,
		Status:            "processing",
	}
	if err := r.DB.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

// GetReportByID 通过ID获取报告详情，并预加载关联的考试信息
func (r *Repository) GetReportByID(id uint) (*models.AnalysisReport, error) {
	var report models.AnalysisReport
	if err := r.DB.Preload("Exam").First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports 分页并根据条件筛选报告列表
func (r *Repository) ListReports(page, pageSize int, query, status, reportType string) ([]models.AnalysisReport, int64, error) {
	var reports []models.AnalysisReport
	var total int64

	db := r.DB.Model(&models.AnalysisReport{}).Preload("Exam")
	if query != "" {
		db = db.Where("report_name LIKE ?", "%"+query+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if reportType != "" {
		db = db.Where("report_type = ?", reportType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("created_at desc").Scopes(paginate(page, pageSize)).Find(&reports).Error
	return reports, total, err
}

// UpdateReportStatus 更新报告的状态和错误信息
func (r *Repository) UpdateReportStatus(id uint, status string, errorMsg string) error {
	return r.DB.Model(&models.AnalysisReport{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"error_message": errorMsg,
	}).Error
}

// UpdateReportAIStatus 更新报告的AI分析状态
func (r *Repository) UpdateReportAIStatus(id uint, status string) error {
	return r.DB.Model(&models.AnalysisReport{}).Where("id = ?", id).Update("ai_analysis_status", status).Error
}

// SaveFullReport 保存完整的分析结果到报告记录中
func (r *Repository) SaveFullReport(id uint, reportData *types.AnalysisReport) error {
	jsonData, err := json.Marshal(reportData)
	if err != nil {
		return err
	}
	return r.DB.Model(&models.AnalysisReport{}).Where("id = ?", id).Update("full_report_data", string(jsonData)).Error
}

// DeleteReportByID 通过ID删除报告
func (r *Repository) DeleteReportByID(id uint) error {
	result := r.DB.Delete(&models.AnalysisReport{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateAnalysisReport 更新整个报告对象（用于AI分析结果回写）
func (r *Repository) UpdateAnalysisReport(report *models.AnalysisReport) error {
	return r.DB.Save(report).Error
}

//==============================================================================
//--- 考试 (Exam) 相关方法 ---
//==============================================================================

// GetExamByID 通过ID获取考试信息
func (r *Repository) GetExamByID(id uint) (*models.Exam, error) {
	var exam models.Exam
	if err := r.DB.First(&exam, id).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

// CreateExamWithSubjects 创建一场考试并关联学科
func (r *Repository) CreateExamWithSubjects(examIn schemas.ExamWithSubjectsCreate) (*models.Exam, error) {
	var dbExam models.Exam
	var subjectsToCache []models.Subject
	examDate, err := examIn.ParsedExamDate()
	if err != nil {
		return nil, fmt.Errorf("考试日期格式无效: %w", err)
	}
	err = r.DB.Transaction(func(tx *gorm.DB) error {
		dbExam = models.Exam{
			Name:     examIn.Name,
			ExamDate: examDate,
			Status:   "draft",
		}
		if err := tx.Create(&dbExam).Error; err != nil {
			return err
		}
		for _, subIn := range examIn.Subjects {
			var subject models.Subject
			if err := tx.Where(models.Subject{Name: subIn.Name}).FirstOrCreate(&subject).Error; err != nil {
				return err
			}
			subjectsToCache = append(subjectsToCache, subject)
			examSubject := models.ExamSubject{
				ExamID:    dbExam.ID,
				SubjectID: subject.ID,
				FullMark:  subIn.FullMark,
			}
			if err := tx.Create(&examSubject).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, subject := range subjectsToCache {
		r.cacheSubject(subject)
	}
	return &dbExam, nil
}

// ListExams 获取所有考试列表，按日期降序
func (r *Repository) ListExams() ([]models.Exam, error) {
	var exams []models.Exam
	err := r.DB.Order("exam_date desc").Find(&exams).Error
	return exams, err
}

// GetExamDetailsByID 获取考试详情，包含其关联的科目和满分
func (r *Repository) GetExamDetailsByID(id uint) (*schemas.ExamDetailSchema, error) {
	var exam models.Exam
	if err := r.DB.First(&exam, id).Error; err != nil {
		return nil, err
	}
	var examSubjects []models.ExamSubject
	if err := r.DB.Where("exam_id = ?", id).Find(&examSubjects).Error; err != nil {
		return nil, err
	}
	details := &schemas.ExamDetailSchema{
		ExamSchema: schemas.ExamSchema{
			ID:       exam.ID,
			Name:     exam.Name,
			ExamDate: exam.ExamDate,
			Status:   exam.Status,
		},
		Subjects: make([]schemas.ExamSubjectDetailSchema, len(examSubjects)),
	}
	for i, es := range examSubjects {
		subjectName, ok := r.getSubjectNameByID(es.SubjectID)
		if !ok {
			return nil, fmt.Errorf("科目ID %d 未找到", es.SubjectID)
		}
		details.Subjects[i] = schemas.ExamSubjectDetailSchema{
			Name:     subjectName,
			FullMark: es.FullMark,
		}
	}
	return details, nil
}

// UpdateExamStatus 更新考试状态
func (r *Repository) UpdateExamStatus(id uint, status string) error {
	return r.DB.Model(&models.Exam{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteExamByID 删除考试，包含安全检查
func (r *Repository) DeleteExamByID(id uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var exam models.Exam
		if err := tx.Preload("Scores").First(&exam, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if exam.Status != "draft" {
			return fmt.Errorf("无法删除，考试 '%s' 已锁定或已提交分析", exam.Name)
		}
		if len(exam.Scores) > 0 {
			return fmt.Errorf("无法删除，考试 '%s' 已录入部分成绩", exam.Name)
		}
		if err := tx.Where("exam_id = ?", id).Delete(&models.ExamSubject{}).Error; err != nil {
			return err
		}
		return tx.Delete(&exam).Error
	})
}

//==============================================================================
//--- 数据加载与回写 ---
//==============================================================================

// LoadAnalysisData 为单场考试加载分析所需的所有数据
func (r *Repository) LoadAnalysisData(examID uint, scopeLevel string, scopeIDs []uint) (*types.AnalysisInput, map[uint]*types.StudentHistory, error) {
	var exam models.Exam
	if err := r.DB.First(&exam, examID).Error; err != nil {
		return nil, nil, err
	}
	var examSubjects []models.ExamSubject
	if err := r.DB.Where("exam_id = ?", exam.ID).Find(&examSubjects).Error; err != nil {
		return nil, nil, err
	}
	fullMarks := make(map[string]float64, len(examSubjects))
	for _, es := range examSubjects {
		subjectName, ok := r.getSubjectNameByID(es.SubjectID)
		if !ok {
			return nil, nil, fmt.Errorf("科目ID %d 未找到", es.SubjectID)
		}
		fullMarks[subjectName] = es.FullMark
	}
	var scores []models.Score
	if err := r.DB.Where("exam_id = ?", examID).Preload("Student.Class.Grade").Find(&scores).Error; err != nil {
		return nil, nil, err
	}

	scopeSet := make(map[uint]bool, len(scopeIDs))
	for _, id := range scopeIDs {
		scopeSet[id] = true
	}
	switch scopeLevel {
	case "FULL_EXAM":
	case "GRADE", "CLASS":
		if len(scopeSet) == 0 {
			return nil, nil, fmt.Errorf("%s 分析范围缺少目标ID", scopeLevel)
		}
	default:
		return nil, nil, fmt.Errorf("未知分析范围: %s", scopeLevel)
	}
	matchesScope := func(score models.Score) bool {
		if scopeLevel == "FULL_EXAM" {
			return true
		}
		switch scopeLevel {
		case "CLASS":
			return scopeSet[score.Student.ClassID]
		case "GRADE":
			return scopeSet[score.Student.Class.GradeID]
		default:
			return true
		}
	}

	studentIDsInScope := make(map[uint]bool)
	tablesMap := make(map[string]*types.ClassInputData)
	for _, score := range scores {
		if !matchesScope(score) {
			continue
		}
		subjectName, ok := r.getSubjectNameByID(score.SubjectID)
		if !ok {
			return nil, nil, fmt.Errorf("科目ID %d 未找到", score.SubjectID)
		}
		studentIDsInScope[score.StudentID] = true
		className := reportTableName(score.Student.Class)
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
		studentInput.Scores[subjectName] = score.Score
	}
	analysisInput := &types.AnalysisInput{
		GroupName:      exam.Name,
		FullMarks:      fullMarks,
		ExamID:         exam.ID,
		PersistMetrics: scopeLevel == "FULL_EXAM",
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

// loadHistoricalData 加载历史数据
func (r *Repository) loadHistoricalData(studentIDs map[uint]bool, currentExamDate time.Time) (map[uint]*types.StudentHistory, error) {
	if len(studentIDs) == 0 {
		return make(map[uint]*types.StudentHistory), nil
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

	tempHistoryMap := make(map[uint]map[uint]*types.HistoricalExam)

	for _, score := range historicalScores {
		examID := score.ExamID

		if _, ok := tempHistoryMap[score.StudentID]; !ok {
			tempHistoryMap[score.StudentID] = make(map[uint]*types.HistoricalExam)
		}
		if _, ok := tempHistoryMap[score.StudentID][examID]; !ok {
			tempHistoryMap[score.StudentID][examID] = &types.HistoricalExam{
				ExamName: score.Exam.Name,
				ExamDate: score.Exam.ExamDate.Format("2006-01-02"),
				Scores:   make(map[string]float64),
			}
		}
		subjectName, ok := r.getSubjectNameByID(score.SubjectID)
		if !ok {
			return nil, fmt.Errorf("科目ID %d 未找到", score.SubjectID)
		}
		tempHistoryMap[score.StudentID][examID].Scores[subjectName] = score.Score
		if score.TScore > 0 {
			// 这里简单累加，后续在组合时需要除以科目数来求平均
			tempHistoryMap[score.StudentID][examID].TotalTScore += score.TScore
		}
		if score.GradePercentileRank > 0 {
			// 同上
			tempHistoryMap[score.StudentID][examID].GradePercentileRank += score.GradePercentileRank
		}
	}

	finalHistoryMap := make(map[uint]*types.StudentHistory)
	for studentID, examsMap := range tempHistoryMap {
		studentHistory := &types.StudentHistory{AllExams: []*types.HistoricalExam{}}
		for _, examData := range examsMap {
			var totalScore float64
			for _, s := range examData.Scores {
				totalScore += s
			}
			examData.TotalScore = totalScore
			if len(examData.Scores) > 0 {
				examData.TotalTScore /= float64(len(examData.Scores))
				examData.GradePercentileRank /= float64(len(examData.Scores))
			}
			studentHistory.AllExams = append(studentHistory.AllExams, examData)
		}
		sort.Slice(studentHistory.AllExams, func(i, j int) bool {
			return studentHistory.AllExams[i].ExamDate < studentHistory.AllExams[j].ExamDate
		})
		if len(studentHistory.AllExams) > 0 {
			studentHistory.LastExam = studentHistory.AllExams[len(studentHistory.AllExams)-1]
		}
		finalHistoryMap[studentID] = studentHistory
	}

	return finalHistoryMap, nil
}

// UpdateScoresWithMetrics 异步回写分析结果
func (r *Repository) UpdateScoresWithMetrics(report *types.AnalysisReport, examID uint) error {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	log.Printf("启动指标回写任务，考试ID: %d...", examID)

	for _, table := range report.Tables {
		for _, student := range table.Students {
			for subjectName := range student.Scores.RawScores {
				subjectID, ok := r.getSubjectIDByName(subjectName)
				if !ok {
					continue
				}

				tScore := student.Scores.TScores[subjectName]
				percentileRank := student.Ranks.Subjects[subjectName].GradePercentileRank

				res := tx.Model(&models.Score{}).
					Where("student_id = ? AND exam_id = ? AND subject_id = ?", student.ID, examID, subjectID).
					// 添加一个 score = ? 的条件可以增加更新的安全性，确保我们更新的是正确的原始分数记录
					// 但如果分数可能被修改，这个条件可能导致更新失败，这里暂时去掉以保证回写的健壮性
					Updates(map[string]interface{}{
						"t_score":               tScore,
						"grade_percentile_rank": percentileRank,
					})

				if res.Error != nil {
					tx.Rollback()
					log.Printf("错误: 回写学生 %d, 科目 %d 指标失败。正在回滚事务。错误: %v", student.ID, subjectID, res.Error)
					return res.Error
				}
			}
		}
	}

	log.Println("指标回写成功。正在提交事务。")
	return tx.Commit().Error
}
