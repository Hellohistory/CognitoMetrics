<template>
  <el-dialog
    v-model="dialogVisible"
    :title="`在“${targetGrade?.name || ''}”下新建班级`"
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
      <el-form-item label="所属年级">
        <el-input :model-value="targetGrade?.name" disabled />
      </el-form-item>
      <el-form-item label="班级名称" prop="name">
        <el-input v-model="form.name" placeholder="例如：(1)班" />
      </el-form-item>
      <el-form-item label="入学年份" prop="enrollment_year">
        <el-input-number
          v-model="form.enrollment_year"
          :min="2000"
          :max="2050"
          controls-position="right"
          style="width: 100%"
        />
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
import { ElDialog, ElForm, ElFormItem, ElInput, ElButton, ElMessage, ElInputNumber, type FormInstance, type FormRules } from 'element-plus';
import { useClassStore } from '@/stores/classStore';
import type { IClassCreate, IGradeNode } from '@/types/dataModels';

const props = defineProps<{
  modelValue: boolean;
  targetGrade: IGradeNode | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'success'): void;
}>();

const classStore = useClassStore();
const formRef = ref<FormInstance>();
const isSubmitting = ref(false);

const createInitialFormState = () => ({
  name: '',
  enrollment_year: new Date().getFullYear(),
});

const form = reactive(createInitialFormState());
const rules = reactive<FormRules>({
  name: [{ required: true, message: '班级名称不能为空', trigger: 'blur' }],
  enrollment_year: [{ required: true, message: '入学年份不能为空', trigger: 'change' }],
});

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

const resetForm = () => {
  Object.assign(form, createInitialFormState());
  formRef.value?.clearValidate();
};

const onDialogClosed = () => {
  resetForm();
};

const handleSubmit = async () => {
  if (!props.targetGrade) return;
  await formRef.value?.validate(async (valid) => {
    if (!valid) return;

    const fullClassData: IClassCreate = {
      ...form,
      grade_id: props.targetGrade!.id,
    };

    isSubmitting.value = true;
    try {
      const success = await classStore.handleCreateClass(fullClassData);
      if (success) {
        ElMessage.success('班级创建成功');
        emit('success');
        dialogVisible.value = false;
      }
    } finally {
      isSubmitting.value = false;
    }
  });
};
</script>