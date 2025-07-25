<!--src/components/StudentGrid.vue-->
<template>
  <div class="student-grid-container">
    <div class="header-bar">
      <h1>{{ classStore.gridTitle }}</h1>
      <div class="action-buttons">
        <template v-if="!isEditMode">
            <el-button @click="handleExport" :disabled="!selectedClass || gridOptions.data?.length === 0">导出Excel</el-button>
            <el-button type="primary" @click="toggleEditMode(true)" :disabled="!selectedClass">
                进入编辑模式
            </el-button>
        </template>
        <template v-else>
          <el-button @click="handleOpenPasteDialog" :disabled="!selectedClass || isSubmitting">粘贴名单</el-button>
          <el-button type="primary" @click="handleUploadClick" :disabled="!selectedClass || isSubmitting">上传Excel</el-button>
          <el-divider direction="vertical" />
          <el-button type="success" @click="handleSaveChanges" :loading="isSubmitting" :disabled="isSubmitting">
            {{ isSubmitting ? '保存中...' : '保存全部更改' }}
          </el-button>
          <el-button @click="toggleEditMode(false)" :disabled="isSubmitting">取消编辑</el-button>
        </template>
      </div>
    </div>
    <div class="filters-bar">
        <el-switch
          v-model="showInactive"
          @change="handleFilterChange"
          size="large"
          inline-prompt
          style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949"
          active-text="显示所有学生"
          inactive-text="仅显示在读学生"
        />
    </div>
    <div v-if="selectedRows.length > 0 && !isEditMode" class="bulk-actions-bar">
        <span>已选择 {{ selectedRows.length }} 项</span>
        <el-divider direction="vertical" />
        <el-button type="primary" link @click="handleBulkAction('activate')">批量激活</el-button>
        <el-button type="warning" link @click="handleBulkAction('deactivate')">批量停用</el-button>
        <el-button type="info" link @click="handleBulkAction('transfer')">批量转移</el-button>
    </div>

    <el-alert
      v-if="isEditMode"
      title="您正处于编辑模式"
      description="在此模式下，您可以进行大范围的复制粘贴和修改。所有操作将在点击“保存全部更改”后生效。提示：支持导入含'姓名'和'学号'的Excel。"
      show-icon
      :closable="false"
      class="edit-mode-banner"
    />

    <div class="grid-wrapper">
      <vxe-grid
        ref="gridRef"
        v-bind="gridOptions"
        @edit-closed="handleEditClosed"
        @checkbox-change="onCheckboxChange"
        @checkbox-all="onCheckboxChange"
      >
        <template #name_cell="{ row }">
            <router-link :to="`/students/${row.id}`" class="student-link">{{ row.name }}</router-link>
        </template>
        <template #actions_cell="{ row }">
            <el-button v-if="row.is_active" type="warning" link @click="handleDeactivate(row)">停用</el-button>
            <el-button v-else type="success" link @click="handleActivate(row)">启用</el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="handleTransfer(row)">转班</el-button>
        </template>
      </vxe-grid>
    </div>

    <el-dialog v-model="pasteDialogVisible" title="粘贴学生姓名名单" width="500px">
      <p>请从Excel中复制列，然后粘贴到下面的文本框中。支持'姓名'或'姓名,学号'格式。</p>
      <el-input
        v-model="pastedNames"
        :rows="10"
        type="textarea"
        placeholder="张三&#10;李四,2025002&#10;王五"
      />
      <template #footer>
        <el-button @click="pasteDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPaste">添加到表格</el-button>
      </template>
    </el-dialog>

    <input type="file" ref="fileInput" @change="handleFileSelected" style="display: none" accept=".xlsx, .xls" />

    <StudentTransferDialog
      v-model="transferDialog.visible"
      :student-count="transferDialog.students.length"
      @confirm="onConfirmTransfer"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { ElButton, ElMessage, ElAlert, ElDialog, ElInput, ElDivider, ElMessageBox, ElSwitch } from 'element-plus';
import * as XLSX from 'xlsx';
import type { VxeGridInstance, VxeGridProps, VxeGridEvents, VxeTableEvents } from 'vxe-table';

import {
  getStudentsByClass,
  updateStudent,
  batchCreateStudents,
  deactivateStudent,
  activateStudent,
  batchTransferStudents,
  batchUpdateStatus
} from '@/api/studentApi';
import type { IClassNode, IStudent, IStudentCreate } from '@/types/dataModels';
import { useClassStore } from "@/stores/classStore.ts";
import StudentTransferDialog from './dialogs/StudentTransferDialog.vue';

const props = defineProps<{ selectedClass: IClassNode | null }>();
const classStore = useClassStore();

const isEditMode = ref(false);
const gridRef = ref<VxeGridInstance<IStudent>>();
const pasteDialogVisible = ref(false);
const pastedNames = ref('');
const fileInput = ref<HTMLInputElement | null>(null);
const isSubmitting = ref(false);
const selectedRows = ref<IStudent[]>([]);
const showInactive = ref(false);
const transferDialog = reactive({
    visible: false,
    students: [] as IStudent[],
});

