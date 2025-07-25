<!--src/components/reports/ReportSummary.vue-->
<template>
  <div class="report-summary">
    <el-descriptions :column="3" border>
      <el-descriptions-item label="样本数 (N)">{{ stats.count }}</el-descriptions-item>
      <el-descriptions-item label="平均分 (Mean)">{{ stats.mean }}</el-descriptions-item>
      <el-descriptions-item label="标准差 (StdDev)">{{ stats.stdDev }}</el-descriptions-item>

      <el-descriptions-item label="最小值 (Min)">{{ stats.min }}</el-descriptions-item>
      <el-descriptions-item label="中位数 (Median)">{{ stats.median }}</el-descriptions-item>
      <el-descriptions-item label="最大值 (Max)">{{ stats.max }}</el-descriptions-item>

      <el-descriptions-item label="及格率 (Pass Rate)">
        <el-tag type="success">{{ (stats.passRate * 100).toFixed(1) }}%</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="优秀率 (Excellent Rate)">
        <el-tag type="warning">{{ (stats.excellentRate * 100).toFixed(1) }}%</el-tag>
      </el-descriptions-item>
       <el-descriptions-item label="难度 (Difficulty)">{{ stats.difficulty }}</el-descriptions-item>
    </el-descriptions>
  </div>
</template>

<script setup lang="ts">
import { ElDescriptions, ElDescriptionsItem, ElTag } from 'element-plus';
import type { IDescriptiveStats } from '@/types/dataModels';

defineProps<{
  stats: IDescriptiveStats;
}>();
</script>

<style scoped>
.report-summary {
  margin-top: 1rem;
}
</style>