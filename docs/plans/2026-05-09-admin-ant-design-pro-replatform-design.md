# 2026-05-09 管理台改造设计：基于 ant-design-pro 的重平台化替换

## 背景

当前项目的管理台位于 `web/admin`，已经具备规则分类、规则管理、检测任务、风险结果、文稿中心、整改处置、操作日志等核心业务能力，但存在以下问题：

- 前端壳层和页面组织来自多轮自定义演进，结构不再统一，后续维护成本偏高。
- 当前实现虽然已经使用 `antd` / `@ant-design/pro-components`，但整体并不是标准 `ant-design-pro` 工程结构。
- 业务路由、页面布局、工作台状态管理均由项目自行维护，不利于后续在企业后台范式内继续扩展。
- 项目需要更贴近中国政企后台的视觉语言与交互习惯，同时保留全部既有业务功能。

用户已经明确本次改造的约束：

- 前端技术路线采用 `ant-design-pro`。
- 信息架构允许按 `ant-design-pro` 方式重组，不要求保持旧路由不变。
- 页面视觉可借鉴 `iview-admin` 的企业后台风格，尤其是暗色左侧导航和标签化工作台。
- 历史 React 管理台不再以目录形式保留，迁移完成后仓库中只保留新的 `web/admin`；旧版本通过 Git tag 留档。
- 不要求兼容 IE，只需要覆盖较老但仍属现代内核范围的 Chrome / Edge / 国产浏览器极速模式。

## 目标

1. 用标准 `ant-design-pro` 工程结构重建 `web/admin`。
2. 保留当前全部核心业务能力与后端接口契约。
3. 重构后台信息架构与导航体系，使其更符合企业级管理台的认知方式。
4. 借鉴 `iview-admin` 的标签化工作台体验，实现“打开新页面不整页刷新”的多页签交互。
5. 采用更稳重的中国企业后台视觉风格，左侧导航主色明确为 `#191a23`。
6. 在迁移完成后彻底删除旧 `web/admin` 代码，仅通过 Git tag 回溯历史实现。

## 非目标

本次改造不包含以下内容：

- 后端 API 的大规模重写。
- 登录体系改成前端账号密码模式；仍沿用当前基于 cookie 与后端跳转的登录流。
- 为兼容 IE / IE Mode 而降级整个技术栈。
- 为了复刻 `iview-admin` 而引入 Vue 或 `View UI`。
- 完整照搬 `iview-admin` 的页面、组件或旧时代视觉细节。

## 历史保留策略

旧版管理台不再迁入 `archive/` 或 `legacy/` 目录。历史保留策略改为版本保留：

1. 在删除旧版 `web/admin` 之前，对旧 React 管理台当前状态创建 annotated tag。
2. 推荐 tag 名称为：`archive/admin-react-legacy-2026-05-09`。
3. 迁移完成后，主分支目录中仅保留新的 `web/admin`。
4. 后续如需回看旧实现，通过 `git checkout archive/admin-react-legacy-2026-05-09` 完成。

这样可以保证仓库树保持纯净，同时保留完整历史上下文。

## 技术路线

### 总体方案

新的 `web/admin` 采用 `ant-design-pro` 官方常见分层：

- `Umi Max` 负责路由、运行时配置、构建与请求基础设施。
- `ant-design-pro` 负责布局骨架、菜单、页面容器与企业后台工程范式。
- `Ant Design` / `ProComponents` 负责表格、表单、描述区、详情布局、弹窗和状态组件。
- 业务请求层尽量复用现有 `services` 语义，避免改动后端契约。

### 浏览器兼容基线

本次前端兼容基线定义为：

- Chrome 88+
- Edge 88+
- 国产浏览器极速模式，前提是内核大体落在上述代际

明确不支持：

- Internet Explorer
- 兼容模式 / IE Mode
- 明显早于上述代际的老旧浏览器内核

为了提升较老现代浏览器的兼容性，需要在 `Umi` / `antd` 层启用兼容配置：

- 在 Umi 中显式配置 `targets`
- 在 antd 中启用兼容型 `StyleProvider`
- 自定义样式避免引入额外激进的新 CSS 特性