const gridOptions = reactive<VxeGridProps<IStudent>>({
  border: true,
  stripe: true,
  height: '100%',
  loading: false,
  columnConfig: { resizable: true },
  keepSource: true,
  data: [],
  columns: [
    { type: 'checkbox', width: 50, fixed: 'left' },
    { type: 'seq', width: 60, title: '#', fixed: 'left' },
    { field: 'name', title: '姓名', sortable: true, editRender: { name: 'input' }, slots: { default: 'name_cell' }, fixed: 'left', width: 120 },
    { field: 'student_no', title: '学号', width: 180, sortable: true, editRender: { name: 'input' } },
    { field: 'is_active', title: '在读状态', width: 100, formatter: ({ cellValue }) => cellValue ? '在读' : '已停用' },
    { title: '操作', width: 180, slots: { default: 'actions_cell' }, fixed: 'right' }
  ],
  editConfig: { trigger: 'click', mode: 'cell', showStatus: true, enabled: false },
  keyboardConfig: { isArrow: true, isEnter: true, isTab: true, isEdit: true },
  mouseConfig: { selected: true },
  checkboxConfig: {
      checkField: 'isChecked',
  },
});

const toggleEditMode = (mode: boolean) => {
  isEditMode.value = mode;
  if (gridOptions.editConfig) gridOptions.editConfig.enabled = mode;
  selectedRows.value = [];
  gridRef.value?.clearCheckboxRow();

  const actionsCol = gridOptions.columns?.find(c => c.slots?.default === 'actions_cell');
  if (actionsCol) actionsCol.visible = !mode;

  const checkboxCol = gridOptions.columns?.find(c => c.type === 'checkbox');
  if(checkboxCol) checkboxCol.visible = !mode;

  gridRef.value?.refreshColumn();

  if (mode) {
    ElMessage.info({ message: '已进入编辑模式，可进行粘贴和修改。', duration: 2000 });
  } else {
    if(props.selectedClass) fetchStudents(props.selectedClass.id);
  }
};

const fetchStudents = async (classId: number) => {
    gridOptions.loading = true;
    try {
        gridOptions.data = await getStudentsByClass(classId, showInactive.value);
        selectedRows.value = [];
        gridRef.value?.clearCheckboxRow();
    } catch (error) {
        gridOptions.data = [];
        ElMessage.error('获取学生列表失败');
    } finally {
        gridOptions.loading = false;
    }
};

const handleFilterChange = () => {
    if (props.selectedClass) {
        fetchStudents(props.selectedClass.id);
    }
};
watch(() => props.selectedClass, (newClass, oldClass) => {
  if (isEditMode.value) {
    ElMessage.warning('请先保存或取消编辑模式，再切换班级！');
    if (oldClass !== undefined && oldClass !== null) {
        classStore.selectNode(oldClass);
    }
    return;
  }
  if (newClass) {
    fetchStudents(newClass.id);
  } else {
    gridOptions.data = [];
  }
}, { immediate: true });

const handleDeactivate = async (student: IStudent) => {
    await ElMessageBox.confirm(`确定要停用学生 "${student.name}" 吗？`, '提示', { type: 'warning' });
    await deactivateStudent(student.id);
    ElMessage.success('学生已停用');
    fetchStudents(props.selectedClass!.id);
};

const handleActivate = async (student: IStudent) => {
    await ElMessageBox.confirm(`确定要重新启用学生 "${student.name}" 吗？`, '提示', { type: 'success', confirmButtonText: '确定启用' });
    await activateStudent(student.id);
    ElMessage.success('学生已激活');
    fetchStudents(props.selectedClass!.id);
};

const handleTransfer = (student: IStudent) => {
    transferDialog.students = [student];
    transferDialog.visible = true;
};

const onConfirmTransfer = async (newClassId: number) => {
    const studentIds = transferDialog.students.map(s => s.id);
    const isBulk = studentIds.length > 1;
    try {
        isBulk
            ? await batchTransferStudents(studentIds, newClassId)
            : await updateStudent(studentIds[0], { class_id: newClassId });
        ElMessage.success('学生转移成功！');
        fetchStudents(props.selectedClass!.id);
    } catch (error) {
        ElMessage.error('学生转移失败');
    } finally {
        transferDialog.visible = false;
        transferDialog.students = [];
    }
};

const onCheckboxChange: VxeTableEvents.CheckboxChange<IStudent> = ({ records }) => {
    selectedRows.value = records;
};

const handleBulkAction = async (action: 'activate' | 'deactivate' | 'transfer') => {
    const studentIds = selectedRows.value.map(s => s.id);
    if (studentIds.length === 0) {
        ElMessage.warning('请至少选择一名学生。');
        return;
    }

    if (action === 'transfer') {
        transferDialog.students = selectedRows.value;
        transferDialog.visible = true;
    } else {
        const isActivating = action === 'activate';
        const actionText = isActivating ? '激活' : '停用';
        const confirmType = isActivating ? 'success' : 'warning';

        try {
            await ElMessageBox.confirm(`确定要批量${actionText}选中的 ${studentIds.length} 名学生吗？`, '批量操作确认', { type: confirmType });
            await batchUpdateStatus(studentIds, isActivating);
            ElMessage.success(`批量${actionText}成功！`);
            fetchStudents(props.selectedClass!.id);
        } catch (error) {
            if (error !== 'cancel') {
                ElMessage.error(`批量${actionText}失败。`);
            }
        }
    }
};

