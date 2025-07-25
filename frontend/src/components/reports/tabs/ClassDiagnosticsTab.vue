<template>
  <div class="class-diagnostics-tab">
    <div class="class-selector-bar">
      <span class="toolbar-label">选择诊断班级:</span>
      <el-select v-model="selectedClassName" placeholder="请选择班级" style="width: 200px;">
        <el-option v-for="name in classNames" :key="name" :label="name" :value="name" />
      </el-select>
    </div>

    <div v-if="selectedClassReport" class="diagnostics-content">
      <el-row :gutter="20" class="section-row">
        <el-col :span="12">
          <el-card shadow="never" class="content-card">
            <template #header>班级学科画像 (vs. 年级)</template>
            <VueEcharts v-if="radarChartOption" :option="radarChartOption" style="height: 350px;" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never" class="content-card">
            <template #header>总分分布对比 (班级 vs. 年级)</template>
            <VueEcharts v-if="overlayHistogramOption" :option="overlayHistogramOption" style="height: 350px;" />
          </el-card>
        </el-col>
      </el-row>

      <el-row class="section-row">
        <el-col :span="24">
          <el-card shadow="never" class="content-card">
            <template #header>核心指标详情 (班级 vs. 年级)</template>
            <el-table :data="metricsTableData" border stripe>
              <el-table-column prop="metricName" label="核心指标" width="160" fixed />
              <el-table-column
                v-for="subject in subjectKeys"
                :key="subject"
                :label="subject === 'totalScore' ? '总分' : subject"
                align="center"
              >
                <template #default="{ row }">
                  <div class="metric-cell">
                    <span class="class-value">{{ row.values[subject]?.classValue ?? 'N/A' }}</span>
                    <span class="grade-value">/ {{ row.values[subject]?.gradeValue ?? 'N/A' }}</span>
                    <span :class="getDiffClass(row.values[subject]?.diff)">{{ formatDiff(row.values[subject]?.diff) }}</span>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <el-row class="section-row">
        <el-col :span="24">
          <el-card shadow="never" class="content-card">
             <template #header>班级学生名单 (按班级排名)</template>
             <el-table :data="studentRosterData" border stripe height="500px">
                <el-table-column type="index" label="#" width="55" />
                <el-table-column prop="studentName" label="姓名" width="110" />
                <el-table-column prop="totalScore" label="总分" sortable />
                <el-table-column prop="ranks.totalScore.classRank" label="班级排名" sortable />
                <el-table-column prop="ranks.totalScore.gradeRank" label="年级排名" sortable />
                <el-table-column prop="profile" label="画像" />
                <el-table-column prop="scores.tScores.totalScore" label="总分T-Score" sortable />
             </el-table>
          </el-card>
        </el-col>
      </el-row>

    </div>
    <el-empty v-else description="请先从上方选择一个班级进行诊断" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watchEffect } from 'vue';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';
import { ElSelect, ElOption, ElCard, ElRow, ElCol, ElTable, ElTableColumn, ElEmpty } from 'element-plus';
import VueEcharts from '@/components/charts/VueEcharts.vue';
import { createClassRadarOption } from '@/utils/chartOptionFactory';
import type { EChartsOption } from 'echarts';

const store = useAnalysisReportStore();

const classNames = computed(() => store.classNames);
const selectedClassName = ref('');

// 当报告数据加载后，默认选中第一个班级
watchEffect(() => {
  if (classNames.value.length > 0 && !selectedClassName.value) {
    selectedClassName.value = classNames.value[0];
  }
});

const fullReport = computed(() => store.fullReport);
const chartData = computed(() => store.chartData);

// 根据选择的班级，筛选出该班的完整报告
const selectedClassReport = computed(() => {
  if (!selectedClassName.value || !fullReport.value) return null;
  return fullReport.value.tables.find(t => t.tableName === selectedClassName.value) || null;
});

// --- 图表计算属性 ---

// 雷达图
const radarChartOption = computed<EChartsOption | null>(() => {
    if (!selectedClassName.value || !chartData.value) return null;
    const radarData = chartData.value?.class_comparison_charts?.class_profile_radar?.[selectedClassName.value];
    if (!radarData) return null;
    return createClassRadarOption(radarData);
});

