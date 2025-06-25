package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	_"warehouse-store/models"
	"warehouse-store/utils"
)

type CombinedReportController struct {
	DB *gorm.DB
}

func NewCombinedReportController(db *gorm.DB) *CombinedReportController {
	return &CombinedReportController{DB: db}
}

// GetCombinedReports returns a combined view of transactions and damage reports
func (ctrl *CombinedReportController) GetCombinedReports(c *gin.Context) {
	var results []struct {
		// Item fields
		ItemName        string `json:"item_name"`
		ItemDescription string `json:"item_description"`
		ItemQuantity    int    `json:"item_quantity"`
		ItemStatus      string `json:"item_status"`
		Category        string `json:"category"`
		Remark          string `json:"remark"`
		
		// Project fields
		Project           string `json:"project"`
		ProjectStartDate  string `json:"project_start_date"`
		ProjectEndDate    string `json:"project_end_date"`
		NumberOfDrone     int    `json:"number_of_drone"`
		
		// Transaction fields
		BorrowQuantity   int    `json:"borrow_quantity"`
		BorrowDate       string `json:"borrow_date"`
		ReturnQuantity   int    `json:"return_quantity"`
		ReturnDate       string `json:"return_date"`
		
		// Damage report fields
		BrokenDrone            int    `json:"broken_drone"`
		DamageReportDescription string `json:"damage_report_description"`
		DamageReportStatus     string `json:"damage_report_status"`
		Reporter               string `json:"reporter"`
	}

	// Query to join all the necessary tables
	err := ctrl.DB.Table("transactions").
		Select(`
			items.name as item_name,
			items.description as item_description,
			items.quantity as item_quantity,
			items.status as item_status,
			categories.name as category,
			items.remark as remark,
			projects.name as project,
			projects.start_date as project_start_date,
			projects.end_date as project_end_date,
			projects.number_of_drone as number_of_drone,
			CASE WHEN transactions.type = 'borrow' THEN transactions.quantity ELSE 0 END as borrow_quantity,
			transactions.borrow_date as borrow_date,
			CASE WHEN transactions.type = 'return' THEN transactions.quantity ELSE 0 END as return_quantity,
			transactions.return_date as return_date,
			damage_reports.broken_drone as broken_drone,
			damage_reports.description as damage_report_description,
			damage_reports.status as damage_report_status,
			users.username as reporter
		`).
		Joins("LEFT JOIN items ON transactions.item_id = items.id").
		Joins("LEFT JOIN categories ON items.category_id = categories.id").
		Joins("LEFT JOIN projects ON transactions.project_id = projects.id").
		Joins("LEFT JOIN damage_reports ON damage_reports.item_id = items.id AND damage_reports.project_id = projects.id").
		Joins("LEFT JOIN users ON damage_reports.reporter_id = users.id").
		Scan(&results).Error

	if err != nil {
		utils.LogError("Failed to fetch combined reports", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch combined reports"})
		return
	}

	c.JSON(http.StatusOK, results)
}