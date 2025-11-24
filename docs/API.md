# CamRec API 文档

## 概览
- 作用：查询已录制的视频片段列表，返回 MinIO 可访问的 URL（公开或预签名）
- 基础地址：`http://<server-host>:<port>`（默认 `:8080`，见 `server.addr`）
- 返回格式：`application/json`
- 时间格式：RFC3339（例如 `2025-11-24T10:00:00Z` 或 `2025-11-24T18:00:00+08:00`）

## 认证与访问
- 当 `public_bucket_policy=true` 时，接口返回公开的 MinIO URL。
- 当 `public_bucket_policy=false` 时，接口返回预签名 URL；有效期由 `presign_expire_seconds` 控制（默认 3600 秒）。

## 错误约定
- `400` 参数错误（缺少必填参数或格式错误）
- `500` 服务内部错误（存储/数据库访问异常）
- `200` 正常返回

## 数据结构
```json
{
  "url": "string",
  "key": "string",
  "start_time": "RFC3339",
  "end_time": "RFC3339",
  "size_bytes": 123456
}
```

## 健康检查
- `GET /health`
- 说明：服务健康状态检查
- 响应示例：
```json
{"ok": true}
```
- 示例：
```bash
curl http://localhost:8080/health
```

## 查询全部视频
- `GET /videos`
- 说明：返回所有已记录的视频片段，按 `start_time` 升序。
- 请求参数：无
- 响应示例：
```json
[
  {
    "url": "https://minio.example.com/camrec/videos/2025/11/24/20251124_100000.mp4",
    "key": "videos/2025/11/24/20251124_100000.mp4",
    "start_time": "2025-11-24T10:00:00Z",
    "end_time": "2025-11-24T10:10:00Z",
    "size_bytes": 1234567
  }
]
```
- 示例：
```bash
curl http://localhost:8080/videos
```

## 按时间范围查询
- `GET /videos/range?start=<RFC3339>&end=<RFC3339>`
- 说明：按范围查询视频列表；`end` 可省略（为空或不传）则等于当前时间。
- 必填参数：
  - `start`：开始时间（RFC3339），例如 `2025-11-24T10:00:00Z`
- 可选参数：
  - `end`：结束时间（RFC3339），例如 `2025-11-24T12:00:00Z`
- 查询语义：返回完全落在 `[start, end]` 区间的片段（`start_time >= start AND end_time <= end`）。
- 响应示例：
```json
[
  {
    "url": "https://minio.example.com/camrec/videos/2025/11/24/20251124_110000.mp4",
    "key": "videos/2025/11/24/20251124_110000.mp4",
    "start_time": "2025-11-24T11:00:00Z",
    "end_time": "2025-11-24T11:10:00Z",
    "size_bytes": 2345678
  }
]
```
- 示例：
```bash
curl "http://localhost:8080/videos/range?start=2025-11-24T10:00:00Z&end=2025-11-24T12:00:00Z"
curl "http://localhost:8080/videos/range?start=2025-11-24T10:00:00Z"
```

## 查询最新 N 条
- `GET /videos/latest?n=<int>`
- 说明：返回最新的 N 条片段，按 `start_time` 降序；`n` 默认 10。
- 响应示例：
```json
[
  {
    "url": "https://minio.example.com/camrec/videos/2025/11/24/20251124_115000.mp4",
    "key": "videos/2025/11/24/20251124_115000.mp4",
    "start_time": "2025-11-24T11:50:00Z",
    "end_time": "2025-11-24T12:00:00Z",
    "size_bytes": 3456789
  }
]
```
- 示例：
```bash
curl "http://localhost:8080/videos/latest?n=5"
```

## URL 返回规则
- 公开 bucket：`url` 为 `http[s]://<endpoint>/<bucket>/<key>`。
- 非公开 bucket：`url` 为预签名临时链接，有效期 `presign_expire_seconds`（默认 3600 秒）。

## 说明与建议
- 若需要“重叠即命中”的查询（任意交集），可将区间逻辑调整为：`NOT (end_time < start OR start_time > end)`。
- 数据量大时建议扩展分页：`page`、`page_size`；也可增加 `day=YYYY-MM-DD` 等筛选参数。
- 管理功能（删除、批量清理、本地保留策略等）可按需扩展并配合鉴权。
