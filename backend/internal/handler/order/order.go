package order

import (
	"net/http"
	"time"

	"smart-campus/internal/model"
	"smart-campus/internal/pkg/errors"
	"smart-campus/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var orders []model.Order
	if err := h.db.Preload("Items").Where("user_id = ?", userID).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *Handler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orderID := c.Param("id")

	var order model.Order
	if err := h.db.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *Handler) PayOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orderID := c.Param("id")

	// 使用事务
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		// 悲观锁查询
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			return err
		}

		// 检查是否是该用户的订单
		if order.UserID != userID {
			return errors.ErrForbidden
		}

		// 检查状态
		if order.Status != model.OrderStatusPending {
			return errors.ErrOrderNotPending
		}

		// 检查支付截止时间
		if time.Now().After(*order.PaymentDeadline) {
			return errors.ErrOrderTimeout
		}

		// 乐观锁更新
		result := tx.Model(&order).
			Where("version = ?", order.Version).
			Updates(map[string]interface{}{
				"status":  model.OrderStatusPaid,
				"paid_at": time.Now(),
				"version": order.Version + 1,
			})

		if result.RowsAffected == 0 {
			return errors.New("concurrent conflict, please retry")
		}

		return nil
	})

	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			c.JSON(appErr.HTTP, gin.H{"error": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment successful"})
}

func (h *Handler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orderID := c.Param("id")

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.First(&order, orderID).Error; err != nil {
			return err
		}

		// 检查是否是该用户的订单
		if order.UserID != userID {
			return errors.ErrForbidden
		}

		// 检查状态
		if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusConfirmed {
			return errors.ErrOrderNotPending
		}

		// 更新状态
		now := time.Now()
		order.Status = model.OrderStatusCancelled
		order.CancelledAt = &now
		order.CancelReason = "user cancelled"

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// 如果是空间预订，释放时间槽
		if order.OrderType == "space" && len(order.Items) > 0 {
			item := order.Items[0]
			if item.ResourceID != nil && item.SlotDate != nil && item.StartTime != nil {
				tx.Model(&model.TimeSlot{}).
					Where("resource_id = ? AND slot_date = ? AND start_time = ?",
						*item.ResourceID, *item.SlotDate, *item.StartTime).
					Updates(map[string]interface{}{
						"status":     "available",
						"booking_id": nil,
					})
			}
		}

		// 如果是活动预订，回补库存
		if order.OrderType == "activity" && len(order.Items) > 0 {
			item := order.Items[0]
			if item.ActivityID != nil {
				tx.Model(&model.Activity{}).
					Where("id = ?", *item.ActivityID).
					Update("remaining_tickets", gorm.Expr("remaining_tickets + ?", item.TicketCount))
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			c.JSON(appErr.HTTP, gin.H{"error": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}
