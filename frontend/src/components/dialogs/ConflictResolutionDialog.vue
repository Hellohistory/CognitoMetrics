<!--
/**
 * @file src/components/dialogs/ConflictResolutionDialog.vue
 * @component ConflictResolutionDialog
 * @description 解决班级中重名学生的分数归属冲突对话框
 * @author Hellohistory
 * @example
 * ```vue
 * <ConflictResolutionDialog
 *   v-model="dialogVisible"
 *   :conflict="conflictData"
 *   @resolve="onResolve"
 * />
 * ```
 */
-->

<template>
  <el-dialog
    v-model="dialogVisible"
    title="解决重名冲突"
    width="500px"
    :close-on-click-modal="false"
  >
    <div v-if="conflict">
      <p>
        班级中有多位学生叫
        <strong>"{{ conflict.conflictItem.name }}"</strong>。
      </p>
      <p>
        请选择将分数
        <strong>{{ conflict.conflictItem.score }}</strong>
        分配给哪位学生：
      </p>
      <el-radio-group
        v-model="selectedStudentId"
        class="radio-group"
      >
        <el-radio
          v-for="student in conflict.candidates"
          :key="student.id"
          :value="student.id"
        >
          {{ student.name }} (学号: {{ student.student_no }})
        </el-radio>
      </el-radio-group>
    </div>

    <template #footer>
      <el-button @click="handleResolve(undefined)">跳过此条</el-button>
      <el-button type="primary" @click="handleResolve(selectedStudentId)">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import {
  ElDialog,
  ElRadioGroup,
  ElRadio,
  ElButton,
} from 'element-plus';
import type { IStudent } from '@/types/dataModels';

/**
 * 组件接收的 props 定义
 */
const props = defineProps<{
  /**
   * 控制对话框显示/隐藏
   */
  modelValue: boolean;
  /**
   * 冲突数据，包含冲突项和候选学生列表
   * @default null
   */
  conflict: ConflictData | null;
}>();

/**
 * 组件触发的事件
 * @emits update:modelValue - 对话框显示状态变化时触发，payload 为布尔值
 * @emits resolve - 解析冲突时触发，payload 为选中学生的 ID 或 null（跳过）
 */
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'resolve', value: number | null): void;
}>();

/**
 * 选中的学生 ID，undefined 表示尚未选择
 */
const selectedStudentId = ref<number | undefined>(undefined);

/**
 * 对话框显示状态的双向绑定计算属性
 */
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val),
});

/**
 * 监听冲突数据变化，重置选中状态
 */
watch(
  () => props.conflict,
  (newConflict) => {
    if (newConflict) {
      selectedStudentId.value = undefined;
    }
  },
);

/**
 * 处理“确定”或“跳过”操作
 * @param id - 选中的学生 ID，undefined 表示跳过
 */
function handleResolve(id: number | undefined) {
  const payload = id === undefined ? null : id;
  emit('resolve', payload);
}

/**
 * 冲突数据接口定义
 */
export interface ConflictData {
  conflictItem: { name: string; score: number };
  candidates: IStudent[];
}
</script>

<style scoped>
.radio-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}
</style>
