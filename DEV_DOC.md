# 智约校园 - 多态空间调度与高并发活动综合平台

## 开发文档

**版本**：v1.0
**更新日期**：2026-05-10

---

## 目录

1. [项目概述](#1-项目概述)
2. [系统架构](#2-系统架构)
3. [技术栈选型](#3-技术栈选型)
4. [数据库设计](#4-数据库设计)
5. [核心模块设计](#5-核心模块设计)
6. [API 接口设计](#6-api-接口设计)
7. [前端架构](#7-前端架构)
8. [容器化部署](#8-容器化部署)
9. [CI/CD 流水线](#9-cicd-流水线)
10. [开发规范](#10-开发规范)

---

## 1. 项目概述

### 1.1 项目背景

高校日常生活中，学生获取与预订校园资源（空教室、会议室、体育场馆、活动门票）面临入口分散、规则复杂、高峰期系统崩溃等问题。本项目旨在构建一个统一的"智约校园"平台，统一管理校园空间预订与活动票务。

### 1.2 核心功能模块

| 模块 | 描述 |
|------|------|
| 模块 A | 多态空间建模与预订系统（学术空间/体育设施） |
| 模块 B | 校园热门活动聚合（高并发秒杀） |
| 模块 C | 复杂状态机与动态规则引擎 |

### 1.3 性能指标要求

| 指标 | 要求 |
|------|------|
| LCP（最大内容绘制） | ≤ 2.5s |
| INP（交互到下一次绘制） | ≤ 200ms |
| CLS（累积布局偏移） | ≤ 0.1 |
| 秒杀并发 | 支持 10000+ QPS |

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         CDN (静态资源加速)                         │
├─────────────────────────────────────────────────────────────────┤
│                      API Gateway (网关层)                         │
│                   限流 / 鉴权 / 路由 / 日志                        │
├──────────┬──────────┬──────────┬──────────┬────────────────────┤
│  前端    │  用户    │  空间    │  秒杀    │   订单             │
│  Web     │  服务    │  预订    │  活动    │   状态机           │
│          │          │  服务    │  服务    │   服务             │
├──────────┴──────────┴──────────┴──────────┴────────────────────┤
│                     消息队列层 (RocketMQ)                        │
│              延迟队列 / 事务消息 / 事件驱动解耦                     │
├─────────────────────────────────────────────────────────────────┤
│     Redis Cluster      │         MySQL Cluster                    │
│  (缓存/分布式锁/会话)    │   (主从复制 + 读写分离)                   │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 服务职责划分

| 服务 | 职责 | 关键技术 |
|------|------|----------|
| API Gateway | 统一入口、限流、鉴权 | Kong / Nginx |
| User Service | 用户管理、认证授权 | JWT、RBAC |
| Space Service | 空间资源管理、预订 | 乐观锁、时间槽 |
| Seckill Service | 活动发布、票务秒杀 | Redis 缓存、MQ 削峰 |
| Order Service | 订单生命周期、状态机 | FSM、延迟队列 |
| Rule Engine | 动态规则评估 | 责任链模式 |

---

## 3. 技术栈选型

### 3.1 后端技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 语言 | Go 1.21+ | 高并发、性能优异 |
| 框架 | Gin + GORM | 轻量级 API 框架 |
| 数据库 | MySQL 8.0 | JSON 列、多版本并发控制 |
| 缓存 | Redis 7.2 | 分布式锁、缓存、Session |
| 消息队列 | RocketMQ 5.0 | 延迟消息、事务消息 |
| 容器 | Docker + K8s | 容器编排 |

### 3.2 前端技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 框架 | React 18 + TypeScript | 类型安全 |
| 构建 | Vite 5 | 快速开发、冷启动 |
| 状态 | Zustand | 轻量状态管理 |
| 样式 | Tailwind CSS | 原子化 CSS |
| UI 库 | Radix UI | 无障碍组件 |
| 测试 | Vitest + Playwright | 单元 + E2E 测试 |

### 3.3 DevOps 工具链

| 类别 | 工具 | 说明 |
|------|------|------|
| CI/CD | GitHub Actions | 自动化流水线 |
| 镜像仓库 | Docker Hub / GHCR | 镜像存储 |
| 监控 | Prometheus + Grafana | 可观测性 |
| 日志 | Loki + Promtail | 日志收集 |

---

## 4. 数据库设计

### 4.1 ER 图概述

```
users ─────┬────── orders ─────── time_slots
           │          │                  │
           │          └──── order_items ─┘
           │                            │
activities ─── tickets                  │
           │                     resources
           │                          │
           └─ activity_tickets ────────┘
```

### 4.2 核心表结构

#### 4.2.1 用户表 (users)

```sql
CREATE TABLE users (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    student_id      VARCHAR(20) UNIQUE NOT NULL COMMENT '学号',
    name            VARCHAR(100) NOT NULL COMMENT '姓名',
    email           VARCHAR(255) UNIQUE NOT NULL,
    phone           VARCHAR(20),
    role            ENUM('undergraduate', 'postgraduate', 'teacher', 'admin')
                    DEFAULT 'undergraduate' COMMENT '用户角色',
    credit_score    INT DEFAULT 100 COMMENT '信用分',
    no_show_count   INT DEFAULT 0 COMMENT '爽约次数',
    quota_hours     INT DEFAULT 10 COMMENT '本周可用额度（小时）',
    status          ENUM('active', 'suspended', 'deleted') DEFAULT 'active',
    password_hash   VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_student_id (student_id),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.2 资源表 (resources) - 多态基表

```sql
CREATE TABLE resources (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    type        ENUM('academic', 'sports') NOT NULL COMMENT '资源类型',
    name        VARCHAR(255) NOT NULL COMMENT '资源名称',
    code        VARCHAR(50) UNIQUE NOT NULL COMMENT '资源编码',
    capacity    INT NOT NULL COMMENT '容纳人数',
    location    VARCHAR(255) COMMENT '位置',
    floor       INT COMMENT '楼层',
    building    VARCHAR(100) COMMENT '楼宇',
    image_url   VARCHAR(500) COMMENT '图片',
    description TEXT COMMENT '描述',
    status      ENUM('active', 'inactive', 'maintenance') DEFAULT 'active',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_location (building, floor)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.3 学术空间属性表 (academic_spaces)

```sql
CREATE TABLE academic_spaces (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    resource_id     BIGINT UNIQUE NOT NULL REFERENCES resources(id),
    buffer_minutes  INT DEFAULT 5 COMMENT '前后缓冲时间（分钟）',
    min_duration    INT DEFAULT 30 COMMENT '最小预订时长（分钟）',
    max_duration    INT DEFAULT 240 COMMENT '最大预订时长（分钟）',
    allow_recurring BOOLEAN DEFAULT FALSE COMMENT '是否允许周期性预订',
    equipment       JSON COMMENT '设备列表：投影仪、麦克风、白板等',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.4 体育设施属性表 (sports_facilities)

```sql
CREATE TABLE sports_facilities (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    resource_id     BIGINT UNIQUE NOT NULL REFERENCES resources(id),
    slot_duration   INT DEFAULT 60 COMMENT '槽位时长（分钟）',
    combinable      BOOLEAN DEFAULT TRUE COMMENT '是否可组合预订',
    court_count     INT DEFAULT 1 COMMENT '场地数量（组合用）',
    sport_type      ENUM('basketball', 'tennis', 'badminton', 'football', 'pingpong', 'other')
                    NOT NULL COMMENT '运动类型',
    indoor          BOOLEAN DEFAULT TRUE COMMENT '室内/室外',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.5 时间槽表 (time_slots) - 核心冲突检测表

```sql
CREATE TABLE time_slots (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    resource_id     BIGINT NOT NULL REFERENCES resources(id),
    slot_date       DATE NOT NULL COMMENT '预订日期',
    start_time      TIME NOT NULL COMMENT '开始时间',
    end_time        TIME NOT NULL COMMENT '结束时间',
    buffer_start    TIME NOT NULL COMMENT '缓冲开始时间',
    buffer_end      TIME NOT NULL COMMENT '缓冲结束时间',
    status          ENUM('available', 'locked', 'booked', 'cancelled') DEFAULT 'available',
    booking_id      BIGINT COMMENT '关联订单ID',
    version         INT DEFAULT 0 COMMENT '乐观锁版本',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- 唯一约束：(resource_id, slot_date, start_time, end_time) 防止重复预订
    UNIQUE KEY uk_resource_slot (resource_id, slot_date, start_time, end_time),
    INDEX idx_resource_date (resource_id, slot_date),
    INDEX idx_buffer_overlap (resource_id, slot_date, buffer_start, buffer_end),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.6 活动表 (activities)

```sql
CREATE TABLE activities (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    title           VARCHAR(255) NOT NULL COMMENT '活动标题',
    description     TEXT COMMENT '活动描述',
    organizer       VARCHAR(100) COMMENT '主办方',
    speaker         VARCHAR(100) COMMENT '主讲人',
    location        VARCHAR(255) COMMENT '活动地点',
    activity_type   ENUM('lecture', 'concert', 'competition', 'exhibition', 'other')
                    NOT NULL,
    start_time      DATETIME NOT NULL COMMENT '活动开始时间',
    end_time        DATETIME NOT NULL COMMENT '活动结束时间',
    total_tickets   INT NOT NULL COMMENT '总票数',
    remaining_tickets INT NOT NULL COMMENT '剩余票数',
    price           DECIMAL(10,2) DEFAULT 0.00 COMMENT '票价',
    seckill_start   DATETIME COMMENT '秒杀开始时间',
    seckill_end     DATETIME COMMENT '秒杀结束时间',
    status          ENUM('draft', 'seckill', 'ongoing', 'ended', 'cancelled') DEFAULT 'draft',
    cover_image     VARCHAR(500) COMMENT '封面图',
    view_count      INT DEFAULT 0 COMMENT '浏览次数',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_seckill_time (seckill_start, seckill_end),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.7 订单表 (orders) - 状态机核心

```sql
CREATE TABLE orders (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_no            VARCHAR(32) UNIQUE NOT NULL COMMENT '订单号',
    user_id             BIGINT NOT NULL REFERENCES users(id),
    order_type          ENUM('space', 'activity') NOT NULL COMMENT '订单类型',
    status              ENUM('pending', 'confirmed', 'paid', 'cancelled', 'no_show', 'completed')
                        DEFAULT 'pending' COMMENT '订单状态',
    total_amount        DECIMAL(10,2) DEFAULT 0.00 COMMENT '总金额',
    credit_deduction    INT DEFAULT 0 COMMENT '信用扣减分',
    payment_deadline    DATETIME COMMENT '支付截止时间',
    paid_at             DATETIME COMMENT '支付时间',
    cancelled_at        DATETIME COMMENT '取消时间',
    cancel_reason       VARCHAR(255) COMMENT '取消原因',
    version             INT DEFAULT 0 COMMENT '乐观锁版本，防幽灵支付',
    remark              TEXT COMMENT '备注',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_status (status),
    INDEX idx_payment_deadline (payment_deadline)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.8 订单明细表 (order_items)

```sql
CREATE TABLE order_items (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id        BIGINT NOT NULL REFERENCES orders(id),
    resource_id     BIGINT COMMENT '空间资源ID（空间预订时）',
    activity_id     BIGINT COMMENT '活动ID（活动门票时）',
    ticket_count    INT DEFAULT 1 COMMENT '票数',
    unit_price      DECIMAL(10,2) DEFAULT 0.00 COMMENT '单价',
    slot_date       DATE COMMENT '预订日期（空间）',
    start_time      TIME COMMENT '开始时间（空间）',
    end_time        TIME COMMENT '结束时间（空间）',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_order_id (order_id),
    INDEX idx_resource (resource_id),
    INDEX idx_activity (activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.9 规则配置表 (booking_rules)

```sql
CREATE TABLE booking_rules (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    rule_code       VARCHAR(50) UNIQUE NOT NULL COMMENT '规则编码',
    rule_name       VARCHAR(100) NOT NULL COMMENT '规则名称',
    rule_type       ENUM('quota', 'credit', 'duration', 'blacklist', 'custom') NOT NULL,
    rule_config     JSON NOT NULL COMMENT '规则配置（JSON格式）',
    priority        INT DEFAULT 0 COMMENT '执行优先级（越小越先）',
    status          ENUM('enabled', 'disabled') DEFAULT 'enabled',
    effect_start    DATETIME COMMENT '生效开始时间',
    effect_end      DATETIME COMMENT '生效结束时间',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rule_type (rule_type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 4.2.10 操作日志表 (operation_logs)

```sql
CREATE TABLE operation_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT REFERENCES users(id),
    action          VARCHAR(50) NOT NULL COMMENT '操作类型',
    target_type     VARCHAR(50) COMMENT '目标类型',
    target_id       BIGINT COMMENT '目标ID',
    detail          JSON COMMENT '操作详情',
    ip              VARCHAR(45) COMMENT 'IP地址',
    user_agent      VARCHAR(500) COMMENT 'User-Agent',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.3 冲突检测算法

#### 空间预订冲突检测 SQL

```sql
-- 检测时间重叠（包含缓冲期）
-- 新预订：[new_start - buffer, new_end + buffer]
-- 现有预订：[buffer_start, buffer_end]

SELECT COUNT(*) FROM time_slots
WHERE resource_id = ?
  AND slot_date = ?
  AND status IN ('booked', 'locked')
  -- 重叠条件：开始时间 < 新预订.buffer_end AND 结束时间 > 新预订.buffer_start
  AND start_time < ? + INTERVAL 5 MINUTE   -- 新预订.buffer_end
  AND end_time   > ? - INTERVAL 5 MINUTE;  -- 新预订.buffer_start
```

#### 乐观锁更新（防并发超卖）

```sql
UPDATE time_slots
SET status = 'booked',
    booking_id = ?,
    version = version + 1
WHERE id = ?
  AND version = ?          -- 乐观锁版本检查
  AND status = 'available'; -- 必须是可预订状态
```

---

## 5. 核心模块设计

### 5.1 模块 A：多态空间建模与预订系统

#### 5.1.1 空间预订服务流程

```
用户请求预订
     │
     ▼
┌─────────────────┐
│  参数校验       │ ←── 时间格式、时长限制、日期范围
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  规则引擎评估   │ ←── 信用分、额度、权限、黑名单
└────────┬────────┘
         │ 规则通过
         ▼
┌─────────────────┐
│  冲突检测       │ ←── 查询 time_slots 时间重叠
└────────┬────────┘
         │ 无冲突
         ▼
┌─────────────────┐
│  乐观锁插入     │ ←── UPDATE time_slots SET version = version + 1
└────────┬────────┘
         │ 成功
         ▼
┌─────────────────┐
│  创建订单       │ ←── 状态为 pending，15分钟支付倒计时
└────────┬────────┘
         │
         ▼
    返回订单信息
```

#### 5.1.2 核心代码结构 (Go)

```go
// internal/service/space_service.go
package service

type SpaceService struct {
    repo        *repository.SpaceRepository
    ruleEngine  *engine.RuleEngine
    txManager   *db.TransactionManager
}

func (s *SpaceService) CreateBooking(ctx context.Context, req *BookingRequest) (*Order, error) {
    // 1. 参数校验
    if err := s.validateRequest(req); err != nil {
        return nil, err
    }

    // 2. 规则引擎评估
    result, err := s.ruleEngine.Evaluate(ctx, &RuleContext{
        UserID:   req.UserID,
        ResourceID: req.ResourceID,
        Duration:  req.Duration,
        Date:     req.Date,
    })
    if !result.Passed {
        return nil, errors.New(result.Reason)
    }

    // 3. 计算缓冲时间
    slot, err := s.calculateTimeSlot(req)
    if err != nil {
        return nil, err
    }

    // 4. 开启事务
    order, err := s.txManager.WithTransaction(ctx, func(tx *gorm.DB) (*Order, error) {
        // 4.1 乐观锁插入时间槽
        affected, err := s.repo.UpdateSlotWithLock(tx, slot)
        if err != nil || affected == 0 {
            return nil, errors.New("时间段已被预订")
        }

        // 4.2 创建订单
        return s.createOrder(tx, req)
    })

    return order, err
}
```

### 5.2 模块 B：高并发秒杀系统

#### 5.2.1 秒杀架构（三层防护）

```
                        ┌──────────────────┐
                        │   用户请求        │
                        └────────┬─────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │  第一层：限流器          │
                    │  Redis 计数器 / 令牌桶   │
                    │  过滤 90% 无效请求      │
                    └───────────┬────────────┘
                                │ 通过
                                ▼
                    ┌────────────────────────┐
                    │  第二层：Redis 预扣库存   │
                    │  Lua 脚本原子操作        │
                    │  过滤剩余 90%           │
                    └───────────┬────────────┘
                                │ 有库存
                                ▼
                    ┌────────────────────────┐
                    │  第三层：MQ 异步处理     │
                    │  分布式锁 + 事务消息    │
                    │  确保不超卖             │
                    └────────────────────────┘
                                │
                                │ 成功
                                ▼
                         返回下单成功
```

#### 5.2.2 Redis Lua 脚本（原子扣减）

```lua
-- seckill_stock.lua
local stock_key = KEYS[1]
local user_key = KEYS[2]
local activity_id = ARGV[1]
local user_id = ARGV[2]
local quantity = tonumber(ARGV[3])

-- 检查用户是否已购买
local user_bought = redis.call('SISMEMBER', user_key, user_id)
if user_bought == 1 then
    return -2  -- 已购买
end

-- 检查库存
local stock = tonumber(redis.call('GET', stock_key))
if stock < quantity then
    return -1  -- 库存不足
end

-- 扣减库存
redis.call('DECRBY', stock_key, quantity)
-- 标记用户已购买
redis.call('SADD', user_key, user_id)

return 1  -- 成功
```

#### 5.2.3 MQ 事务消息处理

```go
// internal/service/seckill_service.go
func (s *SeckillService) ProcessSeckill(ctx context.Context, req *SeckillRequest) error {
    // 1. Redis 预扣库存
    result, err := s.redis.Eval(SeckillScript, []string{
        fmt.Sprintf("stock:%d", req.ActivityID),
        fmt.Sprintf("user:%d:%d", req.ActivityID, req.UserID),
    }, req.ActivityID, req.UserID, 1)

    if result.(int64) != 1 {
        return errors.New("秒杀失败")
    }

    // 2. 发送事务消息
    msg := &SeckillMessage{
        ActivityID: req.ActivityID,
        UserID:     req.UserID,
        Quantity:   1,
        Timestamp:  time.Now(),
    }

    return s.mq.SendTransactionMessage("seckill_topic", msg, func(halfTx func() error) error {
        // 本地事务：创建订单
        return s.createSeckillOrder(halfTx, msg)
    })
}
```

### 5.3 模块 C：状态机与动态规则引擎

#### 5.3.1 订单状态流转

```
                        ┌─────────────────────────────────────┐
                        │                                     │
                        ▼                                     │
    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ PENDING  │───▶│CONFIRMED │───▶│  PAID    │───▶│COMPLETED │
    │ (待支付)  │    │ (已确认)  │    │ (已支付)  │    │ (已完成)  │
    └────┬─────┘    └──────────┘    └──────────┘    └──────────┘
         │                 ▲
         │                 │
    支付超时/用户取消        │ 用户取消
         │                 │
         ▼                 │
    ┌──────────┐           │
    │CANCELLED │◀──────────┘
    │ (已取消)  │
    └──────────┘

    爽约检测（定时任务）
         │
         ▼
    ┌──────────┐
    │ NO_SHOW  │
    │ (已爽约)  │
    └──────────┘
```

#### 5.3.2 幽灵支付防护

```go
// internal/service/order_service.go

// 支付操作（带乐观锁）
func (s *OrderService) PayOrder(ctx context.Context, orderID, userID int64) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // 1. 查询订单（FOR UPDATE 悲观锁）
        var order Order
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&order, orderID).Error; err != nil {
            return err
        }

        // 2. 校验状态
        if order.Status != "pending" {
            return errors.New("订单状态不允许支付")
        }

        // 3. 检查支付截止时间（防止幽灵支付）
        if time.Now().After(order.PaymentDeadline) {
            return errors.New("订单已超时")
        }

        // 4. 乐观锁更新
        result := tx.Model(&order).
            Where("version = ?", order.Version).
            Updates(map[string]interface{}{
                "status":    "paid",
                "paid_at":   time.Now(),
                "version":   order.Version + 1,
            })

        if result.RowsAffected == 0 {
            return errors.New("并发冲突，请重试")
        }

        return nil
    })
}
```

#### 5.3.3 延迟消息处理超时

```go
// internal/mq/delayed_handler.go

// 延迟队列消费者：处理订单超时
func (h *DelayedOrderHandler) HandleMessage(ctx context.Context, msg *OrderMessage) error {
    order, err := h.repo.GetOrderByID(msg.OrderID)
    if err != nil {
        return err
    }

    // 只有pending状态才处理（防止幽灵支付后又被取消）
    if order.Status != "pending" {
        return nil // 已支付，直接忽略
    }

    // 再次确认超时
    if time.Now().After(order.PaymentDeadline) {
        return h.cancelOrder(order, "支付超时自动取消")
    }

    return nil
}
```

#### 5.3.4 规则引擎实现（责任链模式）

```go
// internal/engine/rule_engine.go

type RuleEngine struct {
    chains []BookingRule
}

type RuleContext struct {
    UserID     int64
    Role       string
    CreditScore int
    NoShowCount int
    QuotaHours  int
    Duration    int // 分钟
    ResourceType string
}

type RuleResult struct {
    Passed     bool
    Reason     string
    Deductions map[string]int // 各类扣减
}

type BookingRule interface {
    Evaluate(ctx *RuleContext) *RuleResult
    Priority() int
}

// 规则1：黑名单检查
type BlacklistRule struct{}

func (r *BlacklistRule) Evaluate(ctx *RuleContext) *RuleResult {
    if isBlacklisted(ctx.UserID) {
        return &RuleResult{Passed: false, Reason: "您在黑名单中，无法预订"}
    }
    return &RuleResult{Passed: true}
}

func (r *BlacklistRule) Priority() int { return 1 }

// 规则2：信用分检查
type CreditScoreRule struct{}

func (r *CreditScoreRule) Evaluate(ctx *RuleContext) *RuleResult {
    if ctx.CreditScore < 60 {
        return &RuleResult{Passed: false, Reason: "信用分低于60，无法预订"}
    }
    return &RuleResult{Passed: true}
}

func (r *CreditScoreRule) Priority() int { return 2 }

// 规则3：爽约扣减额度
type NoShowDeductionRule struct{}

func (r *NoShowDeductionRule) Evaluate(ctx *RuleContext) *RuleResult {
    deductions := make(map[string]int)
    deduction := ctx.NoShowCount * 10
    deductions["quota_hours"] = deduction
    return &RuleResult{Passed: true, Deductions: deductions}
}

func (r *NoShowDeductionRule) Priority() int { return 3 }

// 规则4：时长限制
type DurationLimitRule struct{}

func (r *DurationLimitRule) Evaluate(ctx *RuleContext) *RuleResult {
    maxHours := 4
    switch ctx.Role {
    case "undergraduate":
        maxHours = 2
    case "postgraduate":
        maxHours = 3
    case "teacher":
        maxHours = 6
    }

    if ctx.Duration > maxHours*60 {
        return &RuleResult{Passed: false,
            Reason: fmt.Sprintf("%s最大预订时长为%d小时", ctx.Role, maxHours)}
    }
    return &RuleResult{Passed: true}
}

func (r *DurationLimitRule) Priority() int { return 4 }

// 执行规则链
func (e *RuleEngine) Evaluate(ctx *RuleContext) *RuleResult {
    totalDeductions := make(map[string]int)

    for _, rule := range e.chains {
        result := rule.Evaluate(ctx)
        if !result.Passed {
            return result
        }
        for k, v := range result.Deductions {
            totalDeductions[k] += v
        }
    }

    return &RuleResult{Passed: true, Deductions: totalDeductions}
}
```

---

## 6. API 接口设计

### 6.1 API 规范

- 基础路径：`/api/v1`
- 认证方式：`Bearer Token` (JWT)
- 数据格式：`application/json`
- 统一响应结构

```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

### 6.2 用户接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /auth/register | 用户注册 |
| POST | /auth/login | 用户登录 |
| POST | /auth/refresh | 刷新Token |
| GET | /users/me | 获取当前用户信息 |
| PUT | /users/me | 更新用户信息 |

### 6.3 空间预订接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /spaces | 获取空间列表（支持类型筛选） |
| GET | /spaces/:id | 获取空间详情 |
| GET | /spaces/:id/slots | 获取空间可用时间槽 |
| POST | /spaces/bookings | 创建预订 |
| GET | /bookings | 获取用户预订列表 |
| GET | /bookings/:id | 获取预订详情 |
| DELETE | /bookings/:id | 取消预订 |

### 6.4 活动秒杀接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /activities | 获取活动列表 |
| GET | /activities/:id | 获取活动详情 |
| POST | /activities/:id/seckill | 参与秒杀 |
| GET | /activities/:id/ticket | 获取我的票 |

### 6.5 订单接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /orders | 获取订单列表 |
| GET | /orders/:id | 获取订单详情 |
| POST | /orders/:id/pay | 支付订单 |
| POST | /orders/:id/cancel | 取消订单 |

---

## 7. 前端架构

### 7.1 项目结构

```
frontend/
├── src/
│   ├── main.tsx                 # 入口文件
│   ├── App.tsx                  # 根组件
│   ├── components/              # 通用组件
│   │   ├── ui/                  # 基础 UI 组件
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Modal.tsx
│   │   │   └── ...
│   │   ├── layout/              # 布局组件
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Footer.tsx
│   │   └── common/              # 通用业务组件
│   │       ├── SpaceCard.tsx
│   │       ├── ActivityCard.tsx
│   │       └── OrderStatus.tsx
│   ├── features/                # 业务功能模块
│   │   ├── auth/                # 认证模块
│   │   │   ├── Login.tsx
│   │   │   ├── Register.tsx
│   │   │   └── hooks/useAuth.ts
│   │   ├── spaces/              # 空间预订模块
│   │   │   ├── SpaceList.tsx
│   │   │   ├── SpaceDetail.tsx
│   │   │   ├── BookingModal.tsx
│   │   │   └── hooks/useSpaces.ts
│   │   ├── activities/          # 活动秒杀模块
│   │   │   ├── ActivityList.tsx
│   │   │   ├── ActivityDetail.tsx
│   │   │   ├── SeckillButton.tsx
│   │   │   └── hooks/useActivities.ts
│   │   └── orders/              # 订单模块
│   │       ├── OrderList.tsx
│   │       ├── OrderDetail.tsx
│   │       └── hooks/useOrders.ts
│   ├── hooks/                   # 通用 Hooks
│   │   ├── useApi.ts
│   │   ├── useToast.ts
│   │   └── useCountdown.ts
│   ├── stores/                  # 状态管理
│   │   ├── authStore.ts
│   │   ├── spaceStore.ts
│   │   └── orderStore.ts
│   ├── services/                # API 服务层
│   │   ├── api.ts              # 基础请求封装
│   │   ├── auth.ts
│   │   ├── spaces.ts
│   │   ├── activities.ts
│   │   └── orders.ts
│   ├── types/                   # TypeScript 类型定义
│   │   ├── api.ts
│   │   ├── space.ts
│   │   ├── activity.ts
│   │   └── order.ts
│   ├── utils/                   # 工具函数
│   │   ├── format.ts           # 格式化工具
│   │   ├── validation.ts       # 校验工具
│   │   └── storage.ts          # 本地存储
│   ├── styles/                  # 全局样式
│   │   └── globals.css
│   └── pages/                   # 页面组件
│       ├── Home.tsx
│       ├── NotFound.tsx
│       └── Forbidden.tsx
├── public/                      # 静态资源
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
├── Dockerfile
└── docker-compose.yml
```

### 7.2 核心组件设计

#### 7.2.1 预订时间选择器

```tsx
// features/spaces/components/BookingModal.tsx
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { timeApi } from '@/services/spaces';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';

interface BookingModalProps {
    spaceId: number;
    spaceType: 'academic' | 'sports';
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

export function BookingModal({
    spaceId,
    spaceType,
    isOpen,
    onClose,
    onSuccess
}: BookingModalProps) {
    const [selectedDate, setSelectedDate] = useState<string>('');
    const [selectedSlot, setSelectedSlot] = useState<TimeSlot | null>(null);

    // 获取可用时间槽
    const { data: slots, isLoading } = useQuery({
        queryKey: ['timeSlots', spaceId, selectedDate],
        queryFn: () => timeApi.getAvailableSlots(spaceId, selectedDate),
        enabled: !!selectedDate,
    });

    const handleSubmit = async () => {
        if (!selectedSlot) return;
        await bookingApi.create({
            resourceId: spaceId,
            slotId: selectedSlot.id,
        });
        onSuccess();
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="预订空间">
            <div className="space-y-4">
                <input
                    type="date"
                    value={selectedDate}
                    onChange={(e) => setSelectedDate(e.target.value)}
                    min={new Date().toISOString().split('T')[0]}
                />

                {isLoading ? (
                    <div className="text-center py-4">加载中...</div>
                ) : (
                    <div className="grid grid-cols-4 gap-2">
                        {slots?.map((slot) => (
                            <button
                                key={slot.id}
                                disabled={slot.status !== 'available'}
                                onClick={() => setSelectedSlot(slot)}
                                className={`p-2 text-sm rounded ${
                                    selectedSlot?.id === slot.id
                                        ? 'bg-blue-500 text-white'
                                        : slot.status === 'available'
                                        ? 'bg-gray-100 hover:bg-gray-200'
                                        : 'bg-gray-300 text-gray-500'
                                }`}
                            >
                                {slot.startTime} - {slot.endTime}
                            </button>
                        ))}
                    </div>
                )}

                <Button onClick={handleSubmit} disabled={!selectedSlot}>
                    确认预订
                </Button>
            </div>
        </Modal>
    );
}
```

#### 7.2.2 秒杀倒计时按钮

```tsx
// features/activities/components/SeckillButton.tsx
import { useState, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { seckillApi } from '@/services/activities';
import { Button } from '@/components/ui/Button';
import { useToast } from '@/hooks/useToast';

interface SeckillButtonProps {
    activityId: number;
    seckillStart: Date;
    remainingTickets: number;
}

export function SeckillButton({
    activityId,
    seckillStart,
    remainingTickets
}: SeckillButtonProps) {
    const [countdown, setCountdown] = useState<number>(0);
    const [status, setStatus] = useState<'waiting' | 'seckill' | 'end'>('waiting');
    const { showToast } = useToast();

    useEffect(() => {
        const timer = setInterval(() => {
            const now = Date.now();
            const diff = seckillStart.getTime() - now;

            if (diff <= 0) {
                setStatus(remainingTickets > 0 ? 'seckill' : 'end');
                setCountdown(0);
            } else {
                setCountdown(Math.ceil(diff / 1000));
            }
        }, 1000);

        return () => clearInterval(timer);
    }, [seckillStart, remainingTickets]);

    const seckillMutation = useMutation({
        mutationFn: () => seckillApi.doSeckill(activityId),
        onSuccess: () => {
            showToast('秒杀成功！', 'success');
        },
        onError: (error) => {
            showToast(error.message, 'error');
        },
    });

    if (status === 'waiting') {
        return (
            <Button disabled className="w-full">
                距秒杀开始：{Math.floor(countdown / 3600)}:
                {String(Math.floor((countdown % 3600) / 60)).padStart(2, '0')}:
                {String(countdown % 60).padStart(2, '0')}
            </Button>
        );
    }

    if (status === 'end') {
        return (
            <Button disabled className="w-full">
                已售罄
            </Button>
        );
    }

    return (
        <Button
            onClick={() => seckillMutation.mutate()}
            loading={seckillMutation.isPending}
            className="w-full bg-red-500 hover:bg-red-600"
        >
            立即秒杀
        </Button>
    );
}
```

### 7.3 无障碍设计要点

```tsx
// components/ui/Button.tsx
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger';
    size?: 'sm' | 'md' | 'lg';
}

// 关键无障碍属性：
// 1. aria-label：图标按钮必须提供
// 2. aria-disabled：禁用状态需同步
// 3. role="status"：状态变化需通知屏幕阅读器
// 4. focus-visible：键盘焦点样式清晰

export function Button({
    variant = 'primary',
    size = 'md',
    className,
    disabled,
    children,
    ...props
}: ButtonProps) {
    return (
        <button
            className={cn(
                'rounded-lg font-medium transition-colors',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
                'disabled:opacity-50 disabled:cursor-not-allowed',
                variantStyles[variant],
                sizeStyles[size],
                className
            )}
            disabled={disabled}
            aria-disabled={disabled}
            {...props}
        >
            {children}
        </button>
    );
}
```

### 7.4 性能优化

```tsx
// hooks/useCountdown.ts
// 使用 requestAnimationFrame 优化的倒计时

export function useCountdown(endTime: Date, onEnd?: () => void) {
    const [seconds, setSeconds] = useState(() => {
        return Math.max(0, Math.floor((endTime.getTime() - Date.now()) / 1000));
    });

    useEffect(() => {
        let rafId: number;
        let lastTick = Date.now();

        const tick = () => {
            const now = Date.now();
            const elapsed = now - lastTick;

            if (elapsed >= 1000) {
                lastTick = now;
                setSeconds((prev) => {
                    const next = prev - 1;
                    if (next <= 0) {
                        onEnd?.();
                        return 0;
                    }
                    return next;
                });
            }

            rafId = requestAnimationFrame(tick);
        };

        rafId = requestAnimationFrame(tick);
        return () => cancelAnimationFrame(rafId);
    }, [endTime, onEnd]);

    return {
        hours: Math.floor(seconds / 3600),
        minutes: Math.floor((seconds % 3600) / 60),
        seconds: seconds % 60,
    };
}
```

---

## 8. 容器化部署

### 8.1 后端多阶段 Dockerfile

```dockerfile
# backend/Dockerfile
# ============ 编译阶段 ============
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译（优化二进制体积）
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o server ./cmd/server

# ============ 运行阶段 ============
FROM alpine:3.19

WORKDIR /app

# 安装时区数据
RUN apk add --no-cache tzdata

# 从编译阶段复制二进制
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
```

### 8.2 前端 Dockerfile

```dockerfile
# frontend/Dockerfile
# ============ 编译阶段 ============
FROM node:20-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY package*.json ./
RUN npm ci --only=production

# 复制源码
COPY . .

# 构建（注入环境变量）
ARG VITE_API_BASE_URL=http://api.example.com
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

# ============ 运行阶段（Nginx） ============
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 8.3 Nginx 配置（前端）

```nginx
# frontend/nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # 启用 gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;

        # 安全头
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
    }

    # API 代理
    location /api/ {
        proxy_pass http://api-gateway:8080/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 8.4 Docker Compose 编排

```yaml
# docker-compose.yml
version: '3.8'

services:
  # ============ 前端 ============
  frontend:
    build:
      context: ./frontend
      args:
        VITE_API_BASE_URL: http://localhost/api
    ports:
      - "80:80"
    depends_on:
      api-gateway:
        condition: service_healthy
    networks:
      - app-network

  # ============ API 网关 ============
  api-gateway:
    build: ./gateway
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - space-service
      - seckill-service
      - order-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
    networks:
      - app-network

  # ============ 用户服务 ============
  user-service:
    build: ./services/user
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - DB_NAME=smartcampus
      - DB_USER=root
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - app-network

  # ============ 空间预订服务 ============
  space-service:
    build: ./services/space
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - MQ_NAMESRV_ADDR=mq:9876
    depends_on:
      db:
        condition: service_healthy
      mq:
        condition: service_healthy
    networks:
      - app-network

  # ============ 秒杀服务（多实例） ============
  seckill-service:
    build: ./services/seckill
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 512M
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - MQ_NAMESRV_ADDR=mq:9876
    depends_on:
      redis:
        condition: service_healthy
      mq:
        condition: service_healthy
    networks:
      - app-network

  # ============ 订单服务 ============
  order-service:
    build: ./services/order
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - MQ_NAMESRV_ADDR=mq:9876
    depends_on:
      db:
        condition: service_healthy
      mq:
        condition: service_healthy
    networks:
      - app-network

  # ============ 消息队列 ============
  mq:
    image: apache/rocketmq:5.0
    ports:
      - "9876:9876"
    environment:
      - NAMESRV_PORT=9876
    volumes:
      - mq-data:/home/rocketmq/data
    networks:
      - app-network

  # ============ Redis ============
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 512mb
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - app-network

  # ============ 数据库 ============
  db:
    image: mysql:8
    environment:
      - MYSQL_ROOT_PASSWORD=${DB_PASSWORD}
      - MYSQL_DATABASE=smartcampus
    volumes:
      - mysql-data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

volumes:
  mysql-data:
  redis-data:
  mq-data:

networks:
  app-network:
    driver: bridge
```

---

## 9. CI/CD 流水线

### 9.1 GitHub Actions 工作流

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # ============ 代码检查 ============
  lint:
    name: Lint & Format Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Go Lint
        run: |
          go install golang.org/x/tools/cmd/golangci-lint@latest
          golangci-lint run ./...

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Frontend Lint
        working-directory: ./frontend
        run: |
          npm ci
          npm run lint

  # ============ 单元测试 ============
  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8
        env:
          MYSQL_ROOT_PASSWORD: test
          MYSQL_DATABASE: test_db
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true

      - name: Run Go Tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: coverage.out

  # ============ 构建与推送镜像 ============
  build:
    name: Build & Push Images
    runs-on: ubuntu-latest
    needs: [lint, test]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and Push Backend
        uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: true
          tags: |
            ghcr.io/${{ github.repository }}/backend:${{ github.sha }}
            ghcr.io/${{ github.repository }}/backend:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Build and Push Frontend
        uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: true
          tags: |
            ghcr.io/${{ github.repository }}/frontend:${{ github.sha }}
            ghcr.io/${{ github.repository }}/frontend:latest

  # ============ 部署到测试环境 ============
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'
    environment: staging

    steps:
      - name: Deploy to Server
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.STAGING_HOST }}
          username: ${{ secrets.STAGING_USER }}
          key: ${{ secrets.STAGING_SSH_KEY }}
          script: |
            cd /app/smart-campus
            docker-compose pull
            docker-compose up -d
            docker image prune -f

  # ============ 端到端测试 ============
  e2e:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: [deploy-staging]
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install Playwright
        working-directory: ./frontend
        run: |
          npm ci
          npx playwright install --with-deps

      - name: Run E2E Tests
        working-directory: ./frontend
        run: |
          npx playwright test
        env:
          BASE_URL: ${{ secrets.STAGING_URL }}
```

---

## 10. 开发规范

### 10.1 Git 提交规范

```
<type>(<scope>): <subject>

feat(user): add user registration endpoint
fix(booking): resolve time slot conflict race condition
docs(api): update API documentation
style(ui): adjust button padding for mobile
refactor(order): extract status machine to separate module
test(seckill): add unit tests for Redis Lua script
chore(deps): upgrade Redis client to 5.0
perf(api): add response caching for space list
```

**Type 类型**：
- `feat`：新功能
- `fix`：Bug 修复
- `docs`：文档更新
- `style`：代码格式（不影响功能）
- `refactor`：重构
- `test`：测试相关
- `chore`：构建/工具相关
- `perf`：性能优化

### 10.2 API 命名规范

```go
// 路由分组
router := gin.Default()

// 按资源分组
v1 := router.Group("/api/v1")
{
    // 空间资源
    spaces := v1.Group("/spaces")
    spaces.GET("", listSpaces)
    spaces.GET("/:id", getSpace)
    spaces.POST("/bookings", createBooking)

    // 活动资源
    activities := v1.Group("/activities")
    activities.GET("", listActivities)
    activities.GET("/:id", getActivity)
    activities.POST("/:id/seckill", doSeckill)
}
```

### 10.3 错误处理规范

```go
// internal/pkg/errors/errors.go

type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
}

// 预定义错误
var (
    ErrNotFound       = &AppError{Code: "NOT_FOUND", Message: "资源不存在"}
    ErrUnauthorized   = &AppError{Code: "UNAUTHORIZED", Message: "未授权"}
    ErrForbidden      = &AppError{Code: "FORBIDDEN", Message: "无权限"}
    ErrConflict       = &AppError{Code: "CONFLICT", Message: "资源冲突"}
    ErrInternalServer = &AppError{Code: "INTERNAL_SERVER", Message: "服务器内部错误"}
)
```

### 10.4 日志规范

```go
// 使用结构化日志
log.Info("booking created",
    "order_id", order.ID,
    "user_id", user.ID,
    "resource_id", resource.ID,
    "slot_date", slot.Date,
    "duration_ms", time.Since(start).Milliseconds(),
)
```

### 10.5 安全规范

```go
// 1. 密码加密
bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 2. SQL 注入防护（使用 GORM 的参数化查询）
db.Where("user_id = ?", userID).First(&order)

// 3. XSS 防护
template.HTMLEscapeString(userInput)

// 4. JWT 验证
claims := &jwt.MapClaims{}
token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
    return []byte(os.Getenv("JWT_SECRET")), nil
})

// 5. 限流
rateLimiter := golang.org/x/time/rate
limiter := rateLimiter.NewLimiter(rate.Limit(100), 100)
if !limiter.Allow() {
    c.JSON(429, gin.H{"error": "请求过于频繁"})
    return
}
```

---

## 附录

### A. 环境变量配置

```bash
# .env.example
DB_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_key
REDIS_PASSWORD=

# 服务端口
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
SPACE_SERVICE_PORT=8082
SECKILL_SERVICE_PORT=8083
ORDER_SERVICE_PORT=8084
```

### B. 常用命令

```bash
# 本地开发
make run          # 启动所有服务
make test         # 运行测试
make lint         # 代码检查
make migrate      # 数据库迁移

# Docker
docker-compose up -d          # 启动所有容器
docker-compose logs -f api    # 查看 API 日志
docker-compose exec db mysql  # 连接数据库

# 前端
cd frontend
npm run dev        # 开发模式
npm run build      # 生产构建
npm run preview    # 预览构建结果
npm run test       # 运行测试
npm run lint       # 代码检查
```

### C. 目录结构总览

```
smart-campus/
├── backend/
│   ├── cmd/
│   │   ├── gateway/
│   │   ├── user-service/
│   │   ├── space-service/
│   │   ├── seckill-service/
│   │   └── order-service/
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   ├── pkg/
│   │   └── engine/
│   ├── configs/
│   ├── scripts/
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── Dockerfile
│   └── nginx.conf
├── scripts/
│   └── init.sql
├── docker-compose.yml
├── Makefile
└── README.md
```
