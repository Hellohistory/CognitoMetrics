<!--src/components/reports/ReportSummary.vue-->
<template>
  <div class="report-summary">
    <el-descriptions :column="descriptionColumns" border>
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { ElDescriptions, ElDescriptionsItem, ElTag } from 'element-plus';
import type { IDescriptiveStats } from '@/types/dataModels';

defineProps<{
  stats: IDescriptiveStats;
}>();

const isCompact = ref(false);
let compactQuery: MediaQueryList | null = null;

const descriptionColumns = computed(() => (isCompact.value ? 1 : 3));

const updateCompact = () => {
  isCompact.value = Boolean(compactQuery?.matches);
};

onMounted(() => {
  compactQuery = window.matchMedia('(max-width: 560px)');
  updateCompact();
  compactQuery.addEventListener('change', updateCompact);
});

onBeforeUnmount(() => {
  compactQuery?.removeEventListener('change', updateCompact);
});
</script>

<style scoped>
.report-summary {
  max-width: 100%;
  margin-top: 1rem;
  overflow: hidden;
}

:deep(.el-descriptions__label),
:deep(.el-descriptions__content) {
  overflow-wrap: anywhere;
  word-break: normal;
}

@media (max-width: 560px) {
  .report-summary {
    margin-top: 0.5rem;
  }

  :deep(.el-descriptions__label) {
    width: 48%;
  }
}
</style>
