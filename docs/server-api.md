# Account Service Server API

本文档描述当前 `service` 后端实际提供的 HTTP API，并给出本地环境 curl 访问示例。

## 基础信息

- 本地服务地址：`http://127.0.0.1:8000`
- API 前缀：`/api/v1`
- 默认管理员：`admin`
- 默认管理员密码：`strongpass`
- 管理端认证方式：JWT Bearer Token
- 请求体格式：`Content-Type: application/json`

除登录、刷新 token、健康检查外，管理接口都需要请求头：

```http
Authorization: Bearer <accessToken>
```

外部服务接口使用 API Key：

```http
Authorization: Bearer <api_key>
```

通用错误响应：

```json
{
  "error": {
    "code": "unauthorized",
    "message": "Access token is required"
  }
}
```

## 快速开始

登录并保存 token：

```bash
curl -sS --location --request POST 'http://127.0.0.1:8000/api/v1/admin/login' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "username": "admin",
    "password": "strongpass"
  }'
```

响应示例：

```json
{
  "user": {
    "id": "admin",
    "username": "admin"
  },
  "accessToken": "<accessToken>",
  "refreshToken": "<refreshToken>"
}
```

后续示例中用环境变量表示 token：

```bash
ACCESS_TOKEN='<accessToken>'
REFRESH_TOKEN='<refreshToken>'
```

## 健康检查

### 存活检查

```http
GET /health/live
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/health/live'
```

成功响应：

```json
{
  "status": "ok"
}
```

### 就绪检查

```http
GET /health/ready
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/health/ready'
```

成功响应：

```json
{
  "status": "ok"
}
```

## 管理员认证

### 登录

```http
POST /api/v1/admin/login
```

请求体：

```json
{
  "username": "admin",
  "password": "strongpass"
}
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/admin/login' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "username": "admin",
    "password": "strongpass"
  }'
```

成功响应：

```json
{
  "user": {
    "id": "admin",
    "username": "admin"
  },
  "accessToken": "<accessToken>",
  "refreshToken": "<refreshToken>"
}
```

### 查询当前管理员

```http
GET /api/v1/admin/me
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/admin/me' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

成功响应：

```json
{
  "user": {
    "id": "admin",
    "username": "admin"
  }
}
```

### 刷新 Token

```http
POST /api/v1/admin/refresh
```

请求体：

```json
{
  "refreshToken": "<refreshToken>"
}
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/admin/refresh' \
  --header 'Content-Type: application/json' \
  --data-raw "{
    \"refreshToken\": \"${REFRESH_TOKEN}\"
  }"
```

成功响应：

```json
{
  "user": {
    "id": "admin",
    "username": "admin"
  },
  "accessToken": "<newAccessToken>",
  "refreshToken": "<newRefreshToken>"
}
```

### 登出

```http
POST /api/v1/admin/logout
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/admin/logout' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

成功响应：

```json
{
  "ok": true
}
```

## 账号管理

账号状态可选值：

- `active`
- `disabled`
- `exhausted`
- `login_failed`
- `token_expired`
- `region_blocked`
- `error`

账号类型可选值：

- `claude`
- `aws`
- `gpt`
- `kiro-aws`
- `kiro-offical`
- `claudecode`
- `codex`

账号对象字段：

```json
{
  "id": "account-id",
  "username": "user@example.com",
  "password": "plain-password",
  "login_url": "https://example.com/login",
  "access_token": "provider-access-token",
  "refresh_token": "provider-refresh-token",
  "region": "us",
  "account_type": "codex",
  "status": "active",
  "quota_total": 1000,
  "quota_used": 100,
  "quota_remaining": 900,
  "max_concurrent_leases": 1,
  "tags": ["openai"],
  "notes": "primary account",
  "CreatedAt": "2026-05-25T11:00:00Z",
  "UpdatedAt": "2026-05-25T11:00:00Z"
}
```

### 创建账号

```http
POST /api/v1/accounts
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/accounts' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "username": "user@example.com",
    "password": "plain-password",
    "login_url": "https://example.com/login",
    "access_token": "provider-access-token",
    "refresh_token": "provider-refresh-token",
    "region": "us",
    "account_type": "codex",
    "status": "active",
    "quota_total": 1000,
    "quota_used": 100,
    "quota_remaining": 900,
    "max_concurrent_leases": 1,
    "tags": ["openai"],
    "notes": "primary account"
  }'
```

成功响应：

