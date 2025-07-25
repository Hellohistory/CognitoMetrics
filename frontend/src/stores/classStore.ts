// src/stores/classStore.ts
import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import {
    getClassTree,
    createGrade,
    createClass,
    updateClass,
    deleteClass,
    updateGrade,
    deleteGrade
} from '@/api/classApi';
import type {
    IGradeNode,
    IClassNode,
    IGradeCreate,
    IClassCreate,
    IClassUpdate,
    IGradeUpdate
} from '@/types/dataModels';

export const useClassStore = defineStore('class', () => {
  const classTree = ref<IGradeNode[]>([]);
  const isLoading = ref(false);
  const isSubmitting = ref(false);
  const selectedGradeId = ref<number | null>(null);
  const selectedClassId = ref<number | null>(null);


  const treeProps = {
    children: 'classes',
    label: 'name',
  };

  const selectedGrade = computed<IGradeNode | null>(() => {
    if (!selectedGradeId.value) return null;
    return classTree.value.find(g => g.id === selectedGradeId.value) || null;
  });

  const selectedClass = computed<IClassNode | null>(() => {
    if (!selectedClassId.value || !selectedGrade.value) return null;
    return selectedGrade.value.classes.find(c => c.id === selectedClassId.value) || null;
  });

  const selectedNodeKey = computed(() => {
      return selectedClassId.value ?? selectedGradeId.value;
  });

  const gridTitle = computed(() => {
    if (selectedClass.value) {
      return `${selectedGrade.value?.name || ''} - ${selectedClass.value.name} - 学生名单`;
    }
    if (selectedGrade.value) {
      return `${selectedGrade.value.name} - 年级总览`;
    }
    return '请从左侧选择一个班级或年级';
  });

  // --- Actions ---
  const fetchClassTree = async () => {
    isLoading.value = true;
    try {
      classTree.value = await getClassTree();
    } finally {
      isLoading.value = false;
    }
  };

  const selectNode = (node: IGradeNode | IClassNode) => {
    const isGrade = 'classes' in node;
    if (isGrade) {
      selectedGradeId.value = node.id;
      selectedClassId.value = null;
    } else {
      const parentGrade = classTree.value.find(g => g.classes.some(c => c.id === node.id));
      selectedGradeId.value = parentGrade?.id || null;
      selectedClassId.value = node.id;
    }
  };

  const handleCreateGrade = async (gradeData: IGradeCreate): Promise<boolean> => {
    isSubmitting.value = true;
    try {
      await createGrade(gradeData);
      await fetchClassTree();
      return true;
    } finally {
      isSubmitting.value = false;
    }
  };

  const handleCreateClass = async (classData: IClassCreate): Promise<boolean> => {
    isSubmitting.value = true;
    try {
      await createClass(classData);
      await fetchClassTree();
      return true;
    } finally {
      isSubmitting.value = false;
    }
  };

  const handleUpdateClass = async (classId: number, classData: IClassUpdate): Promise<boolean> => {
      isSubmitting.value = true;
      try {
          await updateClass(classId, classData);
          await fetchClassTree();
          return true;
      } catch (error) {
          return false;
      } finally {
          isSubmitting.value = false;
      }
  };

  const handleDeleteClass = async (classId: number): Promise<boolean> => {
      isSubmitting.value = true;
      try {
          await deleteClass(classId);
          if(selectedClassId.value === classId) {
              selectedClassId.value = null;
          }
          await fetchClassTree();
          return true;
      } catch (error) {
          return false;
      } finally {
          isSubmitting.value = false;
      }
  };

  const handleUpdateGrade = async (gradeId: number, gradeData: IGradeUpdate): Promise<boolean> => {
    isSubmitting.value = true;
    try {
        await updateGrade(gradeId, gradeData);
        await fetchClassTree();
        return true;
    } catch(error) {
        return false;
    } finally {
        isSubmitting.value = false;
    }
  };

  const handleDeleteGrade = async (gradeId: number): Promise<boolean> => {
      isSubmitting.value = true;
      try {
          await deleteGrade(gradeId);
          if(selectedGradeId.value === gradeId) {
              selectedGradeId.value = null;
              selectedClassId.value = null;
          }
          await fetchClassTree();
          return true;
      } catch (error) {
          return false;
      } finally {
          isSubmitting.value = false;
      }
  };

  return {
    classTree,
    isLoading,
    isSubmitting,
    selectedGrade,
    selectedClass,
    selectedNodeKey,
    treeProps,
    gridTitle,
    fetchClassTree,
    selectNode,
    handleCreateGrade,
    handleCreateClass,
    handleUpdateClass,
    handleDeleteClass,
    handleUpdateGrade,
    handleDeleteGrade,
  };
});