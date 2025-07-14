package controllers

import (
	"net/http"
	"strconv"

	"warehouse-store/models"
	"warehouse-store/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionReturnController struct {
	DB *gorm.DB
}

func NewTransactionReturnController(db *gorm.DB) *TransactionReturnController {
	return &TransactionReturnController{DB: db}
}

type ReturnInput struct {
	BorrowIDStr   string   `json:"borrow_id" binding:"required"`
	QuantityStr   string    `json:"quantity" binding:"required,min=1"`
	ReturnDate string `json:"return_date" binding:"required"`
	BorrowID   uint   `json:"-"`
	Quantity   int    `json:"-"`

}

func (ctrl *TransactionReturnController) ReturnItem(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input  ReturnInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	borrowID, err := strconv.ParseUint(input.BorrowIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Borrow ID"})
		return
	}
	input.BorrowID = uint(borrowID)

	quantity, err := strconv.ParseUint(input.QuantityStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Quantity"})
		return
	}
	input.Quantity = int(quantity)

	tx := ctrl.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var borrow models.TransactionBorrow
	if err := tx.Preload("Item").First(&borrow, input.BorrowID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Original borrow transaction not found"})
		return
	}

	if input.Quantity > borrow.BorrowQuantity {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Return quantity cannot exceed borrowed quantity"})
		return
	}

	borrow.Item.Quantity += input.Quantity
	if err := tx.Save(&borrow.Item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	transaction := models.TransactionReturn{
		UserID:         userID,
		ItemID:         borrow.ItemID,
		ProjectID:      borrow.ProjectID,
		ReturnQuantity: input.Quantity,
		ReturnDate:     input.ReturnDate,
		BorrowID:       borrow.ID,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record return transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"message": "Item returned successfully", "transaction": transaction})
}

func (ctrl *TransactionReturnController) GetAllReturnTransactions(c *gin.Context) {
	var transactions []models.TransactionReturn
	if err := ctrl.DB.Preload("User").Preload("Item.Category").Preload("Project").Preload("Borrow").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch return transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}
