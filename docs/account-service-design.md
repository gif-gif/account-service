# 账号管理系统设计文档

日期：2026-05-24

## 1. 目标

设计一个账号管理系统，供其他内部服务通过 HTTP API 获取可用账号及完整凭据。

系统需要管理用户名、密码、登录地址、accessToken、refreshToken、地区、账号额度、账号类型、账号状态等信息，并同时支持“查询账号”和“租借账号”两种调用方式。

## 2. 第一版范围

第一版包含：

- HTTP REST API：账号查询、账号租借、账号归还、账号创建、账号更新、状态更新、API Key 鉴权。
- 简单管理后台：账号增删改查、额度维护、状态切换、Token 更新、租借记录查看、API Key 管理、审计日志查看。
- PostgreSQL 数据库。
- TTL 自动过期租借。
- 静态额度字段，由管理员或外部服务更新。
- 账号选择策略：优先返回剩余额度最高的账号，同额度随机。

第一版不包含：

- 自动扣减额度。
- 从第三方平台同步真实额度。
- 自动刷新 Token。
- 复杂告警系统。
- 独立 Worker 服务。
- 独立前端部署。
- gRPC 或 SDK。

## 3. 总体架构

采用单体服务架构：同一个后端应用同时提供 REST API、管理后台和租借 TTL 清理任务。

核心模块：

- 账号模块：管理账号凭据、地区、类型、额度、状态、标签和备注。
- 租借模块：处理账号租借、归还、并发上限、TTL 过期。
- 调用方模块：管理内部服务和 API Key。
- 后台模块：提供账号和系统管理页面。
- 审计模块：记录账号访问、修改、租借、归还等敏感操作。
- 安全模块：处理 API Key 鉴权、敏感字段加密、日志脱敏。

推荐部署：

- 1 个应用服务。
- 1 个 PostgreSQL 数据库。
- 通过网关或反向代理提供 HTTPS。
- 后续如果租借吞吐很高，可引入 Redis 做缓存或计数优化。

## 4. 安全策略

本设计按“授权调用方总是返回完整凭据”处理，包括密码、accessToken、refreshToken。这种方式实现简单，但风险较高。

最低安全要求：

- 所有 API 必须通过 HTTPS 调用。
- 每个调用方服务必须使用 API Key 鉴权。
- API Key 入库存储哈希值，不存明文。
- 密码、accessToken、refreshToken 在数据库中加密存储。
- 应用日志不得输出明文密码、accessToken、refreshToken。
- 审计日志记录访问者、动作、资源和时间，但必须脱敏敏感字段。
- 管理后台必须登录后访问。
- API 响应可以返回完整凭据，但错误日志、访问日志、链路追踪和监控事件不得记录明文凭据。

## 5. 账号状态

账号状态使用以下枚举：

- `active`：可用，可被查询或租借。
- `disabled`：人工禁用，不可用。
- `exhausted`：额度不可用或已耗尽。
- `login_failed`：登录失败。
- `token_expired`：Token 已过期。
- `region_blocked`：账号在指定地区受限。
- `error`：其他未知异常。

普通查询可按状态筛选；租借接口默认只选择 `active` 账号。

## 6. 数据模型

### accounts

账号主表。

字段：

- `id`：UUID 主键。
- `username`：用户名。
- `password_encrypted`：加密后的密码。
- `login_url`：登录地址。
- `access_token_encrypted`：加密后的 accessToken。
- `refresh_token_encrypted`：加密后的 refreshToken。
- `region`：地区，例如 `us`、`eu`、`cn`，也允许业务自定义。
- `account_type`：账号类型，例如 `free`、`pro`、`enterprise` 或业务自定义类型。
- `status`：账号状态。
- `quota_total`：总额度。
- `quota_used`：已用额度。
- `quota_remaining`：剩余额度。
- `quota_reset_at`：额度重置时间，可为空。
- `max_concurrent_leases`：最大并发租借数。
- `tags`：标签，用于筛选。
- `metadata`：JSON 扩展字段。
- `notes`：后台备注。
- `created_at`：创建时间。
- `updated_at`：更新时间。

