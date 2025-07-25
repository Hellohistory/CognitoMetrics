import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import type { IExam, IExamWithSubjectsCreate } from '@/types/dataModels';
import {
    getExams,
    createExam,
    deleteExam,
    unlockExam,
    finalizeExam
} from "@/api/examAPI";

export const useExamStore = defineStore('exam', () => {
  const examList = ref<IExam[]>([]);
  const selectedExamId = ref<number | undefined>(undefined);
  const isLoading = ref(false);
  const isSubmitting = ref(false); // 用于创建、删除等操作的加载状态

  /**
   * 当前选中的完整考试对象
   */
  const selectedExam = computed(() =>
    examList.value.find(e => e.id === selectedExamId.value)
  );


  /**
   * 从服务器获取考试列表
   */
  const fetchExams = async () => {
    isLoading.value = true;
    try {
      examList.value = await getExams();
    } catch (error) {
      console.error("获取考试列表失败:", error);
      examList.value = [];
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * 创建一个新考试，成功后返回新考试对象
   */
  const handleCreateExam = async (examData: IExamWithSubjectsCreate): Promise<IExam | null> => {
    isSubmitting.value = true;
    try {
      const newExam = await createExam(examData);
      await fetchExams(); // 重新加载数据以包含新考试
      return newExam; // 返回新创建的考试对象
    } catch (error) {
      // 错误已由API层统一处理
      return null;
    } finally {
      isSubmitting.value = false;
    }
  };

  /**
   * 处理删除考试的逻辑
   */
  const handleDeleteExam = async (examId: number): Promise<boolean> => {
    isSubmitting.value = true;
    try {
      await deleteExam(examId);
      if (selectedExamId.value === examId) {
        selectedExamId.value = undefined;
      }
      await fetchExams();
      return true;
    } catch (error) {
      return false;
    } finally {
      isSubmitting.value = false;
    }
  };

  /**
   * 处理解锁考试的逻辑
   */
  const handleUnlockExam = async (examId: number): Promise<boolean> => {
      isSubmitting.value = true;
      try {
          await unlockExam(examId);
          await fetchExams();
          return true;
      } catch (error) {
          return false;
      } finally {
          isSubmitting.value = false;
      }
  };

  /**
   * 处理定稿考试的逻辑
   */
  const handleFinalizeExam = async (examId: number): Promise<boolean> => {
      isSubmitting.value = true;
      try {
          await finalizeExam(examId);
          await fetchExams();
          return true;
      } catch (error) {
          return false;
      } finally {
          isSubmitting.value = false;
      }
  };

  /**
   * 设置当前选中的考试ID
   */
  const selectExam = (id: number | undefined) => {
    selectedExamId.value = id;
  }

  return {
    // state
    examList,
    selectedExamId,
    isLoading,
    isSubmitting,
    // getters
    selectedExam,
    // actions
    fetchExams,
    handleCreateExam,
    handleDeleteExam,
    handleUnlockExam,
    handleFinalizeExam,
    selectExam,
  };
});