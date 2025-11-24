# CamRec：RTSP 视频分段录制与存储服务

docs目录里有api文档

推荐单独部署

CamRec 是一个使用 Go 开发的服务端应用，负责从 RTSP 拉流，按固定时长（默认 10 分钟）分段录制，通过 MinIO 存储视频片段，并将元数据记录到 PostgreSQL；同时提供查询视频 URL 的 HTTP API。

## 特性
- FFmpeg 分段录制（`segment`），文件命名包含时间戳（`%Y%m%d_%H%M%S.mp4`）
- MinIO 上传与公开/预签名 URL 返回（可配置）
- PostgreSQL 存储对象键与起止时间等元信息
- HTTP API：查询全部、按时间范围、最新 N 条、健康检查
- 保留策略：当天视频片段保留本地；非当天片段在成功上传入库后清理（可按需调整）

## 运行环境
- Go 1.21+
- FFmpeg（命令行可执行 `ffmpeg`）
- MinIO（对象存储）
- PostgreSQL（数据库）
- 操作系统：Linux/WSL/macOS/Windows（Docker 部署推荐）

## 配置文件（`config.yaml`）
```yaml
server:
  addr: ":8080"
  presign_expire_seconds: 3600
  public_bucket_policy: false
recorder:
  ffmpeg_path: "ffmpeg"
  rtsp_url: "rtsp://192.168.137.65:8554/live"
  output_dir: "data"
  segment_seconds: 600
minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "camrec"
  use_ssl: false
  region: "us-east-1"
postgres:
  dsn: "postgres://postgres:postgres@localhost:5432/camrec?sslmode=disable"
```
- 可通过环境变量 `CAMREC_CONFIG` 指定配置路径。
- 若 `public_bucket_policy=true`，API 直接返回公开 URL；否则返回预签名 URL（有效期由 `presign_expire_seconds` 控制）。

## 本地运行
1. 安装 FFmpeg 并确保 `ffmpeg` 在 PATH 或配置 `recorder.ffmpeg_path` 为绝对路径
2. 启动 MinIO 与 PostgreSQL（本地或容器）
3. 在项目根目录：
   ```bash
   go build -o camrec ./cmd/server
   ./camrec
   ```
4. 验证：
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/videos
   ```

## Docker 部署
已提供 `Dockerfile` 与 `docker-compose.yaml`。

### 一键部署
```bash
docker compose up -d --build
```

### 仅部署 camrec 服务
```bash
docker compose up -d --build camrec
```

### 说明
- MinIO 控制台：`http://localhost:9001`（默认账号密码：`minioadmin` / `minioadmin`）
- CamRec 容器配置：`config.docker.yaml` 挂载到 `/etc/camrec/config.yaml`
- 数据卷：本地当天片段保存在容器 `/var/lib/camrec`（映射卷 `camrec_data`）

## 快速 API
- 健康检查：`GET /health`
- 查询全部：`GET /videos`
- 按范围查询：`GET /videos/range?start=<RFC3339>&end=<RFC3339>`（`end` 可省略）
- 最新 N 条：`GET /videos/latest?n=<int>`（默认 10）
- 详见：`docs/API.md`

## 常见问题
- RTSP 不稳定：可启用 `-rtsp_transport tcp`（已默认启用）以提高稳定性。
- 文件名与时间：通过 `-strftime 1` 按时间命名，服务使用文件名解析起止时间。
- 访问 URL：设置 `public_bucket_policy=true` 返回公开 URL；否则返回预签名 URL（默认 3600 秒）。

## 维护与扩展
- 可增加分页与筛选参数（`page`、`page_size`、`day=YYYY-MM-DD`）
- 可扩展管理接口（删除、批量清理）与鉴权策略
- 可改为“区间重叠即命中”的查询语义：`NOT (end_time < start OR start_time > end)`

## 部署脚本（可选）
提供 `deploy.sh`（Linux/WSL/Git Bash）：
```bash
bash deploy.sh
```
- 选项：一键部署 / 仅部署 camrec