## 新工程目录结构

新管理台按照 `ant-design-pro` 风格组织，建议结构如下：

```text
web/admin/
  config/
    config.ts
    routes.ts
    defaultSettings.ts
    proxy.ts
  src/
    app.tsx
    access.ts
    components/
      PageTabs/
      StatusTag/
      SnapshotViewer/
      HitPreview/
      HtmlArticleEditor/
    services/
      auth.ts
      orgs.ts
      categories.ts
      keywords.ts
      tasks.ts
      results.ts
      articles.ts
      logs.ts
    pages/
      User/
        LoginRedirect/
      Inspection/
        TaskList/
        TaskCreate/
        TaskDetail/
        ResultList/
      Rules/
        CategoryList/
        KeywordList/
      Content/
        ArticleList/
        ArticleDetail/
        ArticleRectify/
      Audit/
        OperationLogList/
```

原则：

- `config/routes.ts` 只描述信息架构与菜单。
- `src/app.tsx` 负责全局请求、初始化状态、布局运行时配置。
- `src/access.ts` 负责会话级访问控制。
- `src/services/*` 保留现有接口语义，避免前后端契约漂移。
- `src/components/*` 沉淀跨页面复用的业务组件。
- `src/pages/*` 只承载具体页面编排。

## 信息架构与路由重组

本次接受按 `ant-design-pro` 的方式重组路由，只要求功能完整，不要求主入口路径完全沿用旧结构。

### 一级导航分组

建议重组为四个一级业务域：

- 巡检业务
- 规则中心
- 内容中心
- 审计留痕

### 主路由设计

```text
/                               -> /inspection/tasks
/user/login                     -> 登录引导页（非本地账号密码页）
/inspection/tasks               -> 任务列表
/inspection/tasks/create        -> 新建任务
/inspection/tasks/:taskId       -> 任务详情
/inspection/results             -> 全局结果列表
/rules/categories               -> 规则分类
/rules/keywords                 -> 规则管理
/content/articles               -> 文稿中心
/content/articles/:articleId    -> 文稿详情
/content/articles/:articleId/rectify -> 内容整改
/audit/logs                     -> 操作日志
```

### 旧路由兼容策略

为了降低迁移切换成本，保留轻量级重定向：

- `/tasks` -> `/inspection/tasks`
- `/tasks/new` -> `/inspection/tasks/create`
- `/articles` -> `/content/articles`
- `/logs` -> `/audit/logs`
- 其他常见旧入口按需映射到对应新路由

旧路径兼容仅作为过渡措施；菜单、面包屑与主入口统一以新路由体系为准。

## 标签化工作台设计

### 目标

用户希望保留接近 `iview-admin` 的标签化工作台体验：打开新页面时不整页刷新，并能在多个业务页面之间快速切换。

### 设计原则

- 仍然以 `ant-design-pro` 的 `ProLayout` 作为基础壳层。
- 在内容区顶部增加自定义标签栏，而不是完全自造一套 layout 框架。
- 路由仍使用标准 Umi 路由，避免与页面状态机深度耦合。

### 标签栏行为

标签栏至少支持：

- 点击菜单打开新标签
- 同一路由重复进入时不重复创建标签，而是切换到已有标签
- 关闭当前标签
- 关闭其他标签
- 关闭全部标签
- 刷新当前标签
- 记住最近打开的标签集合与当前活动标签

### 状态策略

- 标签页集合：本地持久化，按 `orgid` 隔离
- 页面筛选条件：继续保留在 URL query 中
- 页内草稿态：仅在确有业务必要的页面保留，例如整改页

这样可以兼顾：

- `ant-design-pro` 的标准路由能力
- `iview-admin` 风格的多页签工作方式
- 可分享、可回退、可刷新恢复的 URL 行为

## 视觉与主题设计

### 总体风格

视觉方向不是复刻 `iview-admin` 的老组件实现，而是吸收其在中国企业后台中常见的审美特征：

- 稳重、克制、信息密度高
- 深色左侧导航 + 白色业务区
- 蓝灰主色调
- 低装饰、强分层、重表格与表单可读性

