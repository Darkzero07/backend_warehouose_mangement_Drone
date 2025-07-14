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

type BorrowInput struct {
	ItemIDStr      string `json:"item_id" binding:"required"`
	ProjectIDStr   string `json:"project_id" binding:"required"`
	BorrowQuantityStr string    `json:"borrow_quantity" binding:"required,min=1"`
	BorrowDate     string `json:"borrow_date" binding:"required"`
	DueDate        string `json:"due_date" binding:"required"`

	// These will be populated after validation
	ItemID    uint `json:"-"`
	ProjectID uint `json:"-"`
	BorrowQuantity int `json:"-"`
}

func (ctrl *TransactionBorrowController) BorrowItem(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input BorrowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string IDs to uint
	itemID, err := strconv.ParseUint(input.ItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}
	input.ItemID = uint(itemID)

	projectID, err := strconv.ParseUint(input.ProjectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	input.ProjectID = uint(projectID)

	borrowQuantity, err := strconv.ParseUint(input.BorrowQuantityStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	input.BorrowQuantity = int(borrowQuantity)

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

	if item.Quantity < input.BorrowQuantity {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not enough %s in stock. Available: %d", item.Name, item.Quantity)})
		return
	}

	item.Quantity -= input.BorrowQuantity
	if err := tx.Save(&item).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	transaction := models.TransactionBorrow{
		UserID:         userID,
		ItemID:         input.ItemID,
		ProjectID:      input.ProjectID,
		BorrowQuantity: input.BorrowQuantity,
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

func (ctrl *TransactionBorrowController) GetAllBorrowTransactions(c *gin.Context) {
	var transactions []models.TransactionBorrow
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch borrow transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

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
