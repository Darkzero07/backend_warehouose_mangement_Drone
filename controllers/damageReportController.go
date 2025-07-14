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
	report.Status = "Pending" 

	tx := ctrl.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item models.Item

	if err := tx.Preload("Category").First(&item, report.ItemID).Error; err != nil {
		utils.LogError("Failed", err)
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

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

func (ctrl *DamageReportController) UpdateDamageReportStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var report models.DamageReport
	if err := ctrl.DB.First(&report, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Damage report not found"})
		return
	}

	var input struct {
		Description  string  `json:"description"`
		Status       string `json:"status"`
		Broken_Drone *uint    `json:"broken_drone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status != "" {
		report.Description = input.Description
		report.Status = input.Status
	}
	if input.Broken_Drone != nil { 
    report.Broken_Drone = *input.Broken_Drone
}

	if err := ctrl.DB.Save(&report).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update damage report"})
		return
	}
	c.JSON(http.StatusOK, report)
}