```json
{
  "account": {
    "id": "account-id",
    "username": "user@example.com",
    "password": "plain-password",
    "login_url": "https://example.com/login",
    "access_token": "provider-access-token",
    "refresh_token": "provider-refresh-token",
    "region": "us",
    "account_type": "codex",
    "status": "active",
    "quota_total": 1000,
    "quota_used": 100,
    "quota_remaining": 900,
    "max_concurrent_leases": 1,
    "tags": ["openai"],
    "notes": "primary account",
    "CreatedAt": "2026-05-25T11:00:00Z",
    "UpdatedAt": "2026-05-25T11:00:00Z"
  }
}
```

### 查询账号列表

```http
POST /api/v1/accounts/query
```

请求体字段：

- `region`：按区域过滤，可为空
- `account_type`：按账号类型过滤，可为空；可选值为 `claude`、`aws`、`gpt`、`kiro-aws`、`kiro-offical`、`claudecode`、`codex`
- `statuses`：按状态列表过滤，可为空
- `tags`：要求账号包含所有标签，可为空
- `min_quota_remaining`：最小剩余额度
- `limit`：最大返回数量，默认上限为 100

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/accounts/query' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "region": "us",
    "account_type": "codex",
    "statuses": ["active"],
    "tags": ["openai"],
    "min_quota_remaining": 1,
    "limit": 10
  }'
```

成功响应：

```json
{
  "accounts": [
    {
      "id": "account-id",
      "username": "user@example.com",
      "region": "us",
      "account_type": "codex",
      "status": "active",
      "quota_remaining": 900,
      "tags": ["openai"]
    }
  ]
}
```

### 查询账号详情

```http
GET /api/v1/accounts/:id
```

curl 示例：

```bash
ACCOUNT_ID='account-id'

curl -i "http://127.0.0.1:8000/api/v1/accounts/${ACCOUNT_ID}" \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

成功响应：

```json
{
  "account": {
    "id": "account-id",
    "username": "user@example.com",
    "region": "us",
    "account_type": "codex",
    "status": "active",
    "CreatedAt": "2026-05-25T11:00:00Z",
    "UpdatedAt": "2026-05-25T11:00:00Z"
  }
}
```

### 更新账号

```http
PATCH /api/v1/accounts/:id
```

请求体只需要传要修改的字段。

curl 示例：

```bash
curl -i --location --request PATCH "http://127.0.0.1:8000/api/v1/accounts/${ACCOUNT_ID}" \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "quota_remaining": 800,
    "notes": "quota updated"
  }'
```

成功响应：

```json
{
  "account": {
    "id": "account-id",
    "quota_remaining": 800,
    "notes": "quota updated"
  }
}
```

### 更新账号状态

```http
POST /api/v1/accounts/:id/status
```

curl 示例：

```bash
curl -i --location --request POST "http://127.0.0.1:8000/api/v1/accounts/${ACCOUNT_ID}/status" \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "status": "token_expired",
    "reason": "refresh failed"
  }'
```

成功响应：

```json
{
  "account": {
    "id": "account-id",
    "status": "token_expired"
  }
}
```

### 删除账号

```http
DELETE /api/v1/accounts/:id
```

curl 示例：

```bash
curl -i --location --request DELETE "http://127.0.0.1:8000/api/v1/accounts/${ACCOUNT_ID}" \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

成功响应：

```json
{
  "ok": true
}
```

## 租约管理

租约状态可选值：

- `active`
- `released`
- `expired`

### 申请账号租约

```http
POST /api/v1/accounts/acquire
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/accounts/acquire' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "region": "us",
    "account_type": "codex",
    "tags": ["openai"],
    "min_quota_remaining": 1,
    "ttl_seconds": 900,
    "purpose": "local-test",
    "caller_id": "caller-id"
  }'
```

成功响应：

```json
{
  "lease_id": "lease-id",
  "account_id": "account-id",
  "caller_id": "caller-id",
  "purpose": "local-test",
  "status": "active",
  "leased_at": "2026-05-25T11:00:00Z",
  "expires_at": "2026-05-25T11:15:00Z",
  "request_filters": {
    "region": "us",
    "account_type": "codex",
    "ttl_seconds": 900,
    "caller_id": "caller-id"
  },
  "account": {
    "id": "account-id",
    "username": "user@example.com"
  }
}
```

### 释放账号租约

```http
POST /api/v1/accounts/release
```

curl 示例：

```bash
LEASE_ID='lease-id'

curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/accounts/release' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw "{
    \"lease_id\": \"${LEASE_ID}\"
  }"
```

成功响应：

```json
{
  "ok": true
}
```

### 查询租约列表

```http
GET /api/v1/leases?status=<status>
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/leases?status=released' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

成功响应：

```json
{
  "leases": [
    {
      "lease_id": "lease-id",
      "account_id": "account-id",
      "caller_id": "caller-id",
      "purpose": "local-test",
      "status": "released"
    }
  ]
}
```

## API Key 管理

API Key 用于维护外部服务调用方凭证。API Key 管理是管理端接口，需要管理员 JWT；返回的明文 `api_key` 只在创建响应中出现一次。`status` 为 `active` 时可用于外部接口鉴权，`disabled` 时不可用。

### 查询 API Key 列表

```http
GET /api/v1/api-keys
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/api-keys' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

### 创建 API Key

```http
POST /api/v1/api-keys
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/api-keys' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "name": "worker",
    "description": "local worker",
    "status": "active"
  }'
```

成功响应：

```json
{
  "caller": {
    "id": "caller-id",
    "name": "worker",
    "status": "active",
    "description": "local worker",
    "created_at": "2026-05-25T11:00:00Z",
    "updated_at": "2026-05-25T11:00:00Z"
  },
  "api_key": "acct_xxx"
}
```

后续示例中用环境变量表示 API Key：

```bash
API_KEY='acct_xxx'
```

### 更新 API Key

```http
PATCH /api/v1/api-keys/{id}
```

curl 示例：

```bash
curl -i --location --request PATCH 'http://127.0.0.1:8000/api/v1/api-keys/caller-id' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "name": "worker",
    "description": "local worker",
    "status": "disabled"
  }'
```

### 删除 API Key

```http
DELETE /api/v1/api-keys/{id}
```

curl 示例：

```bash
curl -i --location --request DELETE 'http://127.0.0.1:8000/api/v1/api-keys/caller-id' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

## 模型配置管理

模型配置管理接口用于维护外部服务读取的模型列表、隐藏模型、模型别名和列表隐藏项。管理接口需要管理员 JWT。外部模型配置接口只返回 `active` 状态的配置项。

模型配置项字段：

- `kind`：`fallback_model`、`hidden_model`、`model_alias`、`hidden_from_list`
- `key`：模型 ID、隐藏模型名称、别名名称或隐藏列表项
- `value`：`hidden_model` 和 `model_alias` 的目标值；其他类型可为空
- `status`：`active` 可用、`disabled` 禁用
- `display_order`：同类型内排序值

### 查询模型配置项

```http
GET /api/v1/model-config/items
```

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/model-config/items' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

### 创建模型配置项

```http
POST /api/v1/model-config/items
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/model-config/items' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "kind": "model_alias",
    "key": "claude-opus-4-7",
    "value": "claude-opus-4.7",
    "status": "active",
    "display_order": 70
  }'
```

### 更新模型配置项

```http
PATCH /api/v1/model-config/items/{id}
```

curl 示例：

```bash
curl -i --location --request PATCH 'http://127.0.0.1:8000/api/v1/model-config/items/item-id' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}" \
  --data-raw '{
    "value": "claude-opus-4.8",
    "status": "disabled"
  }'
```

### 删除模型配置项

```http
DELETE /api/v1/model-config/items/{id}
```

curl 示例：

```bash
curl -i --location --request DELETE 'http://127.0.0.1:8000/api/v1/model-config/items/item-id' \
  --header "Authorization: Bearer ${ACCESS_TOKEN}"
```

## 外部服务账号接口

外部服务接口默认使用 API Key 鉴权，不接受管理员 JWT。鉴权可通过服务端环境变量 `EXTERNAL_API_KEY_AUTH_ENABLED` 控制：本地环境默认关闭，其他环境默认开启。开启时请求头格式：

```http
Authorization: Bearer <api_key>
```

### 外部查询模型配置

```http
GET /api/v1/external/model-config
```

统一返回可用兜底模型、隐藏模型、模型别名和列表隐藏项，便于外部服务动态读取模型配置。

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/external/model-config' \
  --header "Authorization: Bearer ${API_KEY}"
