/**
 * @file src/api/scoreAPI.ts
 * @description 提供成绩相关的 API 调用
 */
import apiClient from './apiClient';
import type { IScoreInput, IScoresBatchInput, ISingleScoreUpdate } from '@/types/dataModels';
import type { ISubmitResponse } from './analysisApi';

/**
 * 获取指定班级在某场考试中的所有成绩
 */
export const getScoresForClass = async (examId: number, classId: number): Promise<IScoreInput[]> => {
    try {
        const response = await apiClient.get<IScoreInput[]>(`/scores/exam/${examId}/class/${classId}`);
        return response.data;
    } catch(err) {
        console.error(`获取班级 ${classId} 在考试 ${examId} 的成绩失败`, err);
        throw err;
    }
};

/**
 * 批量保存成绩 (可用于粘贴功能)
 */
export const saveScoresBatch = async (payload: IScoresBatchInput): Promise<ISubmitResponse> => {
    const response = await apiClient.post<ISubmitResponse>('/scores/batch', payload);
    return response.data;
};

/**
 * 保存单个成绩点 (用于自动保存)
 * @param payload - 必须符合 ISingleScoreUpdate 接口，其中 score 可以为 null
 */
export const saveSingleScore = async (payload: ISingleScoreUpdate): Promise<{ message: string }> => {
    try {
        const response = await apiClient.put<{ message: string }>('/scores/single', payload);
        return response.data;
    } catch (err) {
        console.error('单个成绩保存失败:', err);
        throw err;
    }
};