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
  <el-container class="student-management-container">
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
  height: calc(100vh - 60px);
}

.class-tree-aside {
  border-right: 1px solid #e0e0e0;
}

.content-main {
  padding: 1rem 2rem;
}
</style>
