<template>
  <div class="class-comparison-tab">
    <el-card class="integrated-card" shadow="never">
      <template #header>
        <div class="card-header-toolbar">
          <div class="toolbar-left">
            <span class="toolbar-label">对比指标:</span>
            <el-select v-model="selectedMetricKey" placeholder="请选择指标" size="default">
              <el-option
                v-for="metric in selectableMetrics"
                :key="metric.key"
                :label="metric.name"
                :value="metric.key" >
                <span>{{ metric.name }}</span>
              </el-option>
            </el-select>
            <el-tooltip
              v-if="selectedMetric"
              class="metric-tooltip"
              effect="dark"
              :content="selectedMetric.description"
              placement="top"
            >
              <el-icon><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>
        </div>
      </template>

      <div class="comparison-content-grid">
        <div class="content-panel">
          <h3 class="panel-title">对比图表</h3>
          <div class="panel-body">
            <VueEcharts
              v-if="barChartOption"
              :option="barChartOption"
              style="height: 100%;"
            />
            <el-empty v-else description="暂无图表数据" />
          </div>
        </div>

        <div class="content-panel">
          <h3 class="panel-title">详细数据</h3>
          <div class="panel-body">
            <el-table :data="tableData" border stripe height="100%">
              <el-table-column prop="subject" label="学科" width="100" fixed />
              <el-table-column
                v-for="className in classNames"
                :key="className"
                :label="className"
                :prop="className"
                align="center"
                min-width="90"
              />
              <el-table-column label="年级平均" prop="gradeAverage" align="center" min-width="100">
                 <template #default="{ row }">
                    <strong>{{ row.gradeAverage }}</strong>
                 </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';
import { ElCard, ElSelect, ElOption, ElTable, ElTableColumn, ElEmpty, ElTooltip, ElIcon } from 'element-plus';
import { QuestionFilled } from '@element-plus/icons-vue';
import VueEcharts from '@/components/charts/VueEcharts.vue';
import type { EChartsOption, SeriesOption } from 'echarts';

const store = useAnalysisReportStore();

const selectableMetrics = [
    { key: 'mean', name: '平均分', isRate: false, description: '班级全体学生在该科目上的算术平均分。' },
    { key: 'stdDev', name: '标准差', isRate: false, description: '衡量班级内部分数分布的离散程度，值越小说明学生水平越整齐。' },
    { key: 'passRate', name: '及格率', isRate: true, description: '分数达到总分60%的学生占总人数的百分比。' },
    { key: 'excellentRate', name: '优秀率', isRate: true, description: '分数达到总分85%的学生占总人数的百分比。' },
    { key: 'difficulty', name: '难度', isRate: false, description: '考试的难易程度，等于平均分除以满分，值越高说明考试越简单。' },
    { key: 'highAchieverPenetration', name: '高分层厚度', isRate: false, description: '总分排名前27%学生的平均分，反映了班级顶尖学生的整体实力。' },
    { key: 'strugglerSupportIndex', name: '后进生支撑力', isRate: false, description: '总分排名后27%学生的平均分，此数值反映了班级学业后进生的基本盘水平。' },
];

// v-model 现在绑定到指标的 key (字符串)，而不是整个对象
const selectedMetricKey = ref(selectableMetrics[0].key);

// 创建一个计算属性，根据 key 找到完整的指标对象
// 组件的其余部分可以继续使用 selectedMetric，无需大的改动
const selectedMetric = computed(() => {
    return selectableMetrics.find(m => m.key === selectedMetricKey.value) || selectableMetrics[0];
});

const colorPalette = ['#5470c6', '#ee6666', '#91cc75', '#3ba272', '#fc8452', '#ea7ccc'];
const gradeAvgColor = '#B0B0B0';

const fullReport = computed(() => store.fullReport);
const classNames = computed(() => fullReport.value?.tables.map(t => t.tableName) || []);
const subjectKeys = computed(() => {
  if (!fullReport.value?.groupStats) return [];
  return Object.keys(fullReport.value.groupStats).filter(key => key !== 'correlationMatrix');
});

