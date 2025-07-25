// src/stores/analysisReportStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { getReportDetails, getChartDataForReport } from '@/api/analysisApi';
import type { IAnalysisReportDetail, IFullReport } from '@/types/dataModels';

export const useAnalysisReportStore = defineStore('analysisReport', () => {
  // --- State ---
  const isLoading = ref<boolean>(true);
  const reportDetail = ref<IAnalysisReportDetail | null>(null);
  const chartData = ref<any | null>(null);
  const error = ref<string | null>(null);

  // --- Getters ---
  const fullReport = computed<IFullReport | null>(() => reportDetail.value?.full_report_data || null);
  const subjectStats = computed(() => {
    if (!fullReport.value?.groupStats) return {};
    const stats = { ...fullReport.value.groupStats };
    delete (stats as any).correlationMatrix; // 不在表格中展示
    return stats;
  });
  const allStudents = computed(() => {
    if (!fullReport.value?.tables) return [];
    return fullReport.value.tables.flatMap(table => table.students || []);
  });
  const classNames = computed(() => fullReport.value?.tables?.map(t => t.tableName) || []);

  // --- Actions ---
  async function fetchReport(reportId: number) {
    isLoading.value = true;
    error.value = null;
    reportDetail.value = null;
    chartData.value = null;

    try {
      // 并行获取报告主数据和图表数据
      const [details, charts] = await Promise.all([
        getReportDetails(reportId),
        getChartDataForReport(reportId)
      ]);

      if (details.status === 'completed' && details.full_report_data) {
        reportDetail.value = details;
        chartData.value = charts;
      } else {
        error.value = details.error_message || `报告状态异常: ${details.status}`;
      }
    } catch (e: any) {
      error.value = e.message || '获取报告数据失败';
      console.error(e);
    } finally {
      isLoading.value = false;
    }
  }

  function clearReport() {
    reportDetail.value = null;
    chartData.value = null;
    error.value = null;
  }

  return {
    // State
    isLoading,
    reportDetail,
    chartData,
    error,
    // Getters
    fullReport,
    subjectStats,
    allStudents,
    classNames,
    // Actions
    fetchReport,
    clearReport
  };
});