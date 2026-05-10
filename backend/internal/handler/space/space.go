package space

import (
	"fmt"
	"net/http"
	"time"

	"smart-campus/internal/model"
	"smart-campus/internal/pkg/errors"
	"smart-campus/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

type SpaceListResponse struct {
	Spaces []model.Resource `json:"spaces"`
	Total  int64           `json:"total"`
}

func (h *Handler) ListSpaces(c *gin.Context) {
	var spaces []model.Resource
	var total int64

	query := h.db.Model(&model.Resource{}).Where("status = ?", "active")

	// 按类型筛选
	if spaceType := c.Query("type"); spaceType != "" {
		query = query.Where("type = ?", spaceType)
	}

	// 按楼栋筛选
	if building := c.Query("building"); building != "" {
		query = query.Where("building = ?", building)
	}

	query.Count(&total)
	if err := query.Preload("AcademicSpace").Preload("SportsFacility").Find(&spaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list spaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"spaces": spaces,
		"total":  total,
	})
}

func (h *Handler) GetSpace(c *gin.Context) {
	id := c.Param("id")

	var space model.Resource
	if err := h.db.Preload("AcademicSpace").Preload("SportsFacility").First(&space, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get space"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"space": space})
}

type SlotResponse struct {
	ID          int64  `json:"id"`
	Date        string `json:"date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Status      string `json:"status"`
}

func (h *Handler) GetAvailableSlots(c *gin.Context) {
	resourceID := c.Param("id")
	dateStr := c.Query("date")

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	var slots []model.TimeSlot
	if err := h.db.Where("resource_id = ? AND slot_date = ?", resourceID, date).
		Order("start_time").Find(&slots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get slots"})
		return
	}

	response := make([]SlotResponse, len(slots))
	for i, slot := range slots {
		response[i] = SlotResponse{
			ID:        slot.ID,
			Date:      slot.SlotDate.Format("2006-01-02"),
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
			Status:    slot.Status,
		}
	}

	c.JSON(http.StatusOK, gin.H{"slots": response})
}

type CreateBookingRequest struct {
	ResourceID int64  `json:"resource_id" binding:"required"`
	SlotID     int64  `json:"slot_id" binding:"required"`
	Date       string `json:"date" binding:"required"`
}

func (h *Handler) CreateBooking(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取空间信息
	var resource model.Resource
	if err := h.db.Preload("AcademicSpace").Preload("SportsFacility").First(&resource, req.ResourceID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": "space not found"})
		return
	}

	// 获取时间槽
	date, _ := time.Parse("2006-01-02", req.Date)
	var slot model.TimeSlot
	if err := h.db.Where("id = ? AND resource_id = ? AND slot_date = ?", req.SlotID, req.ResourceID, date).
		First(&slot).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": "slot not found"})
		return
	}

	// 检查槽位状态
	if slot.Status != "available" {
		c.JSON(errors.ErrSlotUnavailable.HTTP, gin.H{"error": errors.ErrSlotUnavailable.Message})
		return
	}

	// 获取用户信息，检查信用分
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": "user not found"})
		return
	}

	// 规则检查
	if user.CreditScore < 60 {
		c.JSON(http.StatusForbidden, gin.H{"error": "credit score too low"})
		return
	}

	result := h.db.Model(&slot).Where("id = ? AND version = ? AND status = ?", slot.ID, slot.Version, "available").
		Updates(map[string]interface{}{
			"status":     "booked",
			"booking_id": userID,
			"version":    slot.Version + 1,
		})

	if result.RowsAffected == 0 {
		c.JSON(errors.ErrSlotConflict.HTTP, gin.H{"error": "slot was booked by another user"})
		return
	}

	// 计算到期时间（15分钟后）
	paymentDeadline := time.Now().Add(15 * time.Minute)

	// 创建订单
	orderNo := fmt.Sprintf("BK%d%d%d", time.Now().UnixNano(), userID, slot.ID)
	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		OrderType:       "space",
		Status:          model.OrderStatusPending,
		TotalAmount:     0,
		PaymentDeadline: &paymentDeadline,
		Version:         0,
	}

	if err := h.db.Create(order).Error; err != nil {
		// 回滚时间槽
		h.db.Model(&slot).Where("id = ?", slot.ID).Updates(map[string]interface{}{
			"status":     "available",
			"booking_id": nil,
			"version":    slot.Version,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// 创建订单项
	slotDateStr := slot.SlotDate.Format("2006-01-02")
	orderItem := &model.OrderItem{
		OrderID:    order.ID,
		ResourceID: &req.ResourceID,
		TicketCount: 1,
		UnitPrice:  0,
		SlotDate:   &slotDateStr,
		StartTime:  &slot.StartTime,
		EndTime:    &slot.EndTime,
	}

	if err := h.db.Create(orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order": order,
		"role":  role,
		"message": fmt.Sprintf("预订成功，请在 %v 前完成支付", paymentDeadline.Format("2006-01-02 15:04:05")),
	})
}

func (h *Handler) ListBookings(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var orders []model.Order
	if err := h.db.Preload("Items").Where("user_id = ? AND order_type = ?", userID, "space").
		Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": orders})
}

func (h *Handler) GetBooking(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orderID := c.Param("id")

	var order model.Order
	if err := h.db.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": order})
}

func (h *Handler) CancelBooking(c *gin.Context) {
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

	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusConfirmed {
		c.JSON(errors.ErrOrderNotPending.HTTP, gin.H{"error": "cannot cancel order in current status"})
		return
	}

	// 更新订单状态
	now := time.Now()
	order.Status = model.OrderStatusCancelled
	order.CancelledAt = &now
	order.CancelReason = "user cancelled"

	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}

	// 释放时间槽
	if len(order.Items) > 0 && order.Items[0].ResourceID != nil {
		h.db.Model(&model.TimeSlot{}).Where("resource_id = ? AND slot_date = ? AND start_time = ?",
			*order.Items[0].ResourceID, *order.Items[0].SlotDate, *order.Items[0].StartTime).
			Updates(map[string]interface{}{
				"status":     "available",
				"booking_id": nil,
			})
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}
