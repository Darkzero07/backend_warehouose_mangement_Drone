package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type DamageReportController struct {
	DB *gorm.DB
}

func NewDamageReportController(db *gorm.DB) *DamageReportController {
	return &DamageReportController{DB: db}
}

// User & Admin: สร้างรายงานอุปกรณ์ชำรุด
func (ctrl *DamageReportController) CreateDamageReport(c *gin.Context) {
	reporterID := c.MustGet("userID").(uint)

	var report models.DamageReport
	if err := c.ShouldBindJSON(&report); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if report.ProjectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	report.ReporterID = reporterID
	report.Status = "Pending" // Default status

	tx := ctrl.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item models.Item
	// Preload Category when fetching the item to ensure it's available if needed (though not stored directly in report)
	if err := tx.Preload("Category").First(&item, report.ItemID).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Category field is removed from DamageReport model. It's accessed via Item.Category.Name
	// report.Category = item.Category

	var project models.Project
	if err := tx.First(&project, report.ProjectID).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if err := tx.Create(&report).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create damage report"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusCreated, report)
}

// Admin: ดูรายงานอุปกรณ์ชำรุดทั้งหมด
func (ctrl *DamageReportController) GetDamageReports(c *gin.Context) {
	var reports []models.DamageReport
	// New: Preload Item.Category
	if err := ctrl.DB.Preload("Item.Category").Preload("Reporter").Preload("Project").Find(&reports).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch damage reports"})
		return
	}
	c.JSON(http.StatusOK, reports)
}

// Admin: อัปเดตสถานะรายงานอุปกรณ์ชำรุด
func (ctrl *DamageReportController) UpdateDamageReportStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var report models.DamageReport
	if err := ctrl.DB.First(&report, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Damage report not found"})
		return
	}

	var input struct {
        Status      string `json:"status"`
        Broken_Drone int   `json:"broken_drone"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        utils.LogError("Failed", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update fields only if they're provided in the input
    if input.Status != "" {
        report.Status = input.Status
    }
    if input.Broken_Drone != 0 { 
        report.Broken_Drone = uint(input.Broken_Drone)
    }

    if err := ctrl.DB.Save(&report).Error; err != nil {
        utils.LogError("Failed", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update damage report"})
        return
    }
    c.JSON(http.StatusOK, report)
}