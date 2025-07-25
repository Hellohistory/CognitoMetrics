// src/api/analysisApi.ts

import apiClient from './apiClient';
import type {
  IAnalysisReport,
  IAnalysisReportDetail,
  IAnalysisSubmissionRequest,
  IComparisonReportRequest,
  IFullReport,
  IClassReportData,
  IStudentReportData,
} from '@/types/dataModels';

export interface IGetReportsParams {
  page?: number;
  pageSize?: number;
  query?: string;
  status?: string;
  report_type?: string;
}

export interface IPaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ISubmitResponse {
  message: string;
  report_id: number;
}

export interface ISubmitAiResponse extends ISubmitResponse {
  ai_analysis_status: string;
  analysis?: string | null; // 当任务已完成时，可能直接返回结果
}


const BASE_PATH = '/analysis';

export const getReports = async (
  params: IGetReportsParams,
): Promise<IPaginatedResponse<IAnalysisReport>> => {
  const response = await apiClient.get<IPaginatedResponse<IAnalysisReport>>(
    `${BASE_PATH}/reports`,
    { params },
  );
  return response.data;
};

export const getReportDetails = async (
  id: number,
): Promise<IAnalysisReportDetail> => {
  const response = await apiClient.get<IAnalysisReportDetail>(
    `${BASE_PATH}/reports/${id}`,
  );
  return response.data;
};

export const submitAnalysisJob = async (
  submissionData: IAnalysisSubmissionRequest,
): Promise<ISubmitResponse> => {
  const response = await apiClient.post<ISubmitResponse>(
    `${BASE_PATH}/submit`,
    submissionData,
  );
  return response.data;
};

export const submitComparisonJob = async (
  submissionData: IComparisonReportRequest,
): Promise<ISubmitResponse> => {
  const response = await apiClient.post<ISubmitResponse>(
    `${BASE_PATH}/compare`,
    submissionData,
  );
  return response.data;
};

export const deleteReport = async (id: number): Promise<void> => {
  await apiClient.delete(`${BASE_PATH}/reports/${id}`);
};

export const retryAnalysis = async (id: number): Promise<ISubmitResponse> => {
  const response = await apiClient.post<ISubmitResponse>(`${BASE_PATH}/reports/${id}/retry`);
  return response.data;
};

/**
 * 【修改】重命名函数并更新返回类型，以匹配异步任务提交的API。
 * @param reportId - 报告 ID
 * @returns Promise<ISubmitAiResponse>
 */
export const submitAiAnalysisTask = async (reportId: number): Promise<ISubmitAiResponse> => {
  const response = await apiClient.post<ISubmitAiResponse>(`${BASE_PATH}/reports/${reportId}/ai-analysis`);
  return response.data;
};


export const getGroupStatsFromReport = async (reportId: number): Promise<IFullReport['groupStats']> => {
    const response = await apiClient.get(`${BASE_PATH}/reports/${reportId}/group-stats`);
    return response.data;
};

export const getClassReportFromReport = async (reportId: number, className: string): Promise<IClassReportData> => {
    const response = await apiClient.get(`${BASE_PATH}/reports/${reportId}/class/${className}`);
    return response.data;
};

export const getStudentReportFromReport = async (reportId: number, studentName: string): Promise<IStudentReportData> => {
    const response = await apiClient.get(`${BASE_PATH}/reports/${reportId}/student/${studentName}`);
    return response.data;
};

export const getChartDataForReport = async (reportId: number): Promise<any> => {
    const response = await apiClient.get(`${BASE_PATH}/reports/${reportId}/charts`);
    return response.data;
};