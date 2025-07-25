// src/types/dataModels.ts
// =============================================================================
// --- 组织结构模型 (Organizational Models) ---
// =============================================================================

export interface IGradeNode {
  id: number;
  name: string;
  classes: IClassNode[];
}

export interface IGradeCreate {
  name: string;
}

export interface IGradeUpdate {
  name?: string;
}

export interface IClassNode {
  id: number;
  name: string;
  enrollment_year: number;
  student_count: number;
}

export interface IClassCreate {
  name: string;
  enrollment_year: number;
  grade_id: number;
}

export interface IClassUpdate {
  name?: string;
  enrollment_year?: number;
}


// =============================================================================
// --- 人员模型 (Personnel Models) ---
// =============================================================================

export interface IStudent {
  id: number;
  student_no: string;
  name: string;
  class_id: number;
  is_active: boolean;
}

export interface IStudentCreate {
  name: string;
  class_id: number;
  student_no?: string;
}

export interface IStudentUpdate {
  name?: string;
  class_id?: number;
  student_no?: string;
}

export interface IStudentCreateBatch {
  students: IStudentCreate[];
}

export interface IStudentDetail extends IStudent {
    grade_name: string;
    class_name: string;
    enrollment_year: number;
}

export interface IPerformanceRecord {
    exam_id: number;
    exam_name: string;
    exam_date: string;
    total_score: number | null;
    class_rank: number | null;
    grade_rank: number | null;
}

export interface IStudentPerformanceHistory {
    records: IPerformanceRecord[];
}


// =============================================================================
// --- 考试与成绩模型 (Exam & Score Models) ---
// =============================================================================

export interface ISubjectInExamCreate {
  name: string;
  full_mark: number;
}

export interface IExamWithSubjectsCreate {
  name: string;
  exam_date: string;
  subjects: ISubjectInExamCreate[];
}

export interface IExam {
  id: number;
  name: string;
  exam_date: string;
  status: 'draft' | 'submitted' | 'processing' | 'completed' | 'failed';
}

export interface IExamSubjectDetail {
  name:string;
  full_mark: number;
}

export interface IExamDetail extends IExam {
  subjects: IExamSubjectDetail[];
}

export interface IScoreInput {
  student_id: number;
  subject_scores: Record<string, number | null>;
}

export interface IScoresBatchInput {
  exam_id: number;
  scores: IScoreInput[];
}


export interface ISingleScoreUpdate {
  exam_id: number;
  student_id: number;
  subject_name: string;
  score: number | null;
}

export type IScoreRecord = IScoreInput;

export type IScoresQueryResponse = IScoreRecord[];


// =============================================================================
// --- 分析与报告模型 (Analysis & Report Models) ---
// =============================================================================

export interface IAnalysisReport {
  id: number;
  report_name: string;
  status: 'processing' | 'completed' | 'failed';
  report_type: 'single' | 'comparison';

  // 【核心修改】新增字段以支持异步AI分析
  ai_analysis_status: 'not_started' | 'processing' | 'completed' | 'failed';
  ai_analysis_cache: string | null;

  created_at: string;
  updated_at: string | null;
  exam?: IExam;
}

export interface IAnalysisScope {
  level: 'FULL_EXAM' | 'GRADE' | 'CLASS';
  ids: number[];
}

export interface IAnalysisSubmissionRequest {
  exam_id: number;
  report_name: string;
  scope: IAnalysisScope;
}

export interface IComparisonReportRequest {
  report_name?: string;
  report_ids: number[];
}

export interface IDescriptiveStats {
  count: number;
  mean: number;
  stdDev: number;
  variance: number;
  min: number;
  q1: number;
  median: number;
  q3: number;
  max: number;
  range: number;
  excellentRate: number;
  goodRate: number;
  passRate: number;
  lowScoreRate: number;
  difficulty: number;
  skewness: number;
  kurtosis: number;
  fullMarkCount: number;
  zeroMarkCount: number;
  highAchieverPenetration?: number;
  strugglerSupportIndex?: number;
  academicCoreDensity?: number;
  frequencyDistribution: Record<string, number>;
  boxPlotData: { min: number; q1: number; median: number; q3: number; max: number; };
}

export interface IStudentReportData {
  studentId: number; // 确保学生ID存在，用于对比分析
  studentName: string;
  tableName: string;
  totalScore: number;
  profile: string;
  ranks: {
    totalScore: {
      gradeRank: number;
      classRank: number;
      gradePercentileRank: number;
      classPercentileRank: number;
    };
    subjects: Record<string, {
      gradeRank: number;
      classRank: number;
    }>;
  };
  scores: {
    rawScores: Record<string, number>;
    zScores: Record<string, number>;
    tScores: Record<string, number>;
    scoreRates: Record<string, number>;
  };
  metrics: {
    imbalanceIndex: number;
    strengthSubjects: { subject: string, tScore: number }[];
    weaknessSubjects: { subject: string, tScore: number }[];
    contributionScore: Record<string, number>;
    specializationIndex: number;
    pointsToPass?: number;
    pointsToExcellent?: number;
    history?: {
        gradePercentileRankSlope?: number;
        [key: string]: any;
    }
  };
}

export interface IClassReportData {
    tableName: string;
    tableStats: Record<string, IDescriptiveStats>;
    students: IStudentReportData[];
}

export interface IFullReport {
    groupName: string;
    fullMarks: Record<string, number>;
    groupStats: Record<string, IDescriptiveStats & { correlationMatrix?: any }>;
    tables: IClassReportData[];
}

// IAnalysisReportDetail 会自动继承 IAnalysisReport 中的所有字段，包括我们新增的AI相关字段
export interface IAnalysisReportDetail extends IAnalysisReport {
  full_report_data: IFullReport | null;
  error_message?: string;
}
