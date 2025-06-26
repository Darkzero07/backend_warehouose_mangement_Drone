package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type TransactionBorrowController struct {
	DB *gorm.DB
}

func NewTransactionBorrowController(db *gorm.DB) *TransactionBorrowController {
	return &TransactionBorrowController{DB: db}
}

// User & Admin: เบิกอุปกรณ์
func (ctrl *TransactionBorrowController) BorrowItem(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input struct {
		ItemID     uint   `json:"item_id" binding:"required"`
		ProjectID  uint   `json:"project_id" binding:"required"`
		Quantity   int    `json:"quantity" binding:"required,min=1"`
		BorrowDate string `json:"borrow_date" binding:"required"`
		DueDate    string `json:"due_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := ctrl.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item models.Item
	if err := tx.Preload("Category").First(&item, input.ItemID).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if item.Quantity < input.Quantity {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not enough %s in stock. Available: %d", item.Name, item.Quantity)})
		return
	}

	// Update item quantity
	item.Quantity -= input.Quantity
	if err := tx.Save(&item).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	// Create borrow transaction record
	transaction := models.TransactionBorrow{
		UserID:         userID,
		ItemID:         input.ItemID,
		ProjectID:      input.ProjectID,
		BorrowQuantity: input.Quantity,
		BorrowDate:     input.BorrowDate,
		DueDate:        input.DueDate,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record borrow transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"message": "Item borrowed successfully", "transaction": transaction})
}

// Admin: ดูข้อมูลการเบิกทั้งหมด
func (ctrl *TransactionBorrowController) GetAllBorrowTransactions(c *gin.Context) {
	var transactions []models.TransactionBorrow
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch borrow transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// Admin: ดูข้อมูลการเบิกของ user ในแต่ละ project
func (ctrl *TransactionBorrowController) GetBorrowTransactionsByProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var transactions []models.TransactionBorrow
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").
		Where("project_id = ?", projectID).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch borrow transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}
