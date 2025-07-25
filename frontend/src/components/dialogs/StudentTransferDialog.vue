<!--src/components/dialogs/StudentTransferDialog.vue-->
<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    width="400px"
    :close-on-click-modal="false"
  >
    <p>请为选中的学生选择目标班级：</p>
    <el-tree
        ref="classTreeRef"
        :data="classStore.classTree"
        :props="classStore.treeProps"
        :expand-on-click-node="false"
        highlight-current
        node-key="id"
        class="scope-tree"
        @node-click="handleNodeClick"
    />
    <p v-if="selectedNode">
      已选择: <strong>{{ selectedNode.displayName }}</strong>
    </p>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button
        type="primary"
        @click="handleSubmit"
        :disabled="!selectedNode || !!selectedNode.classes"
      >
        确定转移
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { ElDialog, ElButton, ElTree, ElMessage } from 'element-plus';
import { useClassStore } from '@/stores/classStore.ts';
import type {IClassNode, IGradeNode} from '@/types/dataModels.ts';

interface NodeData extends IClassNode {
    displayName: string;
    classes?: NodeData[]; // 年级节点会有这个属性
}

const props = defineProps<{
  modelValue: boolean;
  studentCount: number;
}>();

const emit = defineEmits(['update:modelValue', 'confirm']);

const classStore = useClassStore();
const selectedNode = ref<NodeData | null>(null);

const title = computed(() => `转移 ${props.studentCount} 名学生`);

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});


const handleNodeClick = (nodeData: IGradeNode | IClassNode) => {
  classStore.selectNode(nodeData);
};

const handleSubmit = () => {
    if (selectedNode.value) {
        emit('confirm', selectedNode.value.id);
        dialogVisible.value = false;
        selectedNode.value = null;
    } else {
        ElMessage.warning('请选择一个目标班级。');
    }
}
</script>

<style scoped>
.scope-tree {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 8px;
  margin-top: 1rem;
}
</style>