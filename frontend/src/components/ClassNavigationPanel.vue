<template>
  <div class="class-navigation-panel" v-loading="classStore.isLoading || classStore.isSubmitting">
    <div class="aside-header">
      <h3>班级与年级导航</h3>
      <el-button
        type="primary"
        :icon="Plus"
        size="small"
        circle
        plain
        @click="gradeDialogVisible = true"
        title="新建一个年级"
      />
    </div>

    <el-scrollbar class="tree-scrollbar">
      <div v-if="!classStore.isLoading && classStore.classTree.length === 0" class="empty-state">
        <p>系统中还没有任何年级信息。</p>
        <p>请点击右上角“+”按钮添加。</p>
      </div>

      <el-tree
        v-else
        ref="treeRef"
        :data="classStore.classTree"
        :props="classStore.treeProps"
        @node-click="handleNodeClick"
        highlight-current
        :expand-on-click-node="false"
        class="class-tree"
        node-key="id"
        :current-node-key="currentNodeKeyForTree"
        :default-expand-all="true"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <span class="node-label">
              <el-icon v-if="isGradeNode(data)"><CollectionTag /></el-icon>
              <el-icon v-else><Avatar /></el-icon>
              <span>{{ node.label }}</span>
            </span>

            <el-dropdown trigger="click" @command="(command) => handleCommand(command, data)" @visible-change="(v) => handleDropdownVisible(v, data.id)">
              <el-button
                :icon="MoreFilled"
                size="small"
                circle
                plain
                class="more-button"
                :class="{ 'is-active': activeDropdown === data.id }"
                @click.stop
              />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="isGradeNode(data)" command="addClass" :icon="Plus">新建班级</el-dropdown-item>
                  <el-dropdown-item command="edit" :icon="Edit">编辑名称</el-dropdown-item>
                  <el-dropdown-item command="delete" :icon="Delete" divided class="delete-item">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>

          </span>
        </template>
      </el-tree>
    </el-scrollbar>
  </div>

  <GradeCreateDialog v-model="gradeDialogVisible" @success="classStore.fetchClassTree()" />
  <ClassCreateDialog v-model="classDialogVisible" :target-grade="targetGradeForDialog" @success="classStore.fetchClassTree()" />
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import {
  ElButton, ElTree, ElIcon, ElScrollbar, vLoading, ElDropdown,
  ElDropdownMenu, ElDropdownItem, ElMessage, ElMessageBox
} from 'element-plus';
import { Plus, CollectionTag, Avatar, MoreFilled, Edit, Delete } from '@element-plus/icons-vue';
import { useClassStore } from '@/stores/classStore';
import type { IClassNode, IGradeNode } from '@/types/dataModels';
import GradeCreateDialog from './dialogs/GradeCreateDialog.vue';
import ClassCreateDialog from './dialogs/ClassCreateDialog.vue';

const classStore = useClassStore();
const treeRef = ref<InstanceType<typeof ElTree>>();

const currentNodeKeyForTree = computed(() => classStore.selectedNodeKey ?? undefined);

onMounted(() => {
  if (classStore.classTree.length === 0) {
    classStore.fetchClassTree();
  }
});

const handleNodeClick = (nodeData: IGradeNode | IClassNode) => {
  classStore.selectNode(nodeData);
};

// --- Dialogs Logic ---
const gradeDialogVisible = ref(false);
const classDialogVisible = ref(false);
const targetGradeForDialog = ref<IGradeNode | null>(null);

const handleOpenClassDialog = (gradeData: IGradeNode) => {
  targetGradeForDialog.value = gradeData;
  classDialogVisible.value = true;
};

// --- Dropdown & Actions Logic ---
const activeDropdown = ref<number | null>(null);

const handleDropdownVisible = (visible: boolean, nodeId: number) => {
  activeDropdown.value = visible ? nodeId : null;
};

const isGradeNode = (nodeData: any): nodeData is IGradeNode => 'classes' in nodeData;

const handleCommand = (command: string, nodeData: IGradeNode | IClassNode) => {
  switch(command) {
    case 'addClass':
      handleOpenClassDialog(nodeData as IGradeNode);
      break;
    case 'edit':
      handleEdit(nodeData);
      break;
    case 'delete':
      handleDelete(nodeData);
      break;
  }
};

const handleEdit = async (nodeData: IGradeNode | IClassNode) => {
  const isGrade = isGradeNode(nodeData);
  const entityType = isGrade ? '年级' : '班级';

  try {
    const { value } = await ElMessageBox.prompt(`请输入新的${entityType}名称`, `编辑${entityType}`, {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: nodeData.name,
      inputValidator: (val) => val && val.trim() ? true : '名称不能为空',
    });

    let success = false;
    if (isGrade) {
      success = await classStore.handleUpdateGrade(nodeData.id, { name: value });
      if (success) {
        ElMessage.success('年级名称已更新');
      }
    } else {
      success = await classStore.handleUpdateClass(nodeData.id, { name: value });
      if(success) {
        ElMessage.success('班级名称已更新');
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error);
    }
  }
};

const handleDelete = async (nodeData: IGradeNode | IClassNode) => {
  const isGrade = isGradeNode(nodeData);
  const entityType = isGrade ? '年级' : '班级';
  const message = isGrade
    ? `确定删除年级 “${nodeData.name}” 吗？这将同时删除其下所有（无学生的）班级，此操作极其危险且不可撤销！`
    : `确定删除班级 “${nodeData.name}” 吗？只有班级内没有学生时才能删除。`;

  try {
    await ElMessageBox.confirm(message, '严重警告', {
      confirmButtonText: '我已知晓风险，确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    });

    let success = false;
    if (isGrade) {
      success = await classStore.handleDeleteGrade(nodeData.id);
      if (success) {
        ElMessage.success('年级已删除');
      }
    } else {
      success = await classStore.handleDeleteClass(nodeData.id);
      if(success) {
        ElMessage.success('班级已删除');
      }
    }
  } catch (error) {
     if(error !== 'cancel') {
      console.error(error);
    }
  }
};
</script>

<style scoped>
.class-navigation-panel { height: 100%; display: flex; flex-direction: column; background-color: #fff; border-right: 1px solid #e4e7ed; }
.aside-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; border-bottom: 1px solid #e4e7ed; flex-shrink: 0; }
.aside-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: #303133; }
.tree-scrollbar { flex-grow: 1; }
.empty-state { padding: 32px; text-align: center; color: #909399; font-size: 14px; line-height: 1.6; }
.class-tree { padding: 8px 0; background-color: transparent; }
.class-tree :deep(.el-tree-node__content) { height: 40px; }
.custom-tree-node {
  flex-grow: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
}
.node-label { display: flex; align-items: center; gap: 8px; }
.more-button {
  visibility: hidden;
  border: none;
}
.el-tree-node:hover .more-button,
.more-button.is-active {
  visibility: visible;
}
.delete-item {
  color: #F56C6C;
}
.delete-item:hover {
  color: #fff;
  background-color: #F56C6C;
}
</style>