const formatValue = (value: any, isRate: boolean | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    return isRate ? `${(Number(value) * 100).toFixed(1)}%` : Number(value).toFixed(2);
};

const chartFormattedData = computed(() => {
    if (!fullReport.value || !selectedMetric.value) return null;

    const metric = selectedMetric.value;
    const series: SeriesOption[] = [];
    const tables = fullReport.value.tables;

    tables.forEach((table, index) => {
        const data = subjectKeys.value.map(subjKey => table.tableStats[subjKey as keyof typeof table.tableStats]?.[metric.key as keyof object]);
        series.push({
            name: table.tableName,
            type: 'bar',
            data: data,
            barGap: '20%',
            emphasis: { focus: 'series' },
            itemStyle: { color: colorPalette[index % colorPalette.length] }
        });
    });

    const gradeData = subjectKeys.value.map(subjKey => {
      return fullReport.value?.groupStats[subjKey as keyof typeof fullReport.value.groupStats]?.[metric.key as keyof object]
    });
    series.push({
        name: '年级平均',
        type: 'bar',
        data: gradeData,
        itemStyle: { color: gradeAvgColor, borderWidth: 1, borderColor: gradeAvgColor }
    });

    return {
        legendData: [...classNames.value, '年级平均'],
        xAxisData: subjectKeys.value.map(s => s === 'totalScore' ? '总分' : s),
        series: series
    };
});

const tableData = computed(() => {
    if (!fullReport.value || !selectedMetric.value) return [];

    const metric = selectedMetric.value;
    const tables = fullReport.value.tables;
    const groupStats = fullReport.value.groupStats;

    return subjectKeys.value.map(subjKey => {
        const row: Record<string, any> = { subject: subjKey === 'totalScore' ? '总分' : subjKey };
        for (const table of tables) {
            const val = table.tableStats[subjKey as keyof typeof table.tableStats]?.[metric.key as keyof object];
            row[table.tableName] = formatValue(val, metric.isRate);
        }
        const gradeVal = groupStats[subjKey as keyof typeof groupStats]?.[metric.key as keyof object];
        row.gradeAverage = formatValue(gradeVal, metric.isRate);
        return row;
    });
});

const barChartOption = computed<EChartsOption | null>(() => {
    if (!chartFormattedData.value) return null;

    const isRate = selectedMetric.value?.isRate;
    const option: EChartsOption = {
        tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, valueFormatter: (value) => formatValue(value, isRate) },
        legend: { data: chartFormattedData.value.legendData, bottom: 0, type: 'scroll' },
        grid: { top: '12%', left: '3%', right: '4%', bottom: '12%', containLabel: true },
        xAxis: [{ type: 'category', data: chartFormattedData.value.xAxisData }],
        yAxis: [{
            type: 'value',
            axisLabel: {
                formatter: (value: number): string => {
                    return isRate ? `${(value * 100).toFixed(0)}%` : String(value);
                }
            }
        }],
        series: chartFormattedData.value.series
    };
    return option;
});
</script>

<style scoped>
.integrated-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}
.card-header-toolbar {
  display: flex;
  align-items: center;
  width: 100%;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-grow: 1; /* <-- 新增这一行 */
  /* min-width: 0; */ /* 另一个备选的修复技巧 */
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.toolbar-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  white-space: nowrap;
}
.metric-tooltip {
  font-size: 16px;
  color: #909399;
  cursor: help;
}
.comparison-content-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  height: 520px;
}
.content-panel {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  height: 100%;
}
.panel-title {
  font-size: 1rem;
  font-weight: 600;
  color: #303133;
  padding: 0.8rem 1rem;
  border-bottom: 1px solid #f0f0f0;
  margin: 0;
}
.panel-body {
  flex-grow: 1;
  padding: 1rem;
  overflow: hidden;
}
@media (max-width: 1200px) {
  .comparison-content-grid {
    grid-template-columns: 1fr;
    height: auto;
  }
  .content-panel {
    height: 500px;
  }
}
:deep(.el-table th.el-table__cell) {
  background-color: #fafafa;
  color: #606266;
}
</style>