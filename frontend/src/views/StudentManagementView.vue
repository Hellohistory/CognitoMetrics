<!--
/**
 * @file src/views/StudentManagement.vue
 * @component StudentManagement
 * @description 学生管理主视图，包含班级树导航与主内容区域，根据选中节点展示学生列表或年级概览
 * @example
 * ```vue
 * <StudentManagement />
 * ```
 */
-->

<template>
  <el-container class="student-management-container app-page">
    <el-aside width="280px" class="class-tree-aside">
      <class-navigation-panel />
    </el-aside>

    <el-main class="content-main">
      <StudentGrid
        v-if="classStore.selectedClass"
        :selected-class="classStore.selectedClass"
      />
      <GradeOverview
        v-else
        :grade="classStore.selectedGrade"
      />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ElContainer, ElAside, ElMain } from 'element-plus';
import ClassNavigationPanel from '@/components/ClassNavigationPanel.vue';
import StudentGrid from '@/components/StudentGrid.vue';
import GradeOverview from '@/components/GradeOverview.vue';
import { useClassStore } from '@/stores/classStore.ts';

/**
 * 全局班级/年级状态存储
 * 负责维护当前选中年级、班级及树状数据
 */
const classStore = useClassStore();
</script>

<style scoped>
.student-management-container {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  height: calc(100vh - var(--app-header-height) - 48px);
  min-height: 620px;
  gap: 16px;
}

.class-tree-aside {
  width: auto !important;
  min-width: 0;
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius);
  overflow: hidden;
  background: var(--app-surface);
  box-shadow: var(--app-shadow-soft);
}

.content-main {
  min-width: 0;
  padding: 0;
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius);
  overflow: hidden;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: var(--app-shadow-soft);
}

@media (max-width: 980px) {
  .student-management-container {
    grid-template-columns: 1fr;
    height: auto;
    min-height: 0;
  }
  .class-tree-aside {
    min-height: 360px;
  }
  .content-main {
    min-height: 560px;
  }
}
</style>
