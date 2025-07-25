<template>
  <div class="class-comparison">
    <el-table :data="comparisonData" border stripe class="comparison-table">
      <el-table-column prop="metricName" label="核心指标" width="180" fixed>
        <template #default="{ row }">
          <div class="metric-cell">
            <span>{{ row.metricName }}</span>
            <el-tooltip
              v-if="row.description"
              effect="dark"
              :content="row.description"
              placement="top-start"
              :enterable="false"
              :offset="10"
            >
              <el-icon class="metric-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        v-for="className in classNames"
        :key="className"
        :label="className"
      >
        <template #default="{ row }">
          <span>{{ row[className] ?? 'N/A' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="年级平均" width="120">
         <template #default="{ row }">
          <strong>{{ row.gradeAverage ?? 'N/A' }}</strong>
        </template>
      </el-table-column>
    </el-table>

    <div class="chart-section">
      <el-row :gutter="20">
        <el-col :span="12">
           <el-card class="chart-card">
            <template #header>班级成绩分布箱线图 (总分)</template>
            <VueEcharts v-if="boxPlotChartOption" :option="boxPlotChartOption" />
            <el-empty v-else description="无箱线图数据" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <div class="card-header-flex">
                <span>班级学科画像雷达图</span>
                <el-select v-model="selectedClassForRadar" placeholder="选择班级" size="small" style="width: 150px;">
                  <el-option v-for="name in classNames" :key="name" :label="name" :value="name" />
                </el-select>
              </div>
            </template>
            <VueEcharts v-if="radarChartOption" :option="radarChartOption" />
            <el-empty v-else description="请先选择一个班级" />
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElTable, ElTableColumn, ElCard, ElRow, ElCol, ElSelect, ElOption, ElEmpty, ElTooltip, ElIcon } from 'element-plus';
import { QuestionFilled } from '@element-plus/icons-vue';
import type { EChartsOption } from 'echarts';
import type { IFullReport } from '@/types/dataModels';
import VueEcharts from '@/components/charts/VueEcharts.vue';
import { createClassRadarOption } from '@/utils/chartOptionFactory';

const props = defineProps<{
  report: IFullReport;
  chartData: any;
}>();

const metricMapping = [
    { name: '平均分', key: 'mean', description: '班级全体学生在该科目上的算术平均分，反映班级的整体水平。' },
    { name: '标准差', key: 'stdDev', description: '衡量班级内部分数分布的离散程度。值越小，说明学生水平越整齐；值越大，则说明成绩差距悬殊。' },
    { name: '总分及格率 (%)', key: 'passRate', isRate: true, description: '分数达到总分60%的学生占班级总人数的百分比。' },
    { name: '优秀率 (%)', key: 'excellentRate', isRate: true, description: '分数达到总分85%的学生占班级总人数的百分比。' },
    { name: '高分层厚度', key: 'highAchieverPenetration', description: '总分排名前27%学生的平均分。此数值反映了班级顶尖学生的整体实力，值越高，说明“学霸”群体的领先优势越大。' },
    { name: '后进生支撑力', key: 'strugglerSupportIndex', description: '总分排名后27%学生的平均分。此数值反映了班级学业后进生的基本盘水平，值越高，说明班级的“短板”不短，兜底做得好。' },
    { name: '核心密度', key: 'academicCoreDensity', isRate: true, description: '分数在“平均分 ± 0.5个标准差”区间内的学生占比。此值反映中等水平(“腰部力量”)学生的规模，值越高说明学生水平更集中。' },
];

const classNames = computed(() => props.report.tables?.map(t => t.tableName) || []);

const comparisonData = computed(() => {
    return metricMapping.map(metric => {
        const row: Record<string, any> = {
            metricName: metric.name,
            description: metric.description
        };
        props.report.tables?.forEach(table => {
            const value = table.tableStats.totalScore[metric.key as keyof typeof table.tableStats.totalScore];
            row[table.tableName] = metric.isRate ? `${(Number(value) * 100).toFixed(1)}%` : Number(value).toFixed(2);
        });
        const gradeValue = props.report.groupStats.totalScore[metric.key as keyof typeof props.report.groupStats.totalScore];
        row.gradeAverage = metric.isRate ? `${(Number(gradeValue) * 100).toFixed(1)}%` : Number(gradeValue).toFixed(2);
        return row;
    });
});

const boxPlotChartOption = computed<EChartsOption | null>(() => {
    const boxplotData = props.chartData?.class_comparison_charts?.score_distribution_boxplot?.totalScore;
    if (!boxplotData || !boxplotData.data || boxplotData.data.length === 0) return null;
    return {
        tooltip: {
            trigger: 'item',
            axisPointer: { type: 'shadow' },
            formatter: (params: any) => {
                const [min, q1, median, q3, max] = params.value.slice(1);
                return `<strong>${params.name}</strong><br/>最大值: ${max}<br/>上四分位数 (Q3): ${q3}<br/>中位数: ${median}<br/>下四分位数 (Q1): ${q1}<br/>最小值: ${min}`;
            }
        },
        grid: { left: '10%', right: '10%', bottom: '15%', containLabel: true },
        xAxis: { type: 'category', data: boxplotData.categories, name: '班级/分组', nameLocation: 'middle', nameGap: 30 },
        yAxis: { type: 'value', name: '总分' },
        dataZoom: [{ type: 'inside' }, { show: true, type: 'slider', bottom: 10 }],
        series: [{ name: '成绩分布', type: 'boxplot', data: boxplotData.data }]
    };
});

const selectedClassForRadar = ref(classNames.value[0] || '');
const radarChartOption = computed<EChartsOption | null>(() => {
    if (!selectedClassForRadar.value) return null;
    const radarData = props.chartData?.class_comparison_charts?.class_profile_radar?.[selectedClassForRadar.value];
    if (!radarData) return null;
    return createClassRadarOption(radarData);
});

</script>

<style scoped>
.class-comparison {
  padding-top: 1rem;
}
.comparison-table {
  width: 100%;
}
.chart-section {
  margin-top: 2rem;
}
.chart-card {
  height: 450px;
  display: flex;
  flex-direction: column;
}
/* .el-card__body 深度选择器，确保图表容器能撑满 */
:deep(.chart-card .el-card__body) {
  flex-grow: 1;
  height: 100%;
  width: 100%;
}
:deep(.chart-card .el-card__body .echarts-container) {
  height: 100% !important;
}
.card-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-cell {
  display: flex;
  align-items: center;
}
.metric-help-icon {
  margin-left: 8px;
  color: #909399;
  cursor: pointer;
}
</style>