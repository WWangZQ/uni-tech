package user

import (
	"net/http"

	"smart-campus/internal/model"
	"smart-campus/internal/pkg/errors"
	"smart-campus/internal/pkg/jwt"
	"smart-campus/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	db         *gorm.DB
	jwtManager *jwt.JWTManager
}

func NewHandler(db *gorm.DB, jwtManager *jwt.JWTManager) *Handler {
	return &Handler{
		db:         db,
		jwtManager: jwtManager,
	}
}

type RegisterRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterResponse struct {
	User  *model.User `json:"user"`
	Token string     `json:"token"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已存在
	var existing model.User
	if err := h.db.Where("student_id = ? OR email = ?", req.StudentID, req.Email).First(&existing).Error; err == nil {
		c.JSON(errors.ErrUserExists.HTTP, gin.H{"error": errors.ErrUserExists.Message})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 创建用户
	user := &model.User{
		StudentID:    req.StudentID,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Role:        "undergraduate",
		CreditScore: 100,
		QuotaHours:  10,
		Status:      "active",
	}

	if req.Phone != "" {
		user.Phone = req.Phone
	}

	user.PasswordHash = string(hashedPassword)

	if err := h.db.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// 生成 JWT
	token, err := h.jwtManager.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

type LoginRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  *model.User `json:"user"`
	Token string     `json:"token"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.Where("student_id = ? AND status = ?", req.StudentID, "active").First(&user).Error; err != nil {
		c.JSON(errors.ErrInvalidCredentials.HTTP, gin.H{"error": errors.ErrInvalidCredentials.Message})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(errors.ErrInvalidCredentials.HTTP, gin.H{"error": errors.ErrInvalidCredentials.Message})
		return
	}

	// 生成 JWT
	token, err := h.jwtManager.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type CreditScoreResponse struct {
	CreditScore int `json:"credit_score"`
}

func (h *Handler) GetCreditScore(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user model.User
	if err := h.db.Select("credit_score").First(&user, userID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credit_score": user.CreditScore})
}

type QuotaResponse struct {
	QuotaHours int    `json:"quota_hours"`
	Role       string `json:"role"`
}

func (h *Handler) GetQuota(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user model.User
	if err := h.db.Select("quota_hours, role").First(&user, userID).Error; err != nil {
		c.JSON(errors.ErrNotFound.HTTP, gin.H{"error": errors.ErrNotFound.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quota_hours": user.QuotaHours,
		"role":        user.Role,
	})
}
