<template>
  <el-dialog
    v-model="dialogVisible"
    title="新建考试"
    width="600px"
    :close-on-click-modal="false"
    @closed="resetForm"
  >
    <el-form
      ref="formRef"
      :model="examForm"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="考试名称" prop="name">
        <el-input
          v-model="examForm.name"
          placeholder="例如：2025学年第一学期期中考"
        />
      </el-form-item>

      <el-form-item label="考试日期" prop="exam_date">
        <el-date-picker
          v-model="examForm.exam_date"
          type="date"
          placeholder="选择日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          style="width: 100%;"
        />
      </el-form-item>

      <el-divider />

      <el-form-item
        v-for="(subject, index) in examForm.subjects"
        :key="index"
        :label="`科目 ${index + 1}`"
        :prop="`subjects.${index}.name`"
        :rules="{ required: true, message: '科目名不能为空', trigger: 'blur' }"
      >
        <div class="subject-item">
          <el-input
            v-model="subject.name"
            placeholder="科目名"
            style="width: 180px;"
          />
          <el-input-number
            v-model="subject.full_mark"
            :min="0"
            :max="300"
            placeholder="满分"
            controls-position="right"
            style="width: 150px; margin: 0 10px;"
          />
          <el-button
            type="danger"
            :icon="Delete"
            circle
            plain
            @click="removeSubject(index)"
          />
        </div>
      </el-form-item>

      <el-form-item>
        <el-button @click="addSubject" :icon="Plus">
          添加科目
        </el-button>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button
        type="primary"
        @click="submitForm"
        :loading="isSubmitting"
      >
        确定创建
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import {
  ElDialog, ElForm, ElFormItem, ElInput, ElDatePicker, ElDivider,
  ElButton, ElInputNumber, ElMessage, type FormInstance, type FormRules,
} from 'element-plus';
import { Plus, Delete } from '@element-plus/icons-vue';
import { useExamStore } from '@/stores/examStore';
import type { IExam, IExamWithSubjectsCreate } from '@/types/dataModels';

const props = defineProps<{ modelValue: boolean }>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'submitted', payload: IExam): void;
}>();

const examStore = useExamStore();
const formRef = ref<FormInstance>();
const isSubmitting = ref(false);

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

function initialFormState(): IExamWithSubjectsCreate {
  return {
    name: '',
    exam_date: new Date().toISOString().split('T')[0], // 默认当天
    subjects: [{ name: '', full_mark: 100 }],
  };
}

const examForm = reactive<IExamWithSubjectsCreate>(initialFormState());

const rules = reactive<FormRules<IExamWithSubjectsCreate>>({
  name: [{ required: true, message: '请输入考试名称', trigger: 'blur' }],
  exam_date: [{ required: true, message: '请选择考试日期', trigger: 'change' }],
});

function addSubject() {
  examForm.subjects.push({ name: '', full_mark: 100 });
}

function removeSubject(index: number) {
  if (examForm.subjects.length > 1) {
    examForm.subjects.splice(index, 1);
  } else {
    ElMessage.warning('至少需要保留一个科目');
  }
}

function resetForm() {
  Object.assign(examForm, initialFormState());
  formRef.value?.clearValidate();
}

async function submitForm() {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (!valid) return;
    isSubmitting.value = true;
    try {
      const newExam = await examStore.handleCreateExam(examForm);
      if (newExam) {
        ElMessage.success(`考试《${newExam.name}》创建成功！`);
        emit('submitted', newExam);
        dialogVisible.value = false;
      }
    } finally {
      isSubmitting.value = false;
    }
  });
}
</script>

<style scoped>
.subject-item {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>