建议索引：

- `(status, region, account_type)`。
- `(quota_remaining)`。
- `tags` 使用 PostgreSQL GIN 索引。

### account_leases

账号租借记录表。

字段：

- `id`：UUID 主键。
- `account_id`：账号 ID。
- `caller_id`：调用方服务 ID。
- `purpose`：调用方传入的用途说明。
- `request_filters`：租借时的筛选条件 JSON。
- `status`：`active`、`released`、`expired`。
- `leased_at`：租借开始时间。
- `expires_at`：租借过期时间。
- `released_at`：主动归还时间。
- `created_at`：创建时间。
- `updated_at`：更新时间。

建议索引：

- `(account_id, status)`。
- `(expires_at, status)`。
- `(caller_id, status)`。

### api_callers

调用方服务表。

字段：

- `id`：UUID 主键。
- `name`：服务名称。
- `api_key_hash`：API Key 哈希值。
- `status`：`active` 或 `disabled`。
- `description`：说明。
- `created_at`：创建时间。
- `updated_at`：更新时间。

### audit_logs

审计日志表。

字段：

- `id`：UUID 主键。
- `actor_type`：`api_caller` 或 `admin`。
- `actor_id`：调用方或管理员 ID。
- `action`：动作，例如 `account.query`、`account.acquire`、`account.update`。
- `resource_type`：资源类型。
- `resource_id`：资源 ID。
- `request_id`：请求追踪 ID。
- `ip_address`：来源 IP。
- `user_agent`：User-Agent。
- `metadata`：脱敏后的 JSON 元数据。
- `created_at`：事件时间。

## 7. 账号选择策略

`POST /api/v1/accounts/acquire` 按以下规则选择账号：

1. 根据请求条件筛选账号：地区、账号类型、标签、状态、最低剩余额度等。
2. 只保留 `active` 账号。
3. 默认只保留 `quota_remaining > 0` 的账号。
4. 排除当前活跃租借数已达到 `max_concurrent_leases` 的账号。
5. 按 `quota_remaining` 从高到低排序。
6. 如果最高剩余额度有多个账号，则随机选择一个。
7. 在同一个数据库事务中创建租借记录。
8. 返回账号完整凭据和 `lease_id`。

并发控制必须使用数据库事务和行级锁，确保多个服务同时 acquire 时不会超过账号的并发租借上限。

## 8. 租借过期

租借使用 TTL 自动过期。

规则：

- `acquire` 接口支持传入 `ttl_seconds`。
- 不传 `ttl_seconds` 时使用系统默认 TTL。
- 后台定时任务定期把过期的 `active` 租借标记为 `expired`。
- `expired` 租借不再占用账号并发额度。
- 调用方可以在 TTL 到期前手动 release。

建议默认值：

- 默认 TTL：15 分钟。
- 最大 TTL：2 小时。
- 清理间隔：1 分钟。

## 9. REST API

除管理后台登录等特殊接口外，所有 API 都需要：

`Authorization: Bearer <api_key>`

### 查询账号

`POST /api/v1/accounts/query`

请求示例：

```json
{
  "region": "us",
  "account_type": "pro",
  "statuses": ["active"],
  "tags": ["openai"],
  "min_quota_remaining": 1,
  "limit": 10
}
```

返回账号完整凭据列表。

### 租借账号

`POST /api/v1/accounts/acquire`

请求示例：

```json
{
  "region": "us",
  "account_type": "pro",
  "tags": ["openai"],
  "min_quota_remaining": 1,
  "ttl_seconds": 900,
  "purpose": "chat-completion-worker"
}
```

响应示例：

```json
{
  "lease_id": "lease-id",
  "expires_at": "2026-05-24T10:30:00Z",
  "account": {
    "id": "account-id",
    "username": "user@example.com",
    "password": "plaintext-password",
    "login_url": "https://example.com/login",
    "access_token": "access-token",
    "refresh_token": "refresh-token",
    "region": "us",
    "account_type": "pro",
    "status": "active",
    "quota_total": 1000,
    "quota_used": 100,
    "quota_remaining": 900,
    "quota_reset_at": "2026-06-01T00:00:00Z",
    "max_concurrent_leases": 3,
    "tags": ["openai"]
  }
}
```

