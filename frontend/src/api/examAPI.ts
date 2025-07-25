/**
 * @file src/api/examAPI.ts
 * @description 提供考试相关的 API 调用，包括列表查询、详情获取与创建操作
 * @author Hellohistory
 */

/**
 * @module api/examAPI
 * 高内聚：仅负责考试资源的网络请求
 * 低耦合：依赖通用 apiClient，与业务逻辑分离
 */

import apiClient from './apiClient';
import type {
  IExam,
  IExamWithSubjectsCreate,
  IExamDetail,
} from '@/types/dataModels';

/**
 * 获取所有考试的列表（包含状态）
 * @returns Promise&lt;IExam[]&gt; - 成功返回考试数组，失败返回空数组
 * @example
 * ```ts
 * const exams = await getExams();
 * ```
 */
export const getExams = async (): Promise<IExam[]> => {
  try {
    const resp = await apiClient.get<IExam[]>('/exams/');
    return resp.data;
  } catch (err) {
    console.error('【getExams】获取考试列表失败：', err);
    return [];
  }
};

/**
 * 获取单场考试的详细信息，包含科目定义
 * @param id - 考试 ID
 * @returns Promise&lt;IExamDetail&gt; - 考试详细信息
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const detail = await getExamDetails(42);
 * ```
 */
export const getExamDetails = async (
  id: number,
): Promise<IExamDetail> => {
  try {
    const resp = await apiClient.get<IExamDetail>(`/exams/${id}`);
    return resp.data;
  } catch (err) {
    console.error(`【getExamDetails】获取考试 ${id} 详情失败：`, err);
    throw err;
  }
};

/**
 * 创建一场新的考试（包含科目定义），初始状态为 'draft'
 * @param examData - 包含考试名称、日期和科目列表的数据
 * @returns Promise&lt;IExam&gt; - 创建成功的考试信息
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const exam = await createExam({
 *   name: '期末考试',
 *   date: '2025-06-15',
 *   subjects: [{ name: '数学' }, { name: '语文' }],
 * });
 * ```
 */
export const createExam = async (
  examData: IExamWithSubjectsCreate,
): Promise<IExam> => {
  try {
    const resp = await apiClient.post<IExam>('/exams/', examData);
    return resp.data;
  } catch (err) {
    console.error('【createExam】创建考试失败：', err);
    throw err;
  }
};

/**
 * @param examId - 要解锁的考试ID
 * @returns Promise<IExam> - 更新后的考试信息
 * @throws 请求失败时抛出错误
 */
export const unlockExam = async (examId: number): Promise<IExam> => {
  try {
    // 根据 main.py，挂载点是 /api, exams_router 的前缀是 /exams
    const resp = await apiClient.put<IExam>(`/exams/${examId}/unlock`);
    return resp.data;
  } catch (err) {
    console.error(`【unlockExam】解锁考试 ${examId} 失败：`, err);
    throw err;
  }
};

/**
 * @param examId - 要定稿的考试ID
 * @returns Promise<IExam> - 更新后的考试信息
 * @throws 请求失败时抛出错误
 */
export const finalizeExam = async (examId: number): Promise<IExam> => {
  try {
    // 假设后端有一个 PUT /exams/{id}/finalize 的接口来处理定稿
    const resp = await apiClient.put<IExam>(`/exams/${examId}/finalize`);
    return resp.data;
  } catch (err) {
    console.error(`【finalizeExam】定稿考试 ${examId} 失败：`, err);
    throw err;
  }
};

/**
 * @param examId - 要删除的考试ID
 * @throws 删除失败时抛出错误 (例如考试不符合删除条件)
 */
export const deleteExam = async (examId: number): Promise<void> => {
  try {
    await apiClient.delete(`/exams/${examId}`);
  } catch (err) {
    console.error(`【deleteExam】删除考试 ${examId} 失败：`, err);
    throw err;
  }
};