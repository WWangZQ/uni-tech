package seckill

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"smart-campus/internal/model"
	"smart-campus/internal/pkg/errors"
	"smart-campus/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Handler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHandler(db *gorm.DB, redis *redis.Client) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
	}
}

type ActivityListResponse struct {
	Activities []model.Activity `json:"activities"`
	Total     int64           `json:"total"`
}

func (h *Handler) ListActivities(c *gin.Context) {
	var activities []model.Activity
	var total int64

	query := h.db.Model(&model.Activity{}).Where("status != ?", "cancelled")

	// 按状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	if err := query.Order("start_time DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"total":     total,
	})
}

func (h *Handler) GetActivity(c *gin.Context) {
	id := c.Param("id")

	var activity model.Activity
	if err := h.db.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activity"})
		return
	}

	// 增加浏览次数
	h.db.Model(&activity).Update("view_count", activity.ViewCount+1)

	c.JSON(http.StatusOK, gin.H{"activity": activity})
}

func (h *Handler) DoSeckill(c *gin.Context) {
	userID := middleware.GetUserID(c)
	activityID := c.Param("id")

	ctx := context.Background()

	// 获取活动信息
	var activity model.Activity
	if err := h.db.First(&activity, activityID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": "activity not found"})
		return
	}

	// 检查秒杀时间
	now := time.Now()
	if activity.SeckillStart != nil && now.Before(*activity.SeckillStart) {
		c.JSON(http.StatusForbidden, gin.H{"error": "seckill has not started"})
		return
	}
	if activity.SeckillEnd != nil && now.After(*activity.SeckillEnd) {
		c.JSON(http.StatusForbidden, gin.H{"error": "seckill has ended"})
		return
	}

	// 检查库存
	if activity.RemainingTickets <= 0 {
		c.JSON(errors.ErrStockNotEnough.HTTP, gin.H{"error": errors.ErrStockNotEnough.Message})
		return
	}

	// Redis Lua 脚本 - 原子操作
	luaScript := `
		local stock_key = KEYS[1]
		local user_key = KEYS[2]
		local user_id = ARGV[1]
		local quantity = tonumber(ARGV[2])

		-- 检查用户是否已购买
		local already = redis.call('SISMEMBER', user_key, user_id)
		if already == 1 then
			return -2
		end

		-- 检查库存
		local stock = tonumber(redis.call('GET', stock_key))
		if stock == nil or stock < quantity then
			return -1
		end

		-- 原子扣减库存
		redis.call('DECRBY', stock_key, quantity)
		redis.call('SADD', user_key, user_id)

		return 1
	`

	stockKey := fmt.Sprintf("seckill:stock:%s", activityID)
	userKey := fmt.Sprintf("seckill:user:%s", activityID)

	result, err := h.redis.Eval(ctx, luaScript, []string{stockKey, userKey}, userID, 1).Int64()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seckill failed"})
		return
	}

	if result == -2 {
		c.JSON(errors.ErrAlreadyPurchased.HTTP, gin.H{"error": errors.ErrAlreadyPurchased.Message})
		return
	}

	if result == -1 {
		// 回补库存
		c.JSON(errors.ErrStockNotEnough.HTTP, gin.H{"error": errors.ErrStockNotEnough.Message})
		return
	}

	// 创建订单
	orderNo := fmt.Sprintf("SK%d%d%s", time.Now().UnixNano(), userID, activityID)
	paymentDeadline := time.Now().Add(15 * time.Minute)

	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		OrderType:       "activity",
		Status:          model.OrderStatusPending,
		TotalAmount:     activity.Price,
		PaymentDeadline: &paymentDeadline,
		Version:         0,
	}

	if err := h.db.Create(order).Error; err != nil {
		// 回补 Redis 库存
		h.redis.Incr(ctx, stockKey)
		h.redis.SRem(ctx, userKey, userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// 创建订单项
	orderItem := &model.OrderItem{
		OrderID:     order.ID,
		ActivityID:  &activity.ID,
		TicketCount: 1,
		UnitPrice:   activity.Price,
	}

	if err := h.db.Create(orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order item"})
		return
	}

	// 减少数据库库存
	h.db.Model(&activity).Where("remaining_tickets > 0").Update("remaining_tickets", gorm.Expr("remaining_tickets - 1"))

	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"message": fmt.Sprintf("秒杀成功，请在 %v 前完成支付", paymentDeadline.Format("2006-01-02 15:04:05")),
	})
}

func (h *Handler) GetMyTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	activityID := c.Param("id")

	var order model.Order
	if err := h.db.Preload("Items").Where("user_id = ? AND order_type = ? AND activity_id = ?",
		userID, "activity", activityID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "you have not purchased this activity"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket": order})
}