const handleExport = () => {
    const dataToExport = gridOptions.data?.map(({ id, class_id, is_active, ...rest }) => ({
        ...rest,
        '在读状态': is_active ? '在读' : '已停用'
    }));

    const worksheet = XLSX.utils.json_to_sheet(dataToExport || [], {});
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, '学生名单');
    XLSX.writeFile(workbook, `${classStore.gridTitle || '学生名单'}.xlsx`);
    ElMessage.success('导出成功！');
};

const handleOpenPasteDialog = () => {
    pasteDialogVisible.value = true;
};

const handleUploadClick = () => {
    fileInput.value?.click();
};

const handleEditClosed: VxeGridEvents.EditClosed<IStudent> = async ({ row, column }) => {
    const { getUpdateRecords } = gridRef.value!;
    const updatedRecords = getUpdateRecords();
    const isRecordUpdated = updatedRecords.some(record => record.id === row.id);

    if(isRecordUpdated) {
        console.log(`Row ${row.id} at column ${column.field} changed.`);
    }
};

const addRowsToGrid = (newStudents: Partial<IStudentCreate>[]) => {
    gridRef.value?.insertAt(newStudents, -1);
    ElMessage.success(`已成功添加 ${newStudents.length} 名学生到表格，请点击保存。`);
};

const handleSubmitPaste = () => {
    if (!pastedNames.value.trim()) {
        ElMessage.warning('粘贴内容不能为空');
        return;
    }
    const lines = pastedNames.value.trim().split(/[\r\n]+/).filter(line => line.trim() !== '');
    const newStudents = lines.map(line => {
        const parts = line.split(/[\s,，\t]/).map(p => p.trim());
        const name = parts[0] || '';
        const student_no = parts[1] || '';
        return { name, student_no, is_active: true };
    });
    addRowsToGrid(newStudents);
    pasteDialogVisible.value = false;
    pastedNames.value = '';
};

const handleFileSelected = (event: Event) => {
    const target = event.target as HTMLInputElement;
    if (!target.files || target.files.length === 0) return;

    const file = target.files[0];
    const reader = new FileReader();

    reader.onload = (e) => {
        const data = new Uint8Array(e.target!.result as ArrayBuffer);
        const workbook = XLSX.read(data, { type: 'array' });
        const firstSheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[firstSheetName];
        const json = XLSX.utils.sheet_to_json<any>(worksheet);

        const newStudents = json.map(row => ({
            name: row['姓名'] || row['name'] || '',
            student_no: String(row['学号'] || row['student_no'] || ''),
            is_active: true
        }));
        addRowsToGrid(newStudents);
    };

    reader.onerror = () => ElMessage.error('读取文件失败');
    reader.readAsArrayBuffer(file);
    target.value = '';
};

const handleSaveChanges = async () => {
    const { insertRecords, updateRecords } = gridRef.value!.getRecordset();
    if (insertRecords.length === 0 && updateRecords.length === 0) {
        ElMessage.info('没有检测到任何更改。');
        return;
    }

    isSubmitting.value = true;
    try {
        if (insertRecords.length > 0) {
            const newStudents: IStudentCreate[] = insertRecords.map(r => ({
                name: r.name,
                student_no: r.student_no,
                class_id: props.selectedClass!.id
            }));
            await batchCreateStudents({ students: newStudents });
        }
        if (updateRecords.length > 0) {
            for (const student of updateRecords) {
                await updateStudent(student.id, { name: student.name, student_no: student.student_no });
            }
        }
        ElMessage.success('所有更改已成功保存！');
        toggleEditMode(false);
    } catch (error) {
        ElMessage.error('保存更改时出错。');
    } finally {
        isSubmitting.value = false;
    }
};
</script>

<style scoped>
.student-grid-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 1rem;
  box-sizing: border-box;
}
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-shrink: 0;
}
.header-bar h1 {
    font-size: 1.5rem;
    margin: 0;
}
.action-buttons {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}
.grid-wrapper {
  flex-grow: 1;
  min-height: 0;
}
.edit-mode-banner {
  margin-bottom: 1rem;
}
.student-link {
    color: #409EFF;
    text-decoration: none;
    font-weight: 500;
}
.student-link:hover {
    text-decoration: underline;
}
.bulk-actions-bar {
    display: flex;
    align-items: center;
    background-color: #ecf5ff;
    padding: 8px 12px;
    border-radius: 4px;
    margin-bottom: 1rem;
    color: #409eff;
    gap: 8px;
    flex-shrink: 0;
}

.filters-bar {
    margin-bottom: 1rem;
    display: flex;
    align-items: center;
    flex-shrink: 0;
}
</style>