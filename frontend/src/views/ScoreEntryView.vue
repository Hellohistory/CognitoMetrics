<template>
  <div class="score-entry-view">
    <header class="top-bar">
      <div class="exam-selector">
        <el-select
          v-model="examStore.selectedExamId"
          placeholder="请选择或新建一场考试"
          size="large"
          class="exam-select"
          @change="onExamChange"
          clearable
          filterable
        >
          <el-option
            v-for="exam in examStore.examList"
            :key="exam.id"
            :label="`${exam.name} (${getStatusText(exam.status)})`"
            :value="exam.id"
            :disabled="exam.status !== 'draft'"
          />
        </el-select>
        <el-button @click="newExamDialogVisible = true" type="primary" size="large" :icon="Plus">新建考试</el-button>

        <el-button
          v-if="selectedExam && selectedExam.status === 'draft'"
          @click="handleDeleteExam"
          type="danger"
          size="large"
          :icon="Delete"
          plain
        >
          删除考试
        </el-button>

        <el-button
          v-if="selectedExam && selectedExam.status === 'completed'"
          @click="handleUnlockExam"
          type="warning"
          size="large"
          :icon="Unlock"
          plain
        >
          解锁并重新编辑
        </el-button>
      </div>
      <div class="header-actions">
        <el-button @click="handleFinalizeExam" type="success" size="large" :icon="SelectIcon" :disabled="!canEditScores">完成录入并定稿</el-button>
      </div>
    </header>

    <el-container class="main-content">
      <el-aside width="280px" class="class-panel">
        <class-navigation-panel />
      </el-aside>

      <el-main class="grid-panel">
        <div class="grid-toolbar">
          <h2>{{ gridTitle }}</h2>
          <div class="grid-actions">
            <el-button @click="pasteDialogVisible = true" :disabled="!canEditScores || gridOptions.loading" :icon="Scissor">
              粘贴成绩单
            </el-button>
          </div>
        </div>

        <div class="grid-wrapper">
          <vxe-grid v-bind="gridOptions" ref="gridRef" @edit-closed="handleCellEditClosed" :edit-config="gridEditConfig"></vxe-grid>
        </div>

      </el-main>
    </el-container>

    <NewExamDialog v-model="newExamDialogVisible" @submitted="onNewExamCreated"/>
    <PasteScoreDialog v-model="pasteDialogVisible" @submit="handlePasteSubmit" />
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref, watch, nextTick} from 'vue';
import {ElAside, ElButton, ElContainer, ElMain, ElMessage, ElOption, ElSelect, ElMessageBox} from 'element-plus';
import { Plus, Select as SelectIcon, Scissor, Unlock, Delete } from '@element-plus/icons-vue';
import type {VxeGridInstance, VxeGridProps, VxeGridEvents} from 'vxe-table';

import {useExamStore} from '@/stores/examStore';
import {useClassStore} from '@/stores/classStore';

import {getExamDetails} from '@/api/examAPI';
import {getStudentsByClass} from '@/api/studentApi';
import {getScoresForClass, saveSingleScore} from '@/api/scoreAPI';

import type {IExam, IExamSubjectDetail, IStudent, ISingleScoreUpdate} from '@/types/dataModels';
import ClassNavigationPanel from '../components/ClassNavigationPanel.vue';
import NewExamDialog from '../components/dialogs/NewExamDialog.vue';
import PasteScoreDialog, { type PasteSubmitPayload } from '../components/dialogs/PasteScoreDialog.vue';

const examStore = useExamStore();
const classStore = useClassStore();

const gridRef = ref<VxeGridInstance>();
const newExamDialogVisible = ref(false);
const pasteDialogVisible = ref(false);
const currentSubjects = ref<IExamSubjectDetail[]>([]);

onMounted(() => {
  examStore.fetchExams();
  classStore.fetchClassTree();
});

const selectedExam = computed(() => examStore.selectedExam);
const selectedClass = computed(() => classStore.selectedClass);
const canEditScores = computed(() => !!(selectedExam.value && selectedExam.value.status === 'draft' && selectedClass.value));

const gridTitle = computed(() => {
  if (!selectedExam.value) return "请先选择一场考试";
  if (!selectedClass.value) return `考试: ${selectedExam.value.name} (请选择班级)`;
  return `${selectedExam.value.name} - ${selectedClass.value.name} - 成绩单`;
});

const gridEditConfig = computed((): VxeGridProps['editConfig'] => ({
    trigger: 'click',
    mode: 'cell',
    enabled: canEditScores.value
}));

const gridOptions = reactive<VxeGridProps>({
  border: true,
  stripe: true,
  height: '100%',
  loading: false,
  columnConfig: { resizable: true },
  keepSource: true,
  data: [] as (IStudent & Record<string, number | null>)[],
  columns: [],
});

