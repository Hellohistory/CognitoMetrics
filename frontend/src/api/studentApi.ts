/**
 * @file src/api/studentApi.ts
 * @description 提供学生相关的 API 调用，包括学生查询、更新、批量操作等
 * @author Hellohistory
 */

/**
 * @module api/studentApi
 * 高内聚：仅负责学生模块的网络请求
 * 低耦合：依赖通用 apiClient，与业务逻辑分离
 */

import apiClient from './apiClient';
import type {
  IStudent,
  IStudentUpdate,
  IStudentCreateBatch,
  IStudentDetail,
  IStudentPerformanceHistory,
} from '@/types/dataModels';

/**
 * 获取指定班级的学生列表
 * @param classId - 班级 ID
 * @param includeInactive - 是否包含已停用学生，默认 false
 * @returns Promise<IStudent[]> - 学生数组
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const students = await getStudentsByClass(1, true);
 * ```
 */
export const getStudentsByClass = async (
  classId: number,
  includeInactive: boolean = false,
): Promise<IStudent[]> => {
  try {
    const response = await apiClient.get<IStudent[]>(
      `/students/by_class/${classId}`,
      { params: { include_inactive: includeInactive } },
    );
    return response.data;
  } catch (error) {
    console.error(`【getStudentsByClass】获取班级 ${classId} 的学生列表失败：`, error);
    throw error;
  }
};

/**
 * 更新学生信息（可用于转班）
 * @param studentId - 学生 ID
 * @param studentData - 更新数据
 * @returns Promise<IStudent> - 更新后的学生数据
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const updated = await updateStudent(1, { name: '李四', classId: 2 });
 * ```
 */
export const updateStudent = async (
  studentId: number,
  studentData: IStudentUpdate,
): Promise<IStudent> => {
  try {
    const response = await apiClient.put<IStudent>(
      `/students/${studentId}`,
      studentData,
    );
    return response.data;
  } catch (error) {
    console.error(`【updateStudent】更新学生 ${studentId} 信息失败：`, error);
    throw error;
  }
};

/**
 * 停用学生账户
 * @param studentId - 学生 ID
 * @returns Promise<IStudent> - 停用后的学生数据
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const deactivated = await deactivateStudent(1);
 * ```
 */
export const deactivateStudent = async (
  studentId: number,
): Promise<IStudent> => {
  try {
    const response = await apiClient.put<IStudent>(
      `/students/${studentId}/deactivate`,
    );
    return response.data;
  } catch (error) {
    console.error(`【deactivateStudent】停用学生 ${studentId} 失败：`, error);
    throw error;
  }
};

/**
 * 激活学生账户
 * @param studentId - 学生 ID
 * @returns Promise<IStudent> - 激活后的学生数据
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const activated = await activateStudent(1);
 * ```
 */
export const activateStudent = async (
  studentId: number,
): Promise<IStudent> => {
  try {
    const response = await apiClient.put<IStudent>(
      `/students/${studentId}/activate`,
    );
    return response.data;
  } catch (error) {
    console.error(`【activateStudent】激活学生 ${studentId} 失败：`, error);
    throw error;
  }
};

/**
 * 批量创建学生
 * @param batchData - 批量创建所需数据，包括学生列表
 * @returns Promise<IStudent[]> - 创建成功的学生列表
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const newStudents = await batchCreateStudents({
 *   students: [
 *     { name: '张三', class_id: 1 },
 *     { name: '李四', class_id: 2 },
 *   ],
 * });
 * ```
 */
export const batchCreateStudents = async (
  batchData: IStudentCreateBatch,
): Promise<IStudent[]> => {
  try {
    const payload = {
      students: batchData.students.map(({ name, class_id }) => ({ name, class_id })),
    };
    const response = await apiClient.post<IStudent[]>(
      '/students/batch',
      payload,
    );
    return response.data;
  } catch (error) {
    console.error('【batchCreateStudents】批量创建学生失败：', error);
    throw error;
  }
};

/**
 * 批量转移学生到新班级
 * @param studentIds - 学生 ID 列表
 * @param targetClassId - 目标班级 ID
 * @returns Promise<{ message: string }> - 操作结果
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const res = await batchTransferStudents([1, 2, 3], 5);
 * ```
 */
export const batchTransferStudents = async (
  studentIds: number[],
  targetClassId: number,
): Promise<{ message: string }> => {
  try {
    const response = await apiClient.post<{ message: string }>(
      '/students/batch-update-class',
      { student_ids: studentIds, target_class_id: targetClassId },
    );
    return response.data;
  } catch (error) {
    console.error('【batchTransferStudents】批量转移学生失败：', error);
    throw error;
  }
};

/**
 * 批量更新学生在读状态
 * @param studentIds - 学生 ID 列表
 * @param isActive - 目标在读状态
 * @returns Promise<{ message: string }> - 操作结果
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const res = await batchUpdateStatus([1, 2], true);
 * ```
 */
export const batchUpdateStatus = async (
  studentIds: number[],
  isActive: boolean,
): Promise<{ message: string }> => {
  try {
    const response = await apiClient.post<{ message: string }>(
      '/students/batch-update-status',
      { student_ids: studentIds, is_active: isActive },
    );
    return response.data;
  } catch (error) {
    console.error('【batchUpdateStatus】批量更新学生状态失败：', error);
    throw error;
  }
};

/**
 * 获取单个学生的详细信息
 * @param studentId - 学生 ID
 * @returns Promise<IStudentDetail> - 学生详细信息
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const detail = await getStudentDetails(1);
 * ```
 */
export const getStudentDetails = async (
  studentId: number,
): Promise<IStudentDetail> => {
  try {
    const response = await apiClient.get<IStudentDetail>(
      `/students/${studentId}/details`,
    );
    return response.data;
  } catch (error) {
    console.error(`【getStudentDetails】获取学生 ${studentId} 详情失败：`, error);
    throw error;
  }
};

/**
 * 获取学生的历次成绩与表现记录
 * @param studentId - 学生 ID
 * @returns Promise<IStudentPerformanceHistory> - 历次表现记录
 * @throws Error - 请求失败时抛出错误
 * @example
 * ```ts
 * const history = await getStudentPerformanceHistory(1);
 * ```
 */
export const getStudentPerformanceHistory = async (
  studentId: number,
): Promise<IStudentPerformanceHistory> => {
  try {
    const response = await apiClient.get<IStudentPerformanceHistory>(
      `/students/${studentId}/performance`,
    );
    return response.data;
  } catch (error) {
    console.error(
      `【getStudentPerformanceHistory】获取学生 ${studentId} 表现历史失败：`,
      error,
    );
    throw error;
  }
};
