/**
 * @file src/main.ts
 * @description Vue 应用入口，创建应用并注册全局插件：Pinia、ElementPlus、VXE-Table 及其 UI 组件库
 * @author Hellohistory
 */

/**
 * @module main
 * 高内聚：仅负责应用初始化与插件注册
 * 低耦合：所有业务逻辑、组件、路由等在其他模块中维护
 */

import { createApp, type App as VueApp } from 'vue';
import { createPinia } from 'pinia';

import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';

import App from './App.vue';
import router from './router';

import VxeUI from 'vxe-pc-ui';
import 'vxe-pc-ui/lib/style.css';
import VXETable from 'vxe-table';
import 'vxe-table/lib/style.css';

/**
 * 注册并配置 VXE-Table 插件
 *
 * 由于官方类型定义与 Vxe-UI 存在冲突，使用 `as any` 临时绕过类型检查，
 * 确保插件能够正常安装与使用。
 *
 * @param app - Vue 应用实例
 */
function useTable(app: VueApp): void {
  VXETable.use(VxeUI as any);
  app.use(VXETable);
}


const app = createApp(App);

app.use(createPinia());
app.use(router);
app.use(ElementPlus);
// 注册 VXE-Table 插件
app.use(useTable);

// 挂载应用
app.mount('#app');