// 覆盖直方图
const overlayHistogramOption = computed<EChartsOption | null>(() => {
    if (!selectedClassReport.value || !fullReport.value?.groupStats) return null;

    const classDist = selectedClassReport.value.tableStats.totalScore.frequencyDistribution;
    const gradeDist = fullReport.value.groupStats.totalScore.frequencyDistribution;
    const categories = Object.keys(gradeDist);

    return {
        tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
        legend: { data: [`${selectedClassName.value}分布`, '年级分布'], bottom: 0 },
        grid: { top: '15%', left: '3%', right: '4%', bottom: '12%', containLabel: true },
        xAxis: [{ type: 'category', data: categories, axisTick: { alignWithLabel: true } }],
        yAxis: [{ type: 'value', name: '人数' }],
        series: [
            {
                name: '年级分布',
                type: 'bar',
                barWidth: '60%',
                data: Object.values(gradeDist),
                itemStyle: { color: '#D1E9FF' } // 年级作为浅色背景
            },
            {
                name: `${selectedClassName.value}分布`,
                type: 'bar',
                barWidth: '35%',
                data: Object.values(classDist),
                itemStyle: { color: '#5470c6' } // 班级用深色突出
            }
        ]
    };
});

// --- 表格计算属性 ---

const subjectKeys = computed(() => {
    if (!fullReport.value?.groupStats) return [];
    return Object.keys(fullReport.value.groupStats).filter(key => key !== 'correlationMatrix');
});

const metricsToShow = [
    { key: 'mean', name: '平均分', isRate: false },
    { key: 'passRate', name: '及格率', isRate: true },
    { key: 'excellentRate', name: '优秀率', isRate: true },
    { key: 'stdDev', name: '标准差', isRate: false },
    { key: 'highAchieverPenetration', name: '高分层厚度', isRate: false },
    { key: 'strugglerSupportIndex', name: '后进生支撑力', isRate: false },
];

// 指标详情表格数据
const metricsTableData = computed(() => {
    if (!selectedClassReport.value || !fullReport.value?.groupStats) return [];

    const classStats = selectedClassReport.value.tableStats;
    const gradeStats = fullReport.value.groupStats;

    return metricsToShow.map(metric => {
        const rowData: { metricName: string; values: Record<string, any> } = {
            metricName: metric.name,
            values: {}
        };
        subjectKeys.value.forEach(subjKey => {
            const classVal = classStats[subjKey as keyof typeof classStats]?.[metric.key as keyof object];
            const gradeVal = gradeStats[subjKey as keyof typeof gradeStats]?.[metric.key as keyof object];
            const diff = (typeof classVal === 'number' && typeof gradeVal === 'number') ? classVal - gradeVal : null;

            rowData.values[subjKey] = {
                classValue: formatValue(classVal, metric.isRate),
                gradeValue: formatValue(gradeVal, metric.isRate),
                diff: diff,
            };
        });
        return rowData;
    });
});

// 学生名单
const studentRosterData = computed(() => {
    return selectedClassReport.value?.students || [];
});

// --- 辅助函数 ---
const formatValue = (value: any, isRate: boolean) => {
    if (value === undefined || value === null) return 'N/A';
    return isRate ? `${(Number(value) * 100).toFixed(1)}%` : Number(value).toFixed(2);
};

const formatDiff = (diff: number | null) => {
    if (diff === null) return '';
    const sign = diff > 0 ? '+' : '';
    const formatted = formatValue(diff, false);
    return `(${sign}${formatted})`;
};

const getDiffClass = (diff: number | null) => {
    if (diff === null) return '';
    return diff > 0 ? 'diff-positive' : 'diff-negative';
};
</script>

<style scoped>
.class-diagnostics-tab {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.class-selector-bar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background-color: #f5f7fa;
  border-radius: 6px;
}
.toolbar-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  white-space: nowrap;
}
.section-row {
  margin-bottom: 1.5rem;
}
.content-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}
:deep(.el-card__header) {
  font-weight: 600;
  color: #303133;
}
.metric-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  line-height: 1.4;
}
.class-value {
  font-weight: 500;
  font-size: 1.05em;
}
.grade-value {
  font-size: 0.8em;
  color: #909399;
}
.diff-positive {
  font-size: 0.8em;
  color: #67c23a;
}
.diff-negative {
  font-size: 0.8em;
  color: #f56c6c;
}
</style>