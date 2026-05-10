# 智约校园（Smart Campus）

一个面向校园场景的综合预约平台，当前包含：

- 用户注册 / 登录（JWT 鉴权）
- 空间查询与预约
- 热门活动秒杀抢票
- 订单查询、支付、取消

项目采用前后端分离：

- 后端：Go + Gin + GORM
- 前端：React + TypeScript + Vite + Zustand
- 基础设施：MySQL + Redis + RabbitMQ + Docker Compose

---

## 1. 功能概览

### 用户侧当前可用功能

- 注册新账号并自动登录
- 登录后查看个人信息、信用分、配额
- 查看空间列表与可预约时间槽
- 提交空间预约
- 查看活动列表并参与秒杀
- 在订单页完成支付或取消订单

### 当前版本的已知现状

- 后端启动时会自动执行 `AutoMigrate` 创建表
- **目前仓库内没有内置 seed 脚本，也没有后台管理页**
- 如果数据库里没有 `resources`、`time_slots`、`activities` 数据，前端页面会正常打开，但列表可能为空
- Docker Compose 已为所有服务配置 `restart: unless-stopped`

---

## 2. 目录结构

```text
uni-tech/
├── backend/               # Go 后端
│   ├── cmd/server         # 程序入口
│   └── internal/          # handler / model / pkg
├── frontend/              # React 前端
├── docker-compose.yml     # 本地一键启动
├── DEV_DOC.md             # 开发文档
└── ARCHITECTURE.md        # 架构文档
```

---

## 3. 快速开始

### 3.1 环境要求

本地推荐安装：

- Docker
- Docker Compose（新版本通常已内置在 Docker 中）

### 3.2 一键启动

在仓库根目录执行：

```bash
docker compose up -d --build
```

查看服务状态：

```bash
docker compose ps
```

查看后端日志：

```bash
docker compose logs -f backend
```

### 3.3 访问地址

启动成功后可访问：

- 前端页面：http://localhost
- 后端健康检查：http://localhost:8080/health
- RabbitMQ 管理台：http://localhost:15672
  - 用户名：`guest`
  - 密码：`guest`

默认基础设施端口：

- MySQL：`3306`
- Redis：`6379`
- RabbitMQ：`5672`

### 3.4 停止服务

停止但保留数据卷：

```bash
docker compose down
```

停止并删除数据卷（会清空 MySQL / Redis / RabbitMQ 持久化数据）：

```bash
docker compose down -v
```

---

## 4. 首次使用流程

### 4.1 打开前端

浏览器访问：

```text
http://localhost
```

由于首页受登录保护，首次使用建议直接进入注册页：

```text
http://localhost/register
```

### 4.2 注册账号

前端注册页当前需要填写：

- 学号
- 姓名
- 邮箱
- 密码
- 确认密码

注册成功后会自动登录并进入系统首页。

### 4.3 进入各功能页

登录后可通过顶部导航进入：

- 空间预订
- 活动秒杀
- 我的订单

---

## 5. 准备演示数据

如果你刚启动项目，数据库通常只有表结构，没有业务数据。

想在页面中看到“空间”和“活动”，需要先插入一些演示数据。

### 5.1 进入 MySQL

```bash
docker compose exec db mysql -uroot -proot123 smart_campus
```

### 5.2 插入一组空间与时间槽演示数据

```sql
SET @slot_date = DATE_ADD(CURDATE(), INTERVAL 1 DAY);

INSERT INTO resources (
  type, name, code, capacity, location, floor, building, description, status
) VALUES (
  'academic',
  'A101 自习室',
  'A101-DEMO',
  30,
  '教学楼 A101',
  1,
  '教学楼A',
  'README 演示用空间',
  'active'
);

SET @resource_id = LAST_INSERT_ID();

INSERT INTO academic_spaces (
  resource_id, buffer_minutes, min_duration, max_duration, allow_recurring, equipment
) VALUES (
  @resource_id,
  5,
  30,
  120,
  0,
  JSON_OBJECT('projector', true, 'whiteboard', true)
);

INSERT INTO time_slots (
  resource_id, slot_date, start_time, end_time, buffer_start, buffer_end, status, version
) VALUES
  (@resource_id, @slot_date, '09:00:00', '10:00:00', '08:55:00', '10:05:00', 'available', 0),
  (@resource_id, @slot_date, '10:30:00', '11:30:00', '10:25:00', '11:35:00', 'available', 0),
  (@resource_id, @slot_date, '14:00:00', '15:00:00', '13:55:00', '15:05:00', 'available', 0);
```

### 5.3 插入一组活动秒杀演示数据

```sql
INSERT INTO activities (
  title,
  description,
  organizer,
  speaker,
  location,
  activity_type,
  start_time,
  end_time,
  total_tickets,
  remaining_tickets,
  price,
  seckill_start,
  seckill_end,
  status,
  cover_image,
  view_count
) VALUES (
  '校园讲座演示票',
  '用于本地演示秒杀流程',
  '学生处',
  '张老师',
  '图书馆报告厅',
  'lecture',
  DATE_ADD(NOW(), INTERVAL 1 DAY),
  DATE_ADD(NOW(), INTERVAL 26 HOUR),
  50,
  50,
  9.90,
  DATE_SUB(NOW(), INTERVAL 10 MINUTE),
  DATE_ADD(NOW(), INTERVAL 1 DAY),
  'seckill',
  '',
  0
);
```

