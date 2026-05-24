# Web Admin Redesign Spec

日期：2026-05-24

## 目标

将当前前端管理后台从多个页面直接堆叠的原型界面，重设计为基于 shadcn/ui 组件体系的运维控制台。

## 设计方向

采用“运维控制台”布局：

- 登录界面单独设计，不显示后台导航和业务页面。
- 登录成功后进入管理后台 Shell。
- 后台 Shell 使用左侧导航，主区域展示当前模块。
- 顶部展示页面标题、关键状态和主要操作。
- 账号、租借、API Key、审计日志都在统一后台框架内切换。

## 登录界面

登录页是独立入口：

- 居中登录面板。
- 品牌标题为 `Account Admin`。
- 输入项使用 shadcn `Input`、`Label`、`Button`。
- 错误使用 shadcn `Alert` 或清晰的 inline feedback。
- 登录页不展示账号列表、租借、API Key 或审计内容。

## 后台 Shell

登录后显示：

- 左侧 sidebar：Overview、Accounts、Leases、API Keys、Audit Logs。
- 主内容区：页面标题、简短说明、操作区、数据区。
- 顶部状态条：当前用户、登出按钮。
- 使用 shadcn 风格的 `Card`、`Table`、`Tabs`、`Dialog`、`Sheet`、`Badge`、`Alert`、`Button`、`Input`。

## Accounts 页面

账号页是主工作区：

- 顶部指标卡：Active accounts、Total quota、Active leases、Error states。
- 筛选区：region、account type、status、tags、minimum quota。
- 桌面端使用表格。
- 移动端可以退化为卡片列表。
- 新增/编辑账号使用 Dialog 或 Sheet 表单。
- 敏感字段默认隐藏，通过明确的 reveal 操作查看。

## Leases 页面

- 使用筛选条和表格展示 lease。
- 状态用 Badge 表达：active、released、expired。
- 支持按状态过滤。

## API Keys 页面

- 创建 API Key 使用 Dialog 表单。
- 明文 API Key 只在创建后的一次性区域显示。
- dismiss 后从界面移除。

## Audit Logs 页面

- 使用表格展示 actor、action、request id、metadata。
- metadata 中的敏感值必须显示为脱敏状态。
- request id 应可快速复制或扫描。

## 视觉原则

- 风格克制、工作台化、信息密度适中。
- 不使用营销式 hero。
- 不使用大面积单色渐变、装饰光斑或无意义插画。
- 使用语义色和状态 Badge，不只依赖颜色传达状态。
- 控件触达区域不小于 44px。
- 保留键盘焦点状态。

## 实现边界

- 本轮只重设计 `web` 前端。
- 后端 API 不改。
- shadcn/ui 组件可以本地实现为项目内组件，也可以通过 shadcn CLI 生成。
- 保留现有测试覆盖，并新增/调整测试以覆盖登录独立页和后台 Shell。

## 验收

- 未登录时只看到登录界面。
- 登录成功后进入后台 Shell。
- 账号、租借、API Key、审计页面通过左侧导航切换。
- 页面使用 shadcn 风格组件，不再是裸 HTML 堆叠。
- `npm test -- --run` 通过。
- `npm run build` 通过。
