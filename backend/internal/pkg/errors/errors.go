package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

func (e *AppError) WithDetail(detail string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Detail:  detail,
		HTTP:    e.HTTP,
	}
}

// 预定义错误
var (
	ErrNotFound       = New("NOT_FOUND", "资源不存在", http.StatusNotFound)
	ErrUnauthorized   = New("UNAUTHORIZED", "未授权", http.StatusUnauthorized)
	ErrForbidden      = New("FORBIDDEN", "无权限", http.StatusForbidden)
	ErrConflict       = New("CONFLICT", "资源冲突", http.StatusConflict)
	ErrBadRequest     = New("BAD_REQUEST", "请求错误", http.StatusBadRequest)
	ErrInternalServer = New("INTERNAL_SERVER", "服务器内部错误", http.StatusInternalServerError)

	// 业务错误
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "用户名或密码错误", http.StatusUnauthorized)
	ErrUserExists         = New("USER_EXISTS", "用户已存在", http.StatusConflict)
	ErrSlotUnavailable    = New("SLOT_UNAVAILABLE", "时间段已被预订", http.StatusConflict)
	ErrSlotConflict       = New("SLOT_CONFLICT", "时间段冲突", http.StatusConflict)
	ErrOrderNotPending    = New("ORDER_NOT_PENDING", "订单状态不允许此操作", http.StatusConflict)
	ErrOrderTimeout       = New("ORDER_TIMEOUT", "订单已超时", http.StatusConflict)
	ErrStockNotEnough     = New("STOCK_NOT_ENOUGH", "库存不足", http.StatusConflict)
	ErrAlreadyPurchased   = New("ALREADY_PURCHASED", "已购买过此活动门票", http.StatusConflict)
	ErrRateLimited        = New("RATE_LIMITED", "请求过于频繁", http.StatusTooManyRequests)
)

func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

func GetHTTPStatus(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.HTTP
	}
	return http.StatusInternalServerError
}
