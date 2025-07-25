<template>
  <div class="report-view" v-loading="reportStore.isLoading">
    <template v-if="reportStore.fullReport && reportStore.chartData">
      <div class="report-header">
        <h1>{{ reportStore.fullReport.groupName }} - 学情分析报告</h1>
        <p>报告生成时间: {{ formatDateTime(reportStore.reportDetail?.created_at) }}</p>
      </div>

      <el-tabs v-model="activeMainTab" type="border-card" class="main-tabs">

        <el-tab-pane label="班级横向对比" name="comparison">
          <ClassComparisonTab />
        </el-tab-pane>

        <el-tab-pane label="班级深度诊断" name="diagnostics">
          <ClassDiagnosticsTab />
        </el-tab-pane>

        <el-tab-pane label="学生列表" name="roster">
           <StudentDetailListTab />
        </el-tab-pane>

        <el-tab-pane label="AI 智能分析" name="ai-analysis">
           <AiAnalysisReport :report-id="parseInt(props.id)" />
        </el-tab-pane>
      </el-tabs>

    </template>

    <el-result
      v-else-if="reportStore.error"
      status="error"
      title="报告加载失败"
      :sub-title="reportStore.error"
    >
      <template #extra>
        <el-button type="primary" @click="fetchData">重试</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { ElTabs, ElTabPane, ElResult, ElButton, vLoading } from 'element-plus';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';

import ClassComparisonTab from '@/components/reports/tabs/ClassComparisonTab.vue';
import ClassDiagnosticsTab from '@/components/reports/tabs/ClassDiagnosticsTab.vue';
import StudentDetailListTab from '@/components/reports/tabs/StudentDetailListTab.vue';
import AiAnalysisReport from '@/components/reports/AiAnalysisReport.vue';


const props = defineProps<{ id: string }>();

const reportStore = useAnalysisReportStore();
const activeMainTab = ref('summary');

const fetchData = () => {
  const reportId = parseInt(props.id, 10);
  if (!isNaN(reportId)) {
    reportStore.fetchReport(reportId);
  }
};

onMounted(fetchData);

onUnmounted(() => {
  reportStore.clearReport();
});

const formatDateTime = (dateString?: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString('zh-CN');
};
</script>

<style scoped>
.report-view {
  padding: 1rem 2rem;
  background-color: #f9fafc;
}
.report-header {
  margin-bottom: 1.5rem;
}
.report-header h1 {
  font-size: 1.8rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}
.report-header p {
  color: #888;
  font-size: 0.9rem;
}
.main-tabs {
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  box-shadow: none;
}
:deep(.el-tabs__content) {
  padding: 1.5rem;
}
</style>