### 关键主题决策

- 左侧导航主底色：`#191a23`
- 内容区背景：浅灰白体系，如 `#f5f7fa` / `#f0f2f5`
- 卡片背景：纯白
- 主色：企业蓝，但控制饱和度，避免互联网产品感过强
- 边框：冷灰细边框，弱化大阴影和大圆角

### 布局语言

- 使用深色侧栏、浅色内容区、标准顶栏和面包屑
- 统一列表页的“标题 + 查询区 + 表格区 + 批量操作”结构
- 详情页优先使用 `Card`、`Descriptions`、`Tabs`、`ProDescriptions` 等标准布局
- 审计与日志页面采用更克制的表达，突出筛选与状态流转

### 明确避免

- 不保留当前中性灰黑、实验性较强的自定义工作台视觉语言
- 不使用过强的渐变、大圆角、大阴影
- 不照搬 `iview-admin` 的旧时代 UI 细节
- 不做演示站式视觉堆砌

## 会话、权限与登录流

当前项目的认证方式已经存在明确契约：

- 通过 `/api/v1/auth/session` 获取当前登录态
- 未登录时跳转 `/auth/login`
- 登出后仍跳回 `/auth/login`

新管理台继续沿用该机制，不引入本地账号密码登录页。

### 运行时登录处理

- `src/app.tsx` 中初始化会话信息
- 如果会话接口失败或返回 401，则跳转 `/auth/login`
- `user/login` 页面仅作为路由兼容与登录引导页，不承载真正表单
- `src/access.ts` 根据当前会话决定页面可见性与访问能力

### 组织维度

现有管理台已经以会话中的 `orgid` / `orgname` 作为活跃机构来源；新管理台延续该模型。

- 页面默认以上下文中的活跃机构作为业务作用域
- 标签状态、部分缓存状态按 `orgid` 隔离
- 不再把机构 ID 暴露成可随意编辑的表单输入

## 数据层复用策略

当前 `web/admin/src/services/*` 已经覆盖主要业务能力。迁移时尽量复用这些接口语义：

- `auth.ts`
- `orgs.ts`
- `categories.ts`
- `keywords.ts`
- `tasks.ts`
- `results.ts`
- `articles.ts`
- `logs.ts`

迁移原则：

- 优先保留请求路径、参数结构、响应结构不变
- 仅在新框架接入时调整调用位置与类型组织
- 只有当旧实现明显依赖旧壳层约束时，才对调用层进行轻量改写

## 页面级功能映射

### 巡检业务

1. 任务列表
   - 现状：支持分页、筛选、删除、进入结果页
   - 新版：迁移到 `/inspection/tasks`
   - 表达方式：标准 `ProTable` + 查询表单 + 批量/单项操作

2. 新建任务
   - 现状：支持按机构、新规则、时间范围创建任务
   - 新版：迁移到 `/inspection/tasks/create`
   - 表达方式：`ProForm` + 标准提交反馈

3. 任务详情
   - 现状：展示任务摘要、命中结果、快照、关联日志
   - 新版：迁移到 `/inspection/tasks/:taskId`
   - 表达方式：摘要卡 + Tabs + 描述区

4. 全局结果列表
   - 现状：支持批量下线处置、跳转文稿详情和整改页
   - 新版：迁移到 `/inspection/results`
   - 表达方式：列表工作区 + 批量操作工具栏

### 规则中心

1. 规则分类
   - 现状：支持新增、编辑、删除、启停、排序
   - 新版：迁移到 `/rules/categories`

2. 规则管理
   - 现状：支持筛选、增删改、分类归属、风险等级、作用范围
   - 新版：迁移到 `/rules/keywords`

### 内容中心

1. 文稿中心
   - 现状：支持按标题 / 文稿 ID 查询，进入详情
   - 新版：迁移到 `/content/articles`

2. 文稿详情
   - 现状：查看文章信息、命中记录、操作日志、字段变更
   - 新版：迁移到 `/content/articles/:articleId`

