/**
 * @file src/api/classApi.ts
 * @description 提供年级与班级相关的 API 调用
 * @author Hellohistory
 */

/**
 * @module api/classApi
 * 高内聚：仅负责年级/班级网络请求
 * 低耦合：依赖通用 apiClient，与业务逻辑分离
 */

import apiClient from './apiClient';
import type {
  IClassCreate,
  IGradeCreate,
  IGradeNode,
  IClassNode,
  IClassUpdate, IGradeUpdate, // 引入新类型
} from '@/types/dataModels';

/**
 * 获取用于导航的年级-班级树状数据
 */
export const getClassTree = async (): Promise<IGradeNode[]> => {
  try {
    const resp = await apiClient.get<IGradeNode[]>('/classes/tree');
    return resp.data;
  } catch (err) {
    console.error('【getClassTree】获取年级-班级树失败：', err);
    return [];
  }
};

/**
 * 创建一个新年级
 */
export const createGrade = async (
  gradeData: IGradeCreate,
): Promise<IGradeNode> => {
  try {
    const resp = await apiClient.post<IGradeNode>(
      '/grades/',
      gradeData,
    );
    return resp.data;
  } catch (err) {
    console.error('【createGrade】创建年级失败：', err);
    throw err;
  }
};

/**
 * 创建一个新班级
 */
export const createClass = async (
  classData: IClassCreate,
): Promise<IClassNode> => {
  try {
    const resp = await apiClient.post<IClassNode>(
      '/classes/',
      classData,
    );
    return resp.data;
  } catch (err) {
    console.error('【createClass】创建班级失败：', err);
    throw err;
  }
};

/**
 * @param classId - 要更新的班级ID
 * @param classData - 包含要更新字段的数据对象
 * @returns 更新后的班级节点
 * @throws 更新失败时抛出错误
 */
export const updateClass = async (
  classId: number,
  classData: IClassUpdate,
): Promise<IClassNode> => {
  try {
    const resp = await apiClient.put<IClassNode>(`/classes/${classId}`, classData);
    return resp.data;
  } catch (err) {
    console.error(`【updateClass】更新班级 ${classId} 失败：`, err);
    throw err;
  }
};

/**
 * @param classId - 要删除的班级ID
 * @throws 删除失败时抛出错误 (例如班级下有学生)
 */
export const deleteClass = async (classId: number): Promise<void> => {
  try {
    await apiClient.delete(`/classes/${classId}`);
  } catch (err) {
    console.error(`【deleteClass】删除班级 ${classId} 失败：`, err);
    throw err;
  }
};

/**
 * @param gradeId - 要删除的年级ID
 * @throws 删除失败时抛出错误 (例如年级下有学生)
 */
export const deleteGrade = async (gradeId: number): Promise<void> => {
  try {
    // 根据 main.py，挂载点是 /api，grades_router 的前缀是 /grades
    // apiClient 的 baseURL 可能是 /api，因此路径是 /grades/{id}
    await apiClient.delete(`/grades/${gradeId}`);
  } catch (err) {
    console.error(`【deleteGrade】删除年级 ${gradeId} 失败：`, err);
    throw err;
  }
};

/**
 * @param gradeId - 要更新的年级ID
 * @param gradeData - 包含要更新字段的数据对象
 * @returns 更新后的年级节点
 * @throws 更新失败时抛出错误
 */
export const updateGrade = async (
  gradeId: number,
  gradeData: IGradeUpdate,
): Promise<IGradeNode> => {
  try {
    const resp = await apiClient.put<IGradeNode>(`/grades/${gradeId}`, gradeData);
    return resp.data;
  } catch (err) {
    console.error(`【updateGrade】更新年级 ${gradeId} 失败：`, err);
    throw err;
  }
};