<template>
  <div class="report-view" v-loading="reportStore.isLoading">
    <template v-if="reportStore.fullReport && reportStore.chartData">
      <div class="report-header">
        <div class="report-title-block">
          <p class="page-kicker">Analysis Report</p>
          <h1>{{ reportStore.fullReport.groupName }} 学情分析报告</h1>
          <p>生成时间：{{ formatDateTime(reportStore.reportDetail?.created_at) }}</p>
        </div>
        <div class="report-summary-strip">
          <div class="mini-stat">
            <span>班级</span>
            <strong>{{ reportStore.fullReport.tables.length }}</strong>
          </div>
          <div class="mini-stat">
            <span>学生</span>
            <strong>{{ studentCount }}</strong>
          </div>
          <div class="mini-stat">
            <span>科目</span>
            <strong>{{ subjectCount }}</strong>
          </div>
        </div>
      </div>

      <el-tabs v-model="activeMainTab" type="border-card" class="main-tabs">
        <el-tab-pane name="overview">
          <template #label>
            <span class="tab-label"><el-icon><DataBoard /></el-icon> 年级总览</span>
          </template>
          <GradeOverviewTab />
        </el-tab-pane>

        <el-tab-pane name="comparison">
          <template #label>
            <span class="tab-label"><el-icon><TrendCharts /></el-icon> 班级横向对比</span>
          </template>
          <ClassComparisonTab />
        </el-tab-pane>

        <el-tab-pane name="diagnostics">
          <template #label>
            <span class="tab-label"><el-icon><DataAnalysis /></el-icon> 班级深度诊断</span>
          </template>
          <ClassDiagnosticsTab />
        </el-tab-pane>

        <el-tab-pane name="roster">
          <template #label>
            <span class="tab-label"><el-icon><UserFilled /></el-icon> 学生列表</span>
          </template>
           <StudentDetailListTab />
        </el-tab-pane>

        <el-tab-pane name="ai-analysis">
          <template #label>
            <span class="tab-label"><el-icon><MagicStick /></el-icon> AI 智能分析</span>
          </template>
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
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { ElIcon, ElTabs, ElTabPane, ElResult, ElButton, vLoading } from 'element-plus';
import { DataAnalysis, DataBoard, MagicStick, TrendCharts, UserFilled } from '@element-plus/icons-vue';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';

import GradeOverviewTab from '@/components/reports/tabs/GradeOverviewTab.vue';
import ClassComparisonTab from '@/components/reports/tabs/ClassComparisonTab.vue';
import ClassDiagnosticsTab from '@/components/reports/tabs/ClassDiagnosticsTab.vue';
import StudentDetailListTab from '@/components/reports/tabs/StudentDetailListTab.vue';
import AiAnalysisReport from '@/components/reports/AiAnalysisReport.vue';


const props = defineProps<{ id: string }>();

const reportStore = useAnalysisReportStore();
const activeMainTab = ref('overview');

const studentCount = computed(() =>
  reportStore.fullReport?.tables.reduce((sum, table) => sum + (table.students?.length || 0), 0) || 0
);
const subjectCount = computed(() => Object.keys(reportStore.fullReport?.fullMarks || {}).length);

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
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  min-height: calc(100vh - var(--app-header-height) - 48px);
  overflow-x: hidden;
}
.report-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  min-width: 0;
  max-width: 100%;
  padding: 20px;
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius);
  background: var(--app-surface);
  box-shadow: var(--app-shadow-soft);
}
.report-title-block {
  min-width: 0;
}
.report-header h1 {
  margin: 0;
  color: var(--app-text);
  font-family: "Noto Serif SC", "Source Han Sans SC", serif;
  font-size: 28px;
  font-weight: 700;
  line-height: 1.28;
  overflow-wrap: anywhere;
}
.report-header p {
  margin: 8px 0 0;
  color: var(--app-text-muted);
  font-size: 14px;
}
.report-summary-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(80px, 1fr));
  gap: 10px;
  min-width: 300px;
  max-width: 100%;
}
.mini-stat {
  min-width: 0;
  padding: 12px;
  border: 1px solid var(--app-border);
  border-radius: 7px;
  background: var(--app-surface-soft);
}
.mini-stat span,
.mini-stat strong {
  display: block;
}
.mini-stat span {
  color: var(--app-text-muted);
  font-size: 12px;
}
.mini-stat strong {
  margin-top: 4px;
  color: var(--app-text);
  font-size: 24px;
}
.main-tabs {
  min-width: 0;
  border-radius: 8px;
  border: 1px solid var(--app-border);
  box-shadow: none;
  overflow: hidden;
}
.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 700;
  white-space: nowrap;
}
:deep(.el-tabs__content) {
  padding: 18px;
  min-width: 0;
  overflow: hidden;
}

@media (max-width: 860px) {
  .report-header {
    flex-direction: column;
    padding: 18px;
  }
  .report-summary-strip {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    width: 100%;
    min-width: 0;
  }
  .report-header h1 {
    font-size: 25px;
  }
}

@media (max-width: 560px) {
  .report-header {
    padding: 18px 20px;
  }
  .report-summary-strip {
    gap: 8px;
  }
  .mini-stat {
    padding: 10px;
  }
  .mini-stat strong {
    font-size: 22px;
  }
  .main-tabs :deep(.el-tabs__header) {
    overflow-x: auto;
    overflow-y: hidden;
  }
  .main-tabs :deep(.el-tabs__nav-scroll) {
    overflow-x: auto;
  }
  .main-tabs :deep(.el-tabs__nav) {
    min-width: max-content;
  }
  :deep(.el-tabs__content) {
    padding: 12px;
  }
}
</style>
