<template>
  <el-dialog
    v-model="dialogVisible"
    title="新建年级"
    width="450px"
    :close-on-click-modal="false"
    @closed="onDialogClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      style="padding-right: 20px;"
      @submit.prevent="handleSubmit"
    >
      <el-form-item label="年级名称" prop="name">
        <el-input v-model="form.name" placeholder="例如：2025届 或 2025级" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="isSubmitting" @click="handleSubmit">
        确定创建
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import { ElDialog, ElForm, ElFormItem, ElInput, ElButton, ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { useClassStore } from '@/stores/classStore';
import type { IGradeCreate } from '@/types/dataModels';

const props = defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'success'): void;
}>();

const classStore = useClassStore();
const formRef = ref<FormInstance>();
const isSubmitting = ref(false);

const form = reactive<IGradeCreate>({ name: '' });
const rules = reactive<FormRules>({
  name: [{ required: true, message: '年级名称不能为空', trigger: 'blur' }],
});

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

const resetForm = () => {
  form.name = '';
  formRef.value?.clearValidate();
};

const onDialogClosed = () => {
  resetForm();
};

const handleSubmit = async () => {
  await formRef.value?.validate(async (valid) => {
    if (!valid) return;
    isSubmitting.value = true;
    try {
      const success = await classStore.handleCreateGrade(form);
      if (success) {
        ElMessage.success('年级创建成功');
        emit('success');
        dialogVisible.value = false;
      }
    } finally {
      isSubmitting.value = false;
    }
  });
};
</script>