3. 内容整改
   - 现状：支持修改标题、摘要、正文，保存整改，提交复核
   - 新版：迁移到 `/content/articles/:articleId/rectify`
   - 继续保留草稿暂存与返回路径语义

### 审计留痕

1. 操作日志
   - 现状：支持按文章、任务、操作人筛选，并查看请求快照
   - 新版：迁移到 `/audit/logs`
   - 表达方式：克制的审计列表 + 快照查看器

## 迁移步骤

建议按以下顺序执行：

1. 在旧版 `web/admin` 当前状态上创建 annotated tag：`archive/admin-react-legacy-2026-05-09`
2. 删除旧版 `web/admin` 实现
3. 引入 `ant-design-pro` 基础工程结构与配置
4. 接通会话初始化、请求封装、组织上下文、登录跳转
5. 搭建 `ProLayout + dark sider + PageTabs` 壳层
6. 实现新信息架构与主路由
7. 逐个迁移业务页面与接口联调
8. 增加旧路由重定向
9. 收口样式、兼容配置、测试与构建验证

## 测试与验证策略

### 构建验证

- 新 `web/admin` 能完成依赖安装
- 新 `web/admin` 能完成生产构建
- 产物可用于静态部署

### 功能验证

至少覆盖以下核心流程：

- 会话获取成功时进入管理台
- 会话失效时跳转 `/auth/login`
- 任务列表筛选、翻页、删除、进入详情 / 结果页
- 新建任务成功提交
- 结果列表批量处置
- 规则分类与规则管理增删改查
- 文稿列表查询、详情查看、整改保存与提交复核
- 日志列表筛选与快照查看

### 标签页专项验证

- 打开多个页面时不整页刷新
- 同一路由不重复开页签
- 关闭当前 / 关闭其他 / 关闭全部可用
- 刷新浏览器后恢复最近标签状态
- 标签切换不丢失 URL 查询参数

### 浏览器验证

至少执行以下人工回归：

- Chrome 当前稳定版
- 一个较老版本 Chrome 或等价极速模式浏览器

## 风险与应对

### 风险 1：较老现代浏览器兼容性

风险：`antd` 新版本默认使用较新的 CSS 能力，低于兼容基线的浏览器可能表现异常。

应对：

- 在 Umi 中显式声明 `targets`
- 在 antd 中启用兼容型 `StyleProvider`
- 控制自定义 CSS 的特性使用范围
- 用真实目标浏览器做人工回归

### 风险 2：标签化工作台实现复杂度

风险：页签系统若与路由强耦合，可能导致重复状态源和维护复杂度上升。

应对：

- 标签栏只负责“页面打开记录与切换”
- URL 仍作为页面筛选状态的唯一主来源
- 草稿状态仅保留在少量必要页面中

### 风险 3：旧路由切换后的用户适配成本

风险：用户可能依赖旧收藏链接或内部文档中的旧路径。

应对：

- 为常用旧路径保留重定向
- 菜单与面包屑全部切换到新架构
- 通过设计文档和上线说明告知新入口位置

## 验收标准

本次改造完成时，需同时满足以下条件：

1. 旧 React 管理台已通过 Git tag `archive/admin-react-legacy-2026-05-09` 留档。
2. 仓库主干中不再保留旧 `web/admin` 副本目录。
3. 新 `web/admin` 采用标准 `ant-design-pro` / `Umi Max` 组织结构。
4. 左侧导航采用深色方案，主底色为 `#191a23`。
5. 整体视觉具备 `iview-admin` 式中国企业后台气质，但技术实现完全基于 React + ant-design-pro。
6. 所有现有核心功能在新管理台中保持可用。
7. 新信息架构与新路由已落地，旧路径仅保留必要重定向。
8. 标签化工作台可用，打开页面不整页刷新。
9. 浏览器支持基线明确为现代浏览器与较老 Chrome / 极速模式，不承诺 IE。
10. 构建通过、核心联调通过、关键页面验证通过。

## 参考依据

- `ant-design-pro` 官方工程结构与路由组织思路
- `Ant Design` 官方关于现代浏览器支持与兼容样式的说明
- `Umi` 官方关于 `targets` 与 antd `styleProvider` 的配置说明
