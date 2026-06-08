package repository

import (
	"path/filepath"
	"strings"
	"testing"

	"CognitoMetrics/internal/schemas"
)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	repo, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	sqlDB, err := repo.DB.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	return repo
}

func TestLoadAnalysisDataSeparatesDuplicateClassNamesAcrossGrades(t *testing.T) {
	repo := newTestRepository(t)

	gradeA, err := repo.CreateGrade(schemas.GradeCreate{GradeBase: schemas.GradeBase{Name: "Grade A"}})
	if err != nil {
		t.Fatalf("CreateGrade A error = %v", err)
	}
	gradeB, err := repo.CreateGrade(schemas.GradeCreate{GradeBase: schemas.GradeBase{Name: "Grade B"}})
	if err != nil {
		t.Fatalf("CreateGrade B error = %v", err)
	}

	classA, err := repo.CreateClass(schemas.ClassCreate{
		ClassBase: schemas.ClassBase{Name: "Class 1", EnrollmentYear: 2026},
		GradeID:   gradeA.ID,
	})
	if err != nil {
		t.Fatalf("CreateClass A error = %v", err)
	}
	classB, err := repo.CreateClass(schemas.ClassCreate{
		ClassBase: schemas.ClassBase{Name: "Class 1", EnrollmentYear: 2027},
		GradeID:   gradeB.ID,
	})
	if err != nil {
		t.Fatalf("CreateClass B error = %v", err)
	}

	students, err := repo.BatchCreateStudents([]schemas.StudentCreate{
		{Name: "Alice", ClassID: classA.ID},
		{Name: "Bob", ClassID: classB.ID},
	}, []string{"20260001", "20270001"})
	if err != nil {
		t.Fatalf("BatchCreateStudents error = %v", err)
	}

	exam, err := repo.CreateExamWithSubjects(schemas.ExamWithSubjectsCreate{
		Name:     "Midterm",
		ExamDate: "2026-05-01",
		Subjects: []schemas.SubjectInExamCreate{
			{Name: "Math", FullMark: 100},
		},
	})
	if err != nil {
		t.Fatalf("CreateExamWithSubjects error = %v", err)
	}

	scoreA := 90.0
	scoreB := 80.0
	for i, score := range []*float64{&scoreA, &scoreB} {
		err := repo.UpsertSingleScore(schemas.SingleScoreUpdate{
			ExamID:      exam.ID,
			StudentID:   students[i].ID,
			SubjectName: "Math",
			Score:       score,
		})
		if err != nil {
			t.Fatalf("UpsertSingleScore[%d] error = %v", i, err)
		}
	}

	input, _, err := repo.LoadAnalysisData(exam.ID, "FULL_EXAM", nil)
	if err != nil {
		t.Fatalf("LoadAnalysisData error = %v", err)
	}
	if !input.PersistMetrics {
		t.Fatal("FULL_EXAM analysis should persist score metrics")
	}
	if got, want := len(input.Tables), 2; got != want {
		t.Fatalf("table count = %d, want %d", got, want)
	}

	studentsByTable := make(map[string]int)
	for _, table := range input.Tables {
		studentsByTable[table.TableName] = len(table.Students)
	}
	for _, name := range []string{"Grade A-Class 1", "Grade B-Class 1"} {
		if got := studentsByTable[name]; got != 1 {
			t.Fatalf("students in table %q = %d, want 1; all tables = %#v", name, got, studentsByTable)
		}
	}

	classInput, _, err := repo.LoadAnalysisData(exam.ID, "CLASS", []uint{classA.ID})
	if err != nil {
		t.Fatalf("LoadAnalysisData CLASS error = %v", err)
	}
	if classInput.PersistMetrics {
		t.Fatal("CLASS analysis should not persist score metrics")
	}
	if got, want := len(classInput.Tables), 1; got != want {
		t.Fatalf("CLASS table count = %d, want %d", got, want)
	}
	if got, want := classInput.Tables[0].TableName, "Grade A-Class 1"; got != want {
		t.Fatalf("CLASS table name = %q, want %q", got, want)
	}
}

func TestLoadAnalysisDataRejectsEmptyScopedIDs(t *testing.T) {
	repo := newTestRepository(t)

	exam, err := repo.CreateExamWithSubjects(schemas.ExamWithSubjectsCreate{
		Name:     "Midterm",
		ExamDate: "2026-05-01",
		Subjects: []schemas.SubjectInExamCreate{
			{Name: "Math", FullMark: 100},
		},
	})
	if err != nil {
		t.Fatalf("CreateExamWithSubjects error = %v", err)
	}

	_, _, err = repo.LoadAnalysisData(exam.ID, "CLASS", nil)
	if err == nil {
		t.Fatal("LoadAnalysisData with empty CLASS scope returned nil error")
	}
	if !strings.Contains(err.Error(), "缺少目标ID") {
		t.Fatalf("error = %q, want missing target ID", err.Error())
	}
}
