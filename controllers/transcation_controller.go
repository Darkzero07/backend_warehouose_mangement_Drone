package controllers

import (
	"net/http"
	"strconv"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type TransactionController struct {
	DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{DB: db}
}

// User & Admin: เบิกอุปกรณ์
func (ctrl *TransactionController) BorrowItem(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input struct {
		ItemID    uint `json:"item_id" binding:"required"`
		ProjectID uint `json:"project_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
		BorrowDate  string  `json:"borrow_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := ctrl.DB.Begin() // Start a transaction for atomicity
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item models.Item
	// Preload Category when fetching the item to get its name
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

	// Create transaction record
	transaction := models.Transaction{
		UserID:    userID,
		ItemID:    input.ItemID,
		ProjectID: input.ProjectID,
		Quantity:  input.Quantity,
		BorrowDate: input.BorrowDate,
		Type:      "borrow",
		Status:    "Approved",
		// Category field is removed from Transaction model. It's accessed via Item.Category.Name
		// Category:  item.Category
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

// User & Admin: คืนอุปกรณ์
func (ctrl *TransactionController) ReturnItem(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input struct {
		ItemID    uint `json:"item_id" binding:"required"`
		ProjectID uint `json:"project_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
		ReturnDate  string  `json:"return_date" binding:"required"`
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
	// Preload Category when fetching the item to get its name
	if err := tx.Preload("Category").First(&item, input.ItemID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	item.Quantity += input.Quantity
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	transaction := models.Transaction{
		UserID:    userID,
		ItemID:    input.ItemID,
		ProjectID: input.ProjectID,
		Quantity:  input.Quantity,
		ReturnDate:  input.ReturnDate,
		Type:      "return",
		Status:    "Approved",
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record return transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"message": "Item returned successfully", "transaction": transaction})
}

// Admin: ดูข้อมูลการเบิก-คืน ของ user ในแต่ละ project
func (ctrl *TransactionController) GetTransactionsByProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var transactions []models.Transaction
	// New: Preload Item.Category
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").
		Where("project_id = ?", projectID).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// Admin: ดูข้อมูลการเบิก-คืนทั้งหมด
func (ctrl *TransactionController) GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	// New: Preload Item.Category
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// Admin: ดูสรุปยอดของที่เหลือในคลัง
func (ctrl *TransactionController) GetInventorySummary(c *gin.Context) {
	var items []models.Item
	// Preload Category to get its Name for summary
	if err := ctrl.DB.Preload("Category").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory summary"})
		return
	}
	c.JSON(http.StatusOK, items)
}