const onExamChange = async (examId: number | string | undefined) => {
  if (examId && classStore.selectedClass) {
    await updateGridStructureAndData();
  } else if (!examId) {
    gridOptions.columns = [];
    gridOptions.data = [];
    currentSubjects.value = [];
  }
};

watch(() => classStore.selectedClass, async () => {
    if (examStore.selectedExamId) {
        await updateGridStructureAndData();
    }
}, { deep: true });

async function updateGridStructureAndData() {
    if (!selectedExam.value || !selectedClass.value) {
        gridOptions.data = [];
        gridOptions.columns = [];
        currentSubjects.value = [];
        return;
    }

    gridOptions.loading = true;
    try {
        const examId = selectedExam.value.id;
        const classId = selectedClass.value.id;

        const [examDetails, students, existingScores] = await Promise.all([
            getExamDetails(examId),
            getStudentsByClass(classId),
            getScoresForClass(examId, classId)
        ]);

        currentSubjects.value = examDetails.subjects;
        const scoreMap = new Map(
            existingScores.map(scoreRecord => [scoreRecord.student_id, scoreRecord.subject_scores])
        );

        gridOptions.columns = [
          {type: 'seq', width: 60, title: '#', fixed: 'left'},
          {field: 'name', title: '姓名', width: 120, fixed: 'left'},
          {field: 'student_no', title: '学号', width: 180},
          ...examDetails.subjects.map(sub => ({
            field: sub.name,
            title: `${sub.name} (满分${sub.full_mark})`,
            editRender: {name: '$input', props: {type: 'float', digits: 1}},
          }))
        ];

        gridOptions.data = students.map(student => {
            const studentWithScores: any = { ...student };
            const savedScores = scoreMap.get(student.id);
            examDetails.subjects.forEach(subject => {
                studentWithScores[subject.name] = savedScores?.[subject.name] ?? null;
            });
            return studentWithScores;
        });

    } catch (e) {
        ElMessage.error("加载考试或成绩数据失败");
    } finally {
        gridOptions.loading = false;
    }
}

const handleCellEditClosed: VxeGridEvents.EditClosed = async ({ row, column }) => {
    const field = column.field;
    if(!field) return;

    const isChanged = gridRef.value?.isUpdateByRow(row);
    if (!isChanged) return;

    const cellValue = row[field];
    const newScore = (cellValue === '' || cellValue == null) ? null : parseFloat(String(cellValue));

    if (newScore !== null && isNaN(newScore)) {
        ElMessage.error("请输入有效的数字成绩");
        gridRef.value?.revertData(row);
        return;
    }

    const payload: ISingleScoreUpdate = {
        exam_id: selectedExam.value!.id,
        student_id: row.id,
        subject_name: field,
        score: newScore,
    };

    try {
        await saveSingleScore(payload);

        // 【修正 4】: (核心修正点) 为 clearEdit 使用单参数版本，解决 "Expected 0-1 arguments, but got 2" 错误
        gridRef.value?.clearEdit(row);

        nextTick(() => {
            ElMessage({ type: 'success', message: `“${row.name}”的“${field}”成绩已保存`, duration: 1500 });
        });
    } catch (error) {
        // 【修正 5】: (核心修正点) 同样，为 catch 块中的 revertData 使用单参数版本
        gridRef.value?.revertData(row);
    }
};

const handleFinalizeExam = async () => {
    if(!selectedExam.value) return;
    try {
        await ElMessageBox.confirm(
            `确定要定稿考试 “${selectedExam.value.name}” 吗？<br/>定稿后，所有成绩将被锁定，无法再进行编辑。`,
            '确认定稿',
            {
                confirmButtonText: '确定定稿',
                cancelButtonText: '取消',
                type: 'warning',
                dangerouslyUseHTMLString: true
            }
        );
        const success = await examStore.handleFinalizeExam(selectedExam.value.id);
        if(success) ElMessage.success("考试已定稿，成绩已锁定！");
    } catch (error) {
        if(error !== 'cancel') console.error(error);
    }
};

const handleUnlockExam = async () => {
    if (!selectedExam.value) return;
    try {
        await ElMessageBox.confirm(
            `确定要解锁考试 “${selectedExam.value.name}” 吗？<br/>解锁后，其成绩将可以被重新编辑。`,
            '确认解锁',
            {
                confirmButtonText: '确定解锁',
                cancelButtonText: '取消',
                type: 'warning',
                dangerouslyUseHTMLString: true,
            }
        );
        const success = await examStore.handleUnlockExam(selectedExam.value.id);
        if(success) ElMessage.success(`考试 “${selectedExam.value.name}” 已解锁。`);
    } catch (error) {
        if (error !== 'cancel') console.error(error);
    }
};

const handleDeleteExam = async () => {
    if (!selectedExam.value) return;
    try {
        await ElMessageBox.confirm(
            `确定要永久删除考试 “${selectedExam.value.name}” 吗？<br/>只有未录入任何成绩的草稿考试才能被删除，此操作不可撤销。`,
            '危险操作警告',
            {
                confirmButtonText: '确定删除',
                cancelButtonText: '取消',
                type: 'error',
                dangerouslyUseHTMLString: true,
            }
        );
        const success = await examStore.handleDeleteExam(selectedExam.value.id);
        if (success) ElMessage.success("考试已成功删除！");
    } catch (error) {
        if (error !== 'cancel') console.error(error);
    }
};

const onNewExamCreated = (newExam: IExam) => {
    examStore.selectExam(newExam.id);
};

const handlePasteSubmit = (payload: PasteSubmitPayload) => {
    const { text } = payload;
    if (!canEditScores.value || !gridRef.value) return;

    const { fullData } = gridRef.value.getTableData();
    if (fullData.length === 0) return ElMessage.warning("请先选择班级以加载学生名单。");

    const subjectMap = new Map(currentSubjects.value.map(s => [s.name, s.full_mark]));

    const lines = text.trim().split(/\r?\n/).filter(Boolean);
    if (lines.length < 2) return ElMessage.error("粘贴内容格式错误，至少需要包含表头和一行数据。");

    const headers = lines[0].split(/[\s,，\t]+/).filter(Boolean);
    const dataLines = lines.slice(1);

    const columnMapping: { colIndex: number; subjectName: string }[] = [];
    headers.forEach((header, index) => {
        if (index > 0 && subjectMap.has(header)) {
            columnMapping.push({ colIndex: index, subjectName: header });
        }
    });

    if (columnMapping.length === 0) return ElMessage.warning("未在粘贴的表头中找到任何与当前考试匹配的科目。");

    const updateTasks: Promise<any>[] = [];
    let updatedRowCount = 0;

    dataLines.forEach(line => {
        const parts = line.split(/[\s,，\t]+/);
        const name = parts[0];
        const studentRow = fullData.find(s => s.name === name);

        if (studentRow) {
            updatedRowCount++;
            columnMapping.forEach(mapping => {
                const scoreStr = parts[mapping.colIndex];
                if (scoreStr === undefined) return;

                // 【修改点 3】: 同样修正 `isNaN` 的类型错误并处理无效数字
                // 1. 将解析后的值存入一个可变变量 `scoreValue`
                // 2. 检查 `scoreValue` 是否为 `NaN`，如果是，则将其统一处理为 `null`
                // 3. 这样 `scoreValue` 的类型就是明确的 `number | null`，可以安全地使用
                let scoreValue: number | null = (scoreStr === '' || scoreStr == null) ? null : parseFloat(scoreStr);
                if (scoreValue !== null && isNaN(scoreValue)) {
                    scoreValue = null; // 将 "abc" 这样的无效字符串解析结果 NaN 转换成 null
                }

                studentRow[mapping.subjectName] = scoreValue;

                const singleScorePayload: ISingleScoreUpdate = {
                    exam_id: selectedExam.value!.id,
                    student_id: studentRow.id,
                    subject_name: mapping.subjectName,
                    score: scoreValue,
                };
                updateTasks.push(saveSingleScore(singleScorePayload));
            });
        }
    });

    Promise.allSettled(updateTasks).then(() => {
        ElMessage.success(`粘贴操作完成，已为匹配到的 ${updatedRowCount} 名学生提交了成绩保存请求。`);
        gridRef.value?.updateData();
    });
};

const getStatusText = (status: string) => {
    const map: Record<string, string> = {
        draft: '草稿',
        submitted: '分析中',
        processing: '处理中',
        completed: '已锁定',
        failed: '分析失败'
    };
    return map[status] || '未知';
}
</script>

<style scoped>
.score-entry-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
}
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background-color: #ffffff;
  border-bottom: 1px solid #dcdfe6;
  flex-shrink: 0;
}
.exam-selector {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.exam-select {
  width: 350px;
}
.header-actions {
  margin-left: auto;
}
.main-content {
  flex-grow: 1;
  overflow: hidden;
}
.class-panel {
  background-color: #ffffff;
  height: 100%;
}
.grid-panel {
  display: flex;
  flex-direction: column;
  padding: 16px 24px;
  height: 100%;
}
.grid-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
}
.grid-toolbar h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}
.grid-actions {
  display: flex;
  gap: 12px;
}
.grid-wrapper {
  flex-grow: 1;
  min-height: 0;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background-color: #ffffff;
}
</style>