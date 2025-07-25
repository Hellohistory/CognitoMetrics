<!--
/**
 * @file src/components/EChartsWrapper.vue
 * @component EChartsWrapper
 * @description 封装 ECharts 图表组件，支持自动初始化、配置更新与容器尺寸变化时的自适应
 * @author Hellohistory
 * @example
 * ```vue
 * <EChartsWrapper :option="chartOption" />
 * ```
 */
-->

<template>
  <div ref="chartRef" class="echarts-container" />
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, shallowRef } from 'vue';
import * as echarts from 'echarts';
import type { EChartsOption, ECharts } from 'echarts';

/**
 * 组件 Props 定义
 */
const props = defineProps<{
  /** ECharts 配置项，传入 null 则不渲染 */
  option: EChartsOption | null;
}>();

/** 图表容器的 DOM 引用 */
const chartRef = ref<HTMLDivElement | null>(null);
/** ECharts 实例引用 */
const chartInstance = shallowRef<ECharts | null>(null);
/** ResizeObserver 实例，用于监听容器尺寸变化 */
let resizeObserver: ResizeObserver | null = null;

/**
 * 初始化或重建图表实例并应用配置
 * - 高内聚：仅负责图表初始化与配置
 * - 低耦合：不涉及外部业务逻辑
 */
function initChart() {
  const container = chartRef.value;
  if (!container) return;

  // 如果容器尚未渲染或尺寸为 0，则延后初始化
  if (container.clientWidth === 0 || container.clientHeight === 0) {
    return;
  }

  // 销毁已有实例，确保全新渲染
  chartInstance.value?.dispose();
  chartInstance.value = echarts.init(container);

  // 如果传入了配置，则立即应用
  if (props.option) {
    chartInstance.value.setOption(props.option, { notMerge: true });
  }
}

/**
 * 监听 option 变化，更新已有图表配置
 */
watch(
  () => props.option,
  (newOption) => {
    if (chartInstance.value && newOption) {
      chartInstance.value.setOption(newOption, { notMerge: true });
    }
  },
  { deep: true }
);

/**
 * 组件挂载时：创建 ResizeObserver，并启动对容器的监听
 * - 当容器从隐藏变为可见，或尺寸发生变化时，自动初始化或 resize
 */
onMounted(() => {
  const container = chartRef.value;
  if (!container) return;

  resizeObserver = new ResizeObserver(() => {
    if (!container) return;
    const width = container.clientWidth;
    const height = container.clientHeight;
    if (width > 0 && height > 0) {
      if (chartInstance.value) {
        // 已有实例则仅调整大小
        chartInstance.value.resize();
      } else {
        // 无实例则初始化
        initChart();
      }
    }
  });

  resizeObserver.observe(container);
});

/**
 * 组件卸载时：清理 ResizeObserver 与 ECharts 实例
 */
onUnmounted(() => {
  resizeObserver?.disconnect();
  chartInstance.value?.dispose();
});
</script>

<style scoped>
.echarts-container {
  width: 100%;
  height: 400px;
}
</style>