### 归还租借

`POST /api/v1/accounts/release`

请求示例：

```json
{
  "lease_id": "lease-id"
}
```

### 获取账号详情

`GET /api/v1/accounts/{id}`

返回单个账号详情，包括完整凭据。

### 创建账号

`POST /api/v1/accounts`

创建账号。

### 更新账号

`PATCH /api/v1/accounts/{id}`

更新账号凭据、额度、状态、标签、备注或扩展字段。

### 修改账号状态

`POST /api/v1/accounts/{id}/status`

请求示例：

```json
{
  "status": "token_expired",
  "reason": "refresh failed"
}
```

### 查看租借记录

`GET /api/v1/leases`

支持按账号 ID、调用方 ID、状态、时间范围筛选。

### 创建 API Key

`POST /api/v1/api-keys`

创建调用方 API Key。明文 API Key 只在创建时返回一次。

## 10. 管理后台

第一版后台页面：

- 账号列表：支持按地区、账号类型、状态、标签、额度筛选。
- 账号详情：展示账号信息、凭据、额度、租借记录、备注、审计事件。
- 新增和编辑账号。
- 状态切换。
- 额度维护。
- Token 和密码更新。
- 活跃租借和历史租借列表。
- API Key 管理。
- 审计日志查看。

后台列表页默认不展示密码和 Token。账号详情页可以提供显式 reveal 操作查看敏感字段。

## 11. 错误处理

统一错误格式：

```json
{
  "error": {
    "code": "no_available_account",
    "message": "No account matched the requested filters",
    "request_id": "request-id"
  }
}
```

常见错误：

- `401 unauthorized`：缺少或无效 API Key。
- `403 forbidden`：已认证但无权限。
- `404 account_not_found`：账号不存在。
- `404 no_available_account`：没有满足条件的可用账号。
- `409 lease_conflict`：租借已释放或已过期。
- `422 invalid_status`：非法账号状态。
- `429 too_many_requests`：可选的调用方限流。
- `500 internal_error`：系统内部错误。

## 12. 审计事件

至少记录以下事件：

- `account.query`
- `account.get`
- `account.acquire`
- `account.release`
- `account.create`
- `account.update`
- `account.status_update`
- `api_key.create`
- `api_key.disable`
- `admin.login`

审计元数据必须脱敏：

- `password`
- `access_token`
- `refresh_token`
- API Key 明文

## 13. 测试策略

核心测试：

- 账号查询筛选逻辑。
- acquire 优先选择剩余额度最高的账号。
- 同额度账号随机选择。
- 达到最大并发租借数后不再被 acquire。
- TTL 清理任务能过期旧租借。
- 手动 release 能释放并发额度。
- API Key 鉴权接受有效 Key，拒绝无效 Key。
- 敏感字段入库加密。
- 日志和审计元数据脱敏。
- 管理后台账号 CRUD 流程。

并发测试：

- 多个 acquire 并发请求不能超过 `max_concurrent_leases`。
- 过期租借不再占用并发额度。

## 14. 运行配置

建议配置项：

- `DATABASE_URL`
- `APP_BASE_URL`
- `SECRET_ENCRYPTION_KEY`
- `DEFAULT_LEASE_TTL_SECONDS`
- `MAX_LEASE_TTL_SECONDS`
- `LEASE_CLEANUP_INTERVAL_SECONDS`
- `ADMIN_SESSION_SECRET`

建议生产能力：

- 数据库备份。
- 慢查询日志。
- 请求 ID 透传。
- 健康检查接口。
- 基础指标：acquire 成功数、acquire 失败数、活跃租借数、过期租借数、各状态账号数量。

## 15. 后续扩展

后续可以扩展：

- 调用方角色和权限。
- 每个 API Key 配置 IP 白名单。
- 自动刷新 Token。
- 从第三方平台同步额度。
- 额度扣减或使用量上报接口。
- 账号异常告警。
- 常用语言 SDK。
- 独立 Worker 服务。
- Redis 租借计数优化。
