// src/utils/chartOptionFactory.ts

import type { EChartsOption } from 'echarts';

// 为 JSON 数据定义类型
interface HistogramData {
  categories: string[];
  series_data: number[];
  series_name: string;
}

interface HeatmapData {
    x_axis_labels: string[];
    y_axis_labels: string[];
    data: [number, number, number][];
    title: string;
}

interface BoxplotData {
    categories: string[];
    data: number[][];
    title: string;
}

interface ScatterData {
    data: (number | string)[][];
    x_axis_name: string;
    y_axis_name: string;
    title: string;
}

interface RadarData {
    indicator: { name: string; max: number }[];
    series: { name: string; value: number[] }[];
    title: string;
}


// --- 工厂函数 ---


/**
 * 创建学科难度-区分度散点图的 Option
 * @param data - 来自 JSON 的散点图数据
 */
export function createDifficultyDiscriminationScatterOption(data: ScatterData): EChartsOption {
  return {
    title: {
      text: data.title,
      left: 'center',
    },
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        // params.data 的结构是 [难度, 区分度, '学科名']
        const [difficulty, discrimination, subjectName] = params.data;
        return `<strong>${subjectName}</strong><br/>
                ${data.x_axis_name}: ${difficulty.toFixed(3)}<br/>
                ${data.y_axis_name}: ${discrimination.toFixed(3)}`;
      }
    },
    grid: {
      left: '3%',
      right: '8%',
      bottom: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      name: data.x_axis_name,
      nameLocation: 'middle',
      nameGap: 30,
      splitLine: { show: true },
      axisLabel: {
        formatter: '{value}',
      },
    },
    yAxis: {
      type: 'value',
      name: data.y_axis_name,
      nameLocation: 'middle',
      nameGap: 50,
      splitLine: { show: true },
    },
    series: [{
      type: 'scatter',
      data: data.data,
      symbolSize: 18,
      label: {
        show: true,
        position: 'right',
        formatter: (params: any) => params.data[2], // 直接显示学科名
        fontSize: 12,
        color: '#333'
      },
      // markLine (标线) 必须定义在 series 内部
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: {
          type: 'dashed',
          color: '#999'
        },
        label: {
          position: 'end',
          formatter: '{b}',
          color: '#999'
        },
        data: [
          // 垂直于 X 轴的参考线
          {
            name: '中等难度',
            xAxis: 0.5
          },
          // 垂直于 Y 轴的参考线
          {
            name: '良好区分度',
            yAxis: 0.4
          }
        ]
      }
    }],
  };
}

/**
 * 创建分数分布直方图的 Option
 * @param data - 来自 JSON 的直方图数据
 */
export function createHistogramOption(data: HistogramData): EChartsOption {
  return {
    title: {
      text: data.series_name,
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: data.categories,
      axisLabel: {
        interval: 0,
        rotate: 30
      }
    },
    yAxis: {
      type: 'value',
      name: '人数'
    },
    series: [
      {
        name: data.series_name,
        type: 'bar',
        data: data.series_data,
        barWidth: '60%',
        label: {
          show: true,
          position: 'top',
          formatter: (params: any) => (params.value > 0 ? params.value : '')
        }
      }
    ]
  };
}


/**
 * 创建班级对比雷达图的 Option
 * @param data - 来自 JSON 的雷达图数据
 */
export function createClassRadarOption(data: RadarData): EChartsOption {
  return {
    title: {
      text: data.title,
      left: 'center',
      top: '5%',
    },
    tooltip: {
      trigger: 'item',
    },
    legend: {
      bottom: '5%',
      left: 'center',
    },
    radar: {
      indicator: data.indicator,
      radius: '60%',
      center: ['50%', '55%'],
      splitArea: {
        areaStyle: {
            color: ['rgba(114, 172, 209, 0.2)', 'rgba(114, 172, 209, 0.4)'],
        }
      },
    },
    series: [{
      type: 'radar',
      data: data.series.map(s => ({
          ...s,
          symbol: 'circle',
          symbolSize: 6
      })),
      areaStyle: {
        opacity: 0.2
      },
      lineStyle: {
        width: 3
      }
    }],
  };
}

/**
 * 创建学科相关性热力图的 Option
 * @param data - 来自 JSON 的热力图数据
 */
export function createHeatmapOption(data: HeatmapData): EChartsOption {
    return {
        title: {
            text: data.title,
            left: 'center'
        },
        tooltip: {
            position: 'top'
        },
        grid: {
            height: '60%',
            top: '20%'
        },
        xAxis: {
            type: 'category',
            data: data.x_axis_labels,
            splitArea: {
                show: true
            }
        },
        yAxis: {
            type: 'category',
            data: data.y_axis_labels,
            splitArea: {
                show: true
            }
        },
        visualMap: {
            min: -1,
            max: 1,
            calculable: true,
            orient: 'horizontal',
            left: 'center',
            bottom: '5%',
            inRange: {
                color: ['#313695', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026']
            }
        },
        series: [{
            name: '相关系数',
            type: 'heatmap',
            data: data.data,
            label: {
                show: true,
                formatter: params => (params.value as any)[2].toFixed(3)
            },
            emphasis: {
                itemStyle: {
                    shadowBlur: 10,
                    shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
            }
        }]
    };
}

/**
 * 创建成绩分布箱线图的 Option
 * @param data - 来自 JSON 的箱线图数据
 */
export function createBoxplotOption(data: BoxplotData): EChartsOption {
    return {
        title: {
            text: data.title,
            left: 'center'
        },
        tooltip: {
            trigger: 'item',
            axisPointer: {
                type: 'shadow'
            }
        },
        grid: {
            left: '10%',
            right: '10%',
            bottom: '15%'
        },
        xAxis: {
            type: 'category',
            data: data.categories,
            boundaryGap: true,
            nameGap: 30,
        },
        yAxis: {
            type: 'value',
            name: '分数',
            splitArea: {
                show: true
            }
        },
        series: [{
            name: '成绩分布',
            type: 'boxplot',
            data: data.data,
        }]
    };
}