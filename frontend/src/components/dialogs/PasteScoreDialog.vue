<!--
/**
 * @file src/components/dialogs/PasteScoresDialog.vue
 * @component PasteScoresDialog
 * @description 粘贴成绩单对话框组件，用于从 Excel 粘贴表格区域并提交纯文本进行匹配
 * @author Hellohistory
 * @example
 * ```vue
 * <PasteScoresDialog
 *   v-model="dialogVisible"
 *   @submit="handlePasteSubmit"
 * />
 * ```
 */
-->

<template>
  <el-dialog
    v-model="dialogVisible"
    title="粘贴成绩单"
    width="600px"
    @open="onDialogOpen"
  >
    <el-form label-position="top">
      <el-form-item label="请从 Excel 中复制一块区域 (需包含表头，如：姓名、语文、数学等) 并粘贴到下方">
        <el-input
          v-model="pastedText"
          type="textarea"
          :rows="12"
          placeholder="姓名   语文   数学   英语
张三   95     88     92
李四   88     98     85
..."
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定并匹配</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import {
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElButton,
  ElMessage,
} from 'element-plus';

/**
 * 粘贴成绩单提交的 Payload 结构
 */
export interface PasteSubmitPayload {
  /** 用户粘贴的原始文本 */
  text: string;
}

/**
 * 组件 Props 定义
 */
const props = defineProps<{
  /** 控制对话框显示/隐藏 */
  modelValue: boolean;
}>();

/**
 * 组件触发的事件
 * @emits update:modelValue - 对话框显示状态变化时触发，payload 为布尔值
 * @emits submit - 用户确认粘贴并提交时触发，payload 为 PasteSubmitPayload 对象
 */
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'submit', payload: PasteSubmitPayload): void;
}>();

/** 用户粘贴的成绩文本 */
const pastedText = ref<string>('');

/** 对话框显示状态双向绑定 */
const dialogVisible = computed<boolean>({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val),
});

/**
 * 对话框打开时清空上次粘贴内容
 */
function onDialogOpen(): void {
  pastedText.value = '';
}

/**
 * 处理“确定并匹配”操作
 * - 校验非空后，触发 submit 事件并关闭对话框
 */
function handleSubmit(): void {
  if (!pastedText.value.trim()) {
    ElMessage.warning('请粘贴成绩内容');
    return;
  }
  emit('submit', { text: pastedText.value });
  dialogVisible.value = false;
}
</script>

<style scoped>

</style>
