package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type PasswordResetController struct {
	DB *gorm.DB
}

func NewPasswordResetController(db *gorm.DB) *PasswordResetController {
	return &PasswordResetController{DB: db}
}

func (ctrl *PasswordResetController) RequestReset(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("username = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	user.ResetToken = token
	user.ResetTokenExpiry = time.Now().Add(1 * time.Hour)
	if err := ctrl.DB.Save(&user).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent", "token": token})
}

func (ctrl *PasswordResetController) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ParseToken(input.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	var user models.User
	if err := ctrl.DB.First(&user, claims.UserID).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if user.ResetToken != input.Token || time.Now().After(user.ResetTokenExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = hashedPassword
	user.ResetToken = ""
	if err := ctrl.DB.Save(&user).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
