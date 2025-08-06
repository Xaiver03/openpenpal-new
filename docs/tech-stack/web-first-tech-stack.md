## ✅ 一、前端技术栈（Web App）

| 模块    | 技术选型                                           | 理由                 |
| ----- | ---------------------------------------------- | ------------------ |
| 框架    | **React + Vite**（推荐）或 Vue 3 + Vite             | 快速构建、良好生态、支持组件拆分   |
| 状态管理  | **Zustand（React）** 或 Pinia（Vue）                | 简洁轻量，支持模块解耦        |
| UI 框架 | **TailwindCSS** + Headless UI / ShadCN         | 支持响应式设计 + 良好的样式自由度 |
| 动画与动效 | Framer Motion / CSS Transitions                | 提供温柔信封动效与交互反馈      |
| 二维码生成 | `qrcode` npm 包                                 | 生成贴纸二维码并导出图片       |
| 文件上传  | `react-dropzone` / 原生 `input` + Axios FormData | 实现上传手写照片并压缩处理      |
| 富文本输入 | `react-quill` / `tiptap`                       | 写信草稿编辑器（用于手抄参考）    |
| 国际化   | `i18next`（如考虑中英双语）                             | 可选扩展功能             |
| 路由管理  | `React Router v6` / `Vue Router 4`             | 支持动态路由、回信逻辑跳转      |

---

## ✅ 二、后端技术栈（Node 生态）

|模块|技术选型|理由|
|---|---|---|
|应用框架|**NestJS** / Express + TypeScript|模块化强、接口清晰、支持RBAC权限体系|
|ORM/数据库|**Prisma + PostgreSQL**|支持关系型结构（编号/信件/用户/照片）且迁移清晰|
|鉴权系统|JWT + Cookie / Session（选择其一）|区分用户/信使权限|
|文件存储|腾讯云 COS / 阿里云 OSS|存储上传的照片与贴纸生成图|
|日志系统|`winston` / `pino`|接入后台编号操作、任务链路日志|
|内容安全|腾讯云内容安全 API|审核照片/OCR文字内容（如开放展示）|

---

## ✅ 三、后台系统（管理台）

|模块|技术建议|
|---|---|
|前端|Vue 3 + Naive UI 或 Ant Design Vue|
|权限系统|多角色控制（管理员 / 城市总代 / 内容审核人）|
|任务面板|编号状态追踪 + 投递记录表格视图|
|审核系统|展示照片内容 + 审核通过/驳回操作面板|

---

## ✅ 四、DevOps & 部署建议

|项目|推荐方案|
|---|---|
|构建系统|Vite + CI (GitHub Actions / GitLab CI)|
|自动部署|Railway / Vercel / 腾讯云轻量部署（早期版本）|
|数据备份|PostgreSQL 定时快照 + 异地存储|
|日志监控|Sentry + PM2 或 Docker Logs|

---

## ✅ 五、可选增强项（未来可接入）

|模块|技术建议|说明|
|---|---|---|
|OCR识别|腾讯云 OCR SDK|自动识别手写内容（若用户上传）|
|小程序迁移|Taro / UniApp|若未来计划微信小程序版，同构能力强|
|PWA 支持|Workbox + manifest.json|提供“添加至主屏幕”体验|
|AI推荐|ChatGPT API + 标签匹配算法|公共信墙内容智能推荐（后期迭代）|