插入完成后刷新前端页面，即可看到示例数据。

---

## 6. 页面使用说明

### 6.1 空间预订

1. 打开“空间预订”页
2. 选择一个空间
3. 选择日期
4. 选择可用时间槽
5. 点击“确认预订”
6. 系统会创建一笔待支付订单

注意：

- 只有状态为 `available` 的时间槽可预约
- 创建预约后，订单会出现在“我的订单”页
- 当前实现中，空间预约订单金额为 `0`

### 6.2 活动秒杀

1. 打开“活动秒杀”页
2. 找到状态为“秒杀中”的活动
3. 点击“立即秒杀”
4. 抢票成功后系统会创建一笔待支付订单

注意：

- 秒杀依赖 Redis 原子扣减库存
- 同一用户不能重复抢同一活动
- 没库存时会返回失败提示

### 6.3 我的订单

订单页支持：

- 查看所有订单
- 对 `pending` 状态订单点击“立即支付”
- 对 `pending` 状态订单点击“取消订单”

当前订单类型包括：

- `space`：空间预约订单
- `activity`：活动门票订单

---

## 7. API 快速说明

后端基础地址：

```text
http://localhost:8080/api/v1
```

### 7.1 认证接口

#### 注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "student_id": "20260001",
    "name": "测试用户",
    "email": "test@example.com",
    "password": "123456"
  }'
```

#### 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "student_id": "20260001",
    "password": "123456"
  }'
```

登录成功后请保存返回的 `token`。

### 7.2 需要鉴权的接口

后续请求请带上：

```text
Authorization: Bearer <token>
```

示例：查询空间列表

```bash
curl http://localhost:8080/api/v1/spaces \
  -H 'Authorization: Bearer <token>'
```

示例：查询某天的时间槽

```bash
curl 'http://localhost:8080/api/v1/spaces/1/slots?date=2026-05-11' \
  -H 'Authorization: Bearer <token>'
```

示例：创建空间预约

```bash
curl -X POST http://localhost:8080/api/v1/spaces/bookings \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{
    "resource_id": 1,
    "slot_id": 1,
    "date": "2026-05-11"
  }'
```

示例：参与秒杀

```bash
curl -X POST http://localhost:8080/api/v1/activities/1/seckill \
  -H 'Authorization: Bearer <token>'
```

示例：查询订单

```bash
curl http://localhost:8080/api/v1/orders \
  -H 'Authorization: Bearer <token>'
```

---

## 8. 本地开发模式

如果你不想每次都完整重建前后端镜像，可以只把基础设施跑起来，然后分别启动前后端开发服务。

### 8.1 启动基础设施

```bash
docker compose up -d db redis mq
```

### 8.2 启动后端

```bash
cd backend
go run ./cmd/server
```

后端默认读取以下本地默认值：

- `DB_HOST=localhost`
- `DB_PORT=3306`
- `DB_USER=root`
- `DB_PASSWORD=root123`
- `DB_NAME=smart_campus`
- `REDIS_HOST=localhost`
- `REDIS_PORT=6379`
- `RABBITMQ_URL=amqp://guest:guest@localhost:5672/`
- `SERVER_PORT=8080`

运行测试：

```bash
go test ./...
```

### 8.3 启动前端

```bash
cd frontend
npm install
npm run dev
```

访问地址：

```text
http://localhost:3000
```

开发模式下，Vite 已将 `/api` 代理到 `http://localhost:8080`。

---

## 9. 常用命令

### 构建并启动全部服务

```bash
docker compose up -d --build
```

### 查看服务状态

```bash
docker compose ps
```

### 查看某个服务日志

```bash
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f db
```

### 重启某个服务

```bash
docker compose restart backend
```

### 进入后端容器

```bash
docker compose exec backend sh
```

### 进入数据库

```bash
docker compose exec db mysql -uroot -proot123 smart_campus
```

---

## 10. 重要说明

### 10.1 默认账号

项目**没有预置默认业务账号**，请先注册。

### 10.2 默认基础设施密码

当前 `docker-compose.yml` 里使用的是本地开发默认值：

- MySQL root 密码：`root123`
- RabbitMQ：`guest / guest`
- JWT Secret：开发默认值

这些配置只适合本地开发，不适合生产环境。

### 10.3 RabbitMQ 的使用状态

Compose 中已经包含 RabbitMQ，便于后续扩展异步订单处理、延迟任务等能力。
当前用户主流程主要依赖：

- 前端
- 后端
- MySQL
- Redis

---

## 11. 进一步阅读

- 详细开发文档：[`DEV_DOC.md`](./DEV_DOC.md)
- 详细架构文档：[`ARCHITECTURE.md`](./ARCHITECTURE.md)

如果你只是想先把项目跑起来，按这三个步骤即可：

1. `docker compose up -d --build`
2. 打开 `http://localhost/register` 注册账号
3. 向数据库插入一组演示空间 / 时间槽 / 活动数据后开始体验
