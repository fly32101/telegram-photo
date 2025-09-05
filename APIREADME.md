# Telegram 图床 API 文档

## 基础信息

- 基础URL: `http://localhost:8080`
- 所有需要认证的API都需要在请求头中添加 `Authorization: Bearer {token}`
- 响应格式: JSON

## 认证相关

### 重定向到 GitHub 授权页面

```
GET /api/v1/auth/github
```

**响应示例:**

```json
{
  "url": "https://github.com/login/oauth/authorize?client_id=xxx&redirect_uri=xxx&scope=user"
}
```

### GitHub 授权回调 【这个打开弹窗登录，登陆完刷新当前页面】

```
GET /api/v1/auth/github/callback?code={code}
```

**参数:**

- `code`: GitHub 授权码

**响应示例:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "github_user_id",
  "redirect_url": "http://localhost:3000/auth/callback?token=xxx&user_id=xxx"
}
```

## 图片相关 (需要认证)

### 上传图片

```
POST /api/v1/image/upload
```

**请求头:**

```
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

**请求参数:**

- `image`: 图片文件 (form-data)

**响应示例:**

```json
{
  "message": "上传成功",
  "file_id": "telegram_file_id",
  "proxy_url": "http://localhost:8080/proxy/image/telegram_file_id"
}
```

### 获取用户图片列表

```
GET /api/v1/image/list?page={page}&page_size={page_size}
```

**请求头:**

```
Authorization: Bearer {token}
```

**查询参数:**

- `page`: 页码，默认为1
- `page_size`: 每页数量，默认为10

**响应示例:**

```json
{
  "images": [
    {
      "id": 1,
      "file_id": "telegram_file_id_1",
      "created_at": "2023-07-01T12:00:00Z",
      "proxy_url": "http://localhost:8080/proxy/image/telegram_file_id_1"
    },
    {
      "id": 2,
      "file_id": "telegram_file_id_2",
      "created_at": "2023-07-02T12:00:00Z",
      "proxy_url": "http://localhost:8080/proxy/image/telegram_file_id_2"
    }
  ],
  "total": 25,
  "page": 1
}
```

### 删除图片

```
DELETE /api/v1/image/{id}
```

**请求头:**

```
Authorization: Bearer {token}
```

**路径参数:**

- `id`: 图片ID

**响应示例:**

```json
{
  "message": "删除成功"
}
```

## 管理员接口 (需要管理员权限)

### 获取所有图片

```
GET /api/v1/admin/images?page={page}&page_size={page_size}&user_id={user_id}&upload_ip={upload_ip}
```

**请求头:**

```
Authorization: Bearer {token}
```

**查询参数:**

- `page`: 页码，默认为1
- `page_size`: 每页数量，默认为20
- `user_id`: (可选) 按用户ID筛选
- `upload_ip`: (可选) 按上传IP筛选

**响应示例:**

```json
{
  "images": [
    {
      "id": 1,
      "file_id": "telegram_file_id_1",
      "user_id": "github_user_id_1",
      "upload_ip": "127.0.0.1",
      "created_at": "2023-07-01T12:00:00Z",
      "updated_at": "2023-07-01T12:00:00Z",
      "proxy_url": "http://localhost:8080/proxy/image/telegram_file_id_1"
    },
    {
      "id": 2,
      "file_id": "telegram_file_id_2",
      "user_id": "github_user_id_2",
      "upload_ip": "192.168.1.1",
      "created_at": "2023-07-02T12:00:00Z",
      "updated_at": "2023-07-02T12:00:00Z",
      "proxy_url": "http://localhost:8080/proxy/image/telegram_file_id_2"
    }
  ],
  "total": 100,
  "page": 1
}
```

### 获取统计信息

```
GET /api/v1/admin/stats
```

**请求头:**

```
Authorization: Bearer {token}
```

**响应示例:**

```json
{
  "total_images": 100,
  "today_images": 15,
  "user_count": 25,
  "user_rankings": [
    {
      "UserID": "github_user_id_1",
      "Count": 30
    },
    {
      "UserID": "github_user_id_2",
      "Count": 25
    },
    {
      "UserID": "github_user_id_3",
      "Count": 20
    }
  ]
}
```

## 代理访问

### 代理访问图片

```
GET /proxy/image/{file_id}
```

**路径参数:**

- `file_id`: Telegram 文件ID

**响应:**

图片内容，Content-Type 根据图片类型设置

## 错误响应

所有API在发生错误时都会返回相应的HTTP状态码和错误信息：

```json
{
  "error": "错误信息"
}
```

常见HTTP状态码：

- 400 Bad Request: 请求参数错误
- 401 Unauthorized: 未认证或认证失败
- 403 Forbidden: 无权限访问
- 404 Not Found: 资源不存在
- 500 Internal Server Error: 服务器内部错误