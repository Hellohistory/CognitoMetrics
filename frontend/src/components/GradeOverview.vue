<template>
  <div class="grade-overview" v-if="grade">
    <h2>{{ grade.name }} - 年级概览</h2>
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="card-content">
            <div class="card-label">班级总数</div>
            <div class="card-value">{{ grade.classes.length }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="card-content">
            <div class="card-label">学生总人数</div>
            <div class="card-value">{{ totalStudents }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-table :data="grade.classes" stripe border class="class-table" empty-text="该年级下暂无班级">
      <el-table-column type="index" label="#" width="80" />
      <el-table-column prop="name" label="班级名称" sortable />
      <el-table-column prop="student_count" label="学生人数" sortable width="150" />
      <el-table-column prop="enrollment_year" label="入学年份" sortable width="150" />
      <el-table-column label="操作" width="150" align="center">
        <template #default="{ row }">
          <el-button type="primary" link @click="selectClass(row)">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <el-empty v-else description="请在左侧导航栏中选择一个年级以查看概览" />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { ElRow, ElCol, ElCard, ElTable, ElTableColumn, ElButton, ElEmpty } from 'element-plus';
import type { IGradeNode, IClassNode } from '@/types/dataModels';
import { useClassStore } from '@/stores/classStore';

const props = defineProps<{
  grade: IGradeNode | null;
}>();

const classStore = useClassStore();

const totalStudents = computed(() => {
  if (!props.grade) return 0;
  return props.grade.classes.reduce((sum, cls) => sum + cls.student_count, 0);
});

const selectClass = (classNode: IClassNode) => {
  // 调用 store 中新的、无歧义的 selectNode 方法，直接传递班级节点对象
  classStore.selectNode(classNode);
};
</script>

<style scoped>


</style>