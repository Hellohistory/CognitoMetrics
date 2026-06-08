<template>
  <div class="grade-overview-tab">
    <el-tabs v-model="activeSubjectTab" type="card" class="subject-tabs">
      <el-tab-pane
        v-for="(stats, subjectName) in subjectStats"
        :key="subjectName"
        :label="getSubjectDisplayName(subjectName)"
        :name="String(subjectName)"
      >
        <div v-if="activeSubjectTab === String(subjectName)" class="tab-content">
          <ReportSummary :stats="stats" />
          <div class="chart-grid">
            <el-card
              class="chart-card"
              v-if="gradeLevelCharts.score_distribution_histogram?.[subjectName]"
            >
              <template #header>
                <span>{{ getSubjectDisplayName(subjectName) }} - 分数段分布直方图</span>
              </template>
              <VueEcharts
                :option="createHistogramOption(gradeLevelCharts.score_distribution_histogram[subjectName])"
              />
            </el-card>

            <el-card
              class="chart-card"
              v-if="subjectName === 'totalScore' && gradeLevelCharts.subject_correlation_heatmap"
            >
              <template #header>
                <span>学科相关性热力图</span>
              </template>
              <VueEcharts
                :option="createHeatmapOption(gradeLevelCharts.subject_correlation_heatmap)"
              />
            </el-card>

            <el-card
              class="chart-card"
              v-if="subjectName === 'totalScore' && gradeLevelCharts.subject_difficulty_discrimination_scatter"
            >
              <template #header>
                <span>学科难度-区分度诊断</span>
              </template>
              <VueEcharts
                :option="createDifficultyDiscriminationScatterOption(gradeLevelCharts.subject_difficulty_discrimination_scatter)"
              />
            </el-card>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    <el-empty v-if="Object.keys(subjectStats).length === 0" description="暂无年级总览数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watchEffect } from 'vue';
import { ElTabs, ElTabPane, ElCard, ElEmpty } from 'element-plus';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';

import VueEcharts from '@/components/charts/VueEcharts.vue';
import ReportSummary from '@/components/reports/ReportSummary.vue';
import {
  createHistogramOption,
  createHeatmapOption,
  createDifficultyDiscriminationScatterOption
} from '@/utils/chartOptionFactory';

const reportStore = useAnalysisReportStore();
const activeSubjectTab = ref('totalScore');

const subjectStats = computed(() => reportStore.subjectStats);
const gradeLevelCharts = computed(() => reportStore.chartData?.grade_level_charts || {});

watchEffect(() => {
  const keys = Object.keys(subjectStats.value);
  if (keys.length > 0 && !keys.includes(activeSubjectTab.value)) {
    activeSubjectTab.value = keys.includes('totalScore') ? 'totalScore' : keys[0];
  }
});

const getSubjectDisplayName = (subjectKey: string | number): string => {
  const keyStr = String(subjectKey);
  return keyStr === 'totalScore' ? '总分' : keyStr;
};
</script>

<style scoped>
.grade-overview-tab {
  min-width: 0;
}

.subject-tabs {
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius);
  box-shadow: none;
  overflow: hidden;
}

.tab-content {
  padding: 18px;
}

.chart-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
    gap: 1.5rem;
    margin-top: 18px;
}

.chart-card {
    border-radius: 8px;
    height: 400px;
    display: flex;
    flex-direction: column;
}

:deep(.chart-card .el-card__header) {
  font-weight: 600;
}

:deep(.chart-card .el-card__body) {
  flex-grow: 1;
  height: 100%;
  width: 100%;
  padding: 10px;
}

@media (max-width: 560px) {
  .tab-content {
    padding: 12px;
  }
  .chart-grid {
    grid-template-columns: 1fr;
  }
}
</style>