```

成功响应：

```json
{
  "fallback_models": [
    { "model_id": "auto" },
    { "model_id": "claude-sonnet-4" },
    { "model_id": "claude-haiku-4.5" },
    { "model_id": "claude-sonnet-4.5" },
    { "model_id": "claude-opus-4.5" },
    { "model_id": "claude-opus-4.6" },
    { "model_id": "claude-sonnet-4.6" }
  ],
  "hidden_models": {
    "claude-3.7-sonnet": "CLAUDE_3_7_SONNET_20250219_V1_0",
    "claude-opus-4.6": "claude-opus-4.6",
    "claude-sonnet-4.6": "claude-sonnet-4.6"
  },
  "model_aliases": {
    "auto-kiro": "auto",
    "claude-opus-4-6": "claude-opus-4.6",
    "claude-sonnet-4-6": "claude-sonnet-4.6",
    "claude-opus-4-5": "claude-opus-4.5",
    "claude-sonnet-4-5": "claude-sonnet-4.5",
    "claude-haiku-4-5": "claude-haiku-4.5",
    "claude-opus-4-7": "claude-opus-4.7"
  },
  "hidden_from_list": ["auto"]
}
```

### 外部查询账号

```http
POST /api/v1/external/accounts/query
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/external/accounts/query' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${API_KEY}" \
  --data-raw '{
    "region": "us",
    "account_type": "codex",
    "statuses": ["active"],
    "tags": ["openai"],
    "min_quota_remaining": 1,
    "limit": 10
  }'
```

### 外部查询账号列表

```http
GET /api/v1/external/accounts
```

支持查询参数：

- `region`
- `account_type`
- `status`，可用逗号分隔多个状态
- `statuses`，同 `status`
- `tags`，可用逗号分隔多个标签
- `min_quota_remaining`
- `limit`

curl 示例：

```bash
curl -i 'http://127.0.0.1:8000/api/v1/external/accounts?region=us&account_type=codex&status=active&tags=openai&min_quota_remaining=1&limit=10' \
  --header "Authorization: Bearer ${API_KEY}"
```

成功响应：

```json
{
  "accounts": [
    {
      "id": "account-id",
      "username": "worker@example.com",
      "password": "plaintext-password",
      "login_url": "https://example.com/login",
      "access_token": "access-token",
      "refresh_token": "refresh-token",
      "region": "us",
      "account_type": "codex",
      "status": "active",
      "quota_total": 1000,
      "quota_used": 100,
      "quota_remaining": 900,
      "max_concurrent_leases": 1,
      "tags": ["openai"],
      "notes": "",
      "created_at": "2026-05-25T11:00:00Z",
      "updated_at": "2026-05-25T11:00:00Z"
    }
  ]
}
```

### 外部更新账号状态

```http
POST /api/v1/external/accounts/{id}/status
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/external/accounts/account-id/status' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${API_KEY}" \
  --data-raw '{
    "status": "disabled",
    "reason": "maintenance"
  }'
```

成功响应：

```json
{
  "account": {
    "id": "account-id",
    "username": "worker@example.com",
    "password": "plaintext-password",
    "login_url": "https://example.com/login",
    "access_token": "access-token",
    "refresh_token": "refresh-token",
    "region": "us",
    "account_type": "codex",
    "status": "disabled",
    "quota_total": 1000,
    "quota_used": 100,
    "quota_remaining": 900,
    "max_concurrent_leases": 1,
    "tags": ["openai"],
    "notes": "",
    "created_at": "2026-05-25T11:00:00Z",
    "updated_at": "2026-05-25T11:05:00Z"
  }
}
```

### 外部申请账号租约

```http
POST /api/v1/external/accounts/acquire
```

外部接口会使用 API Key 对应的调用方 ID 作为 `caller_id`，请求体里的 `caller_id` 会被忽略。

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/external/accounts/acquire' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${API_KEY}" \
  --data-raw '{
    "region": "us",
    "account_type": "codex",
    "tags": ["openai"],
    "min_quota_remaining": 1,
    "ttl_seconds": 900,
    "purpose": "worker-task"
  }'
```

### 外部释放账号租约

```http
POST /api/v1/external/accounts/release
```

curl 示例：

```bash
curl -i --location --request POST 'http://127.0.0.1:8000/api/v1/external/accounts/release' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer ${API_KEY}" \
  --data-raw "{
    \"lease_id\": \"${LEASE_ID}\"
  }"
```

## 当前未实现接口

前端页面中已经预留了审计日志调用：

```http
GET /api/v1/audit-logs
```

但当前后端服务尚未注册该路由，访问会返回 `404 Not Found`。后续实现审计日志接口时，需要再补充本文档。
