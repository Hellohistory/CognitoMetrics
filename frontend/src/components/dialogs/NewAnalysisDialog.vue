<template>
  <el-dialog
    v-model="dialogVisible"
    title="发起新分析"
    width="600px"
    :close-on-click-modal="false"
    @open="onDialogOpen"
    @closed="resetForm"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      v-loading="isDataLoading"
    >
      <el-form-item label="选择考试" prop="exam_id">
        <el-select
          v-model="form.exam_id"
          placeholder="请选择要分析的考试"
          style="width: 100%"
          filterable
        >
          <el-option
            v-for="exam in availableExams"
            :key="exam.id"
            :label="exam.name"
            :value="exam.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="报告名称" prop="report_name">
        <el-input
          v-model="form.report_name"
          placeholder="为本次分析报告命名"
        />
      </el-form-item>

      <el-divider />

      <el-radio-group v-model="form.scope_level" class="scope-radio-group">
        <el-radio value="FULL_EXAM">整场考试</el-radio>
        <el-radio value="GRADE">按年级</el-radio>
        <el-radio value="CLASS">按班级</el-radio>
      </el-radio-group>

      <el-form-item
        label="选择年级"
        v-if="form.scope_level === 'GRADE'"
        prop="selected_grade_ids"
      >
        <el-select
          v-model="form.selected_grade_ids"
          multiple
          placeholder="请选择一个或多个年级"
          style="width: 100%"
          collapse-tags
          collapse-tags-tooltip
        >
          <el-option
            v-for="grade in classStore.classTree"
            :key="grade.id"
            :label="grade.name"
            :value="grade.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        label="选择班级"
        v-if="form.scope_level === 'CLASS'"
        prop="selected_class_ids"
      >
        <el-tree
          ref="classTreeRef"
          :data="classStore.classTree"
          :props="classStore.treeProps"
          show-checkbox
          node-key="id"
          class="scope-tree"
          check-strictly
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button
        type="primary"
        @click="handleSubmit"
        :loading="isSubmitting"
      >
        提交分析
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import {
  ElDialog,
  ElForm,
  ElFormItem,
  ElSelect,
  ElOption,
  ElInput,
  ElRadioGroup,
  ElRadio,
  ElTree,
  ElButton,
  ElMessage,
  ElDivider,
  vLoading,
  type FormInstance,
  type FormRules,
} from 'element-plus';

import {type ISubmitResponse, submitAnalysisJob} from '@/api/analysisApi';
import { useClassStore } from '@/stores/classStore';
import { getExams } from '@/api/examAPI';
import type {IAnalysisSubmissionRequest, IExam} from "@/types/dataModels.ts";

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'submitted', response: ISubmitResponse): void; // 修正点
}>();

const formRef = ref<FormInstance>();
const classTreeRef = ref<InstanceType<typeof ElTree>>();
const classStore = useClassStore();

const isDataLoading = ref(false);
const isSubmitting = ref(false);
const allExams = ref<IExam[]>([]);

const availableExams = computed(() =>
  allExams.value.filter(e =>
    ['draft', 'submitted', 'completed'].includes(e.status)
  )
);

const createInitialFormState = () => ({
  exam_id: undefined as number | undefined,
  report_name: '',
  scope_level: 'FULL_EXAM' as 'FULL_EXAM' | 'GRADE' | 'CLASS',
  selected_grade_ids: [] as number[],
  selected_class_ids: [] as number[],
});

const form = reactive(createInitialFormState());

const rules = reactive<FormRules>({
  exam_id: [{ required: true, message: '请选择一场考试', trigger: 'change' }],
  report_name: [{ required: true, message: '请输入报告名称', trigger: 'blur' }],
  selected_grade_ids: [
    { required: true, type: 'array', min: 1, message: '请至少选择一个年级', trigger: 'change' },
  ],
});

const dialogVisible = computed({
  get: () => props.modelValue,
  set: val => emit('update:modelValue', val),
});

watch(() => form.exam_id,
  (newExamId) => {
    if (!newExamId) {
      form.report_name = '';
      return;
    }
    const exam = allExams.value.find(e => e.id === newExamId);
    if (exam) {
      const date = new Date().toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }).replace('/', '月') + '日';
      form.report_name = `对《${exam.name}》的分析报告 - ${date}`;
    }
  }
);

async function onDialogOpen() {
  isDataLoading.value = true;
  try {
    const examPromise = allExams.value.length === 0 ? getExams() : Promise.resolve(allExams.value);
    const classTreePromise = classStore.classTree.length === 0 ? classStore.fetchClassTree() : Promise.resolve();
    const [examsResult] = await Promise.all([examPromise, classTreePromise]);
    allExams.value = examsResult;
  } catch {
    ElMessage.error('加载基础数据失败！');
  } finally {
    isDataLoading.value = false;
  }
}

function resetForm() {
  Object.assign(form, createInitialFormState());
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  await formRef.value?.validate(async valid => {
    if (!valid) return;

    let scope_ids: number[] = [];
    if (form.scope_level === 'GRADE') {
      scope_ids = form.selected_grade_ids;
    } else if (form.scope_level === 'CLASS') {
      const checkedNodes = classTreeRef.value?.getCheckedNodes(false, false) || [];
      scope_ids = checkedNodes.filter(node => 'student_count' in node).map(node => node.id); // 只选择班级节点
      if (!scope_ids.length) {
        ElMessage.warning('请至少选择一个班级');
        return;
      }
    }

    const submissionData: IAnalysisSubmissionRequest = {
      exam_id: form.exam_id!,
      report_name: form.report_name,
      scope: {
        level: form.scope_level,
        ids: scope_ids,
      },
    };

    isSubmitting.value = true;
    try {
      const response = await submitAnalysisJob(submissionData);
      ElMessage.success('分析任务已成功提交！');
      emit('submitted', response); // 修正点
      dialogVisible.value = false;
    } finally {
      isSubmitting.value = false;
    }
  });
}
</script>

<style scoped>
.scope-radio-group {
  margin-bottom: 20px;
}
.scope-tree {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 8px;
  max-height: 250px;
  overflow-y: auto;
}
</style>