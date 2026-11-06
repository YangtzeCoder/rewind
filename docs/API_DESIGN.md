# API 设计规范

本规范描述 Rewind 接口的通用约定。具体接口及功能范围见 [PRD](PRD.md)。

## 请求

- REST API 统一使用 `/api/v1` 前缀。
- 需要认证的接口通过 `Authorization: Bearer <token>` 传递 Token；各接口允许的 Token 类型见 PRD。
- 文件上传使用 `multipart/form-data`，备份历史查询和下载使用 JSON 请求体。

## JSON 响应

所有 JSON 响应均使用 HTTP 200。客户端通过 `success` 判断操作是否成功，而不是仅根据 HTTP 状态码判断。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | 布尔值 | 必有；表示操作是否成功。 |
| `data` | 任意 JSON 值 | 成功且有业务数据时返回；没有返回内容时省略。 |
| `error` | 对象 | 失败时返回，包含字符串字段 `code` 和 `message`；成功时省略。 |

有返回内容的成功响应：

```json
{
  "success": true,
  "data": {"username": "admin"}
}
```

没有返回内容的成功响应：

```json
{"success": true}
```

返回数组的接口始终包含数组类型的 `data`。没有结果时返回空数组，不返回 `null` 或省略 `data`：

```json
{"success": true, "data": []}
```

失败响应省略 `data`，通过 `error.code` 标识错误类型，`error.message` 描述原因：

```json
{
  "success": false,
  "error": {
    "code": "<错误码>",
    "message": "<错误说明>"
  }
}
```

`error.code` 和 `error.message` 的默认值均为空字符串；发生错误时填写。错误说明使用英文；客户端应通过稳定的错误码判断错误类型。错误码以[代码中的定义](../backend/internal/server/response.go)为准。内部错误只返回通用说明，不暴露实现细节。

## 文件下载

文件下载成功时直接返回原始二进制内容，不使用 JSON 包装；下载失败时返回统一的 JSON 错误响应。响应内容已经开始发送后，无法再替换为 JSON 错误响应。
