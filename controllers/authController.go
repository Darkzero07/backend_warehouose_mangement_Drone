package controllers

import (
	"net/http"
	_ "time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = hashedPassword

	if user.Role == "" {
		user.Role = "user" // Default role
	}

	if err := ctrl.DB.Create(&user).Error; err != nil {
		utils.LogError("Failed to create User", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Create audit log entry
	auditLog := models.AuditLog{
		UserID:    user.ID, // The newly created user's ID
		Action:    "register",
		TableName: "users",
		RecordID:  user.ID,
		NewValue:  "User registered", // Or you can marshal the user struct
		IPAddress: c.ClientIP(),
	}
	ctrl.DB.Create(&auditLog)

	utils.LogInfo("user created successfully", zap.Any("user", user))
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed to bind json body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		// Log failed login attempt
		auditLog := models.AuditLog{
			Action:    "login_failed",
			TableName: "users",
			NewValue:  "Failed login attempt for username: " + input.Username,
			IPAddress: c.ClientIP(),
		}
		ctrl.DB.Create(&auditLog)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		// Log failed login attempt
		auditLog := models.AuditLog{
			Action:    "login_failed",
			TableName: "users",
			RecordID:  user.ID,
			NewValue:  "Failed login attempt for user ID: " + string(user.ID),
			IPAddress: c.ClientIP(),
		}
		ctrl.DB.Create(&auditLog)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.LogError("Failed to generate token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.LogError("Failed to generate refresh token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Log successful login
	auditLog := models.AuditLog{
		UserID:    user.ID,
		Action:    "login",
		TableName: "users",
		RecordID:  user.ID,
		NewValue:  "User logged in",
		IPAddress: c.ClientIP(),
	}
	ctrl.DB.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

// Add new RefreshToken endpoint
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ParseToken(input.RefreshToken)
	if err != nil {
		utils.LogError("Failed to parse", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user models.User
	if err := ctrl.DB.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Verify the role in the refresh token matches the user's current role
	// if claims.Role != user.Role {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "Role mismatch"})
	//     return
	// }

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.LogError("Failed to generate Token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
