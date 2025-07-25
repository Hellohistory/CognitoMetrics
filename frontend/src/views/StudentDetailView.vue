<template>
    <div class="student-detail-view" v-loading="isLoading">
        <template v-if="student">
            <el-page-header @back="goBack" class="page-header">
                <template #content>
                    <span class="text-large font-600 mr-3"> 学生个人档案 </span>
                </template>
            </el-page-header>

            <el-descriptions :column="3" border class="profile-descriptions">
                <el-descriptions-item label="姓名">{{ student.name }}</el-descriptions-item>
                <el-descriptions-item label="学号">{{ student.student_no }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                    <el-tag :type="student.is_active ? 'success' : 'info'">
                        {{ student.is_active ? '在读' : '已停用' }}
                    </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="年级">{{ student.grade_name }}</el-descriptions-item>
                <el-descriptions-item label="班级">{{ student.class_name }}</el-descriptions-item>
                <el-descriptions-item label="入学年份">{{ student.enrollment_year }}</el-descriptions-item>
            </el-descriptions>

            <el-card class="chart-card" v-if="performanceHistory">
                <template #header>历次考试成绩趋势</template>
                <VueEcharts v-if="performanceHistory.records.length > 0" :option="chartOption" style="height: 400px;" />
                <el-empty v-else description="暂无该学生的考试成绩记录"></el-empty>
            </el-card>

        </template>
        <el-empty v-else-if="!isLoading" description="无法加载学生信息"></el-empty>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { ElPageHeader, ElDescriptions, ElDescriptionsItem, ElTag, ElCard, ElEmpty, vLoading } from 'element-plus';
import type { EChartsOption } from 'echarts';
import { getStudentDetails, getStudentPerformanceHistory } from '@/api/studentApi';
import type { IStudentDetail, IStudentPerformanceHistory } from '@/types/dataModels';
import VueEcharts from '@/components/charts/VueEcharts.vue';

const props = defineProps<{
    id: string;
}>();

const router = useRouter();
const isLoading = ref(true);
const student = ref<IStudentDetail | null>(null);
const performanceHistory = ref<IStudentPerformanceHistory | null>(null);

onMounted(async () => {
    const studentId = parseInt(props.id, 10);
    if (isNaN(studentId)) return;

    try {
        const [details, history] = await Promise.all([
            getStudentDetails(studentId),
            getStudentPerformanceHistory(studentId)
        ]);
        student.value = details;
        performanceHistory.value = history;
    } catch (error) {
        console.error("加载学生详细信息失败", error);
    } finally {
        isLoading.value = false;
    }
});

const chartOption = computed<EChartsOption>(() => {
    const records = performanceHistory.value?.records || [];
    const examNames = records.map(r => r.exam_name);
    const scores = records.map(r => r.total_score);
    const gradeRanks = records.map(r => r.grade_rank);

    return {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: ['总分', '年级排名']
        },
        xAxis: {
            type: 'category',
            data: examNames
        },
        yAxis: [
            {
                type: 'value',
                name: '分数',
                position: 'left',
            },
            {
                type: 'value',
                name: '排名',
                position: 'right',
                inverse: true // 排名越小越好，所以Y轴反向
            }
        ],
        series: [
            {
                name: '总分',
                type: 'bar',
                yAxisIndex: 0,
                data: scores
            },
            {
                name: '年级排名',
                type: 'line',
                yAxisIndex: 1,
                data: gradeRanks
            }
        ]
    };
});

const goBack = () => {
    router.back();
}
</script>

<style scoped>
.student-detail-view {
    padding: 1rem 2rem;
}
.page-header {
    margin-bottom: 1.5rem;
}
.profile-descriptions {
    margin-bottom: 2rem;
}
.chart-card {
    border-radius: 8px;
}
</style>