package controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type ItemController struct {
	DB *gorm.DB
}

func NewItemController(db *gorm.DB) *ItemController {
	return &ItemController{DB: db}
}

func (ctrl *ItemController) CreateItem(c *gin.Context) {
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := ctrl.DB.First(&category, item.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Category ID provided"})
		return
	}

	if err := ctrl.DB.Create(&item).Error; err != nil {
		utils.LogError("Failed to create item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}
	utils.LogInfo("Item created successfully", zap.Any("item", item))
	c.JSON(http.StatusCreated, item)

	ctrl.DB.Preload("Category").First(&item, item.ID)
	c.JSON(http.StatusCreated, item)
}

func (ctrl *ItemController) GetItemByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.Item
	if err := ctrl.DB.Preload("Category").First(&item, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (ctrl *ItemController) UpdateItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.Item
	if err := ctrl.DB.First(&item, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var input struct { 
		Name        string `json:"Name"`
		Description string `json:"Description"`
		Quantity    int    `json:"Quantity"`
		Status      string `json:"Status"`
		CategoryID  uint   `json:"CategoryID"` 
		Remark      string `json:"Remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := ctrl.DB.First(&category, input.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Category ID provided"})
		return
	}

	item.Name = input.Name
	item.Description = input.Description
	item.Quantity = input.Quantity
	item.Status = input.Status
	item.CategoryID = input.CategoryID 
	item.Remark = input.Remark

	if err := ctrl.DB.Save(&item).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}
	
	ctrl.DB.Preload("Category").First(&item, item.ID)
	c.JSON(http.StatusOK, item)
}

func (ctrl *ItemController) DeleteItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := ctrl.DB.Delete(&models.Item{}, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

func (ctrl *ItemController) GetItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	status := c.Query("status")
	categoryID := c.Query("category_id")

	offset := (page - 1) * limit

	query := ctrl.DB.Preload("Category")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	var items []models.Item
	if err := query.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	var total int64
	query.Model(&models.Item{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}
