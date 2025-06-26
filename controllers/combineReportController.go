package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
)

type CombinedController struct {
	DB *gorm.DB
}

func NewCombinedController(DB *gorm.DB) CombinedController {
	return CombinedController{DB}
}

// GetFullCombinedData retrieves all data in a flattened structure with enhanced filtering
func (cc *CombinedController) GetFullCombinedData(ctx *gin.Context) {
	// Get optional query parameters for filtering
	projectID := ctx.Query("project_id")
	itemID := ctx.Query("item_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	transactionType := ctx.Query("type")

	// Base query for projects
	projectQuery := cc.DB.Model(&models.Project{})
	if projectID != "" {
		projectQuery = projectQuery.Where("id = ?", projectID)
	}

	var projects []models.Project
	if err := projectQuery.Find(&projects).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Prepare the final response
	var response []map[string]interface{}

	for _, project := range projects {
		// Build borrow query
		borrowQuery := cc.DB.Model(&models.TransactionBorrow{}).
			Preload("Item").
			Preload("Item.Category").
			Preload("User").
			Where("project_id = ?", project.ID)

		if itemID != "" {
			borrowQuery = borrowQuery.Where("item_id = ?", itemID)
		}
		if startDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
				borrowQuery = borrowQuery.Where("borrow_date >= ?", parsedDate)
			}
		}
		if endDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", endDate); err == nil {
				borrowQuery = borrowQuery.Where("borrow_date <= ?", parsedDate)
			}
		}

		var borrows []models.TransactionBorrow
		if transactionType == "" || transactionType == "borrow" {
			if err := borrowQuery.Find(&borrows).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		// Build return query
		returnQuery := cc.DB.Model(&models.TransactionReturn{}).
			Preload("Item").
			Preload("Item.Category").
			Preload("User").
			Preload("Borrow").
			Where("project_id = ?", project.ID)

		if itemID != "" {
			returnQuery = returnQuery.Where("item_id = ?", itemID)
		}
		if startDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
				returnQuery = returnQuery.Where("return_date >= ?", parsedDate)
			}
		}
		if endDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", endDate); err == nil {
				returnQuery = returnQuery.Where("return_date <= ?", parsedDate)
			}
		}

		var returns []models.TransactionReturn
		if transactionType == "" || transactionType == "return" {
			if err := returnQuery.Find(&returns).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		// Get damage reports for this project
		var damageReports []models.DamageReport
		if err := cc.DB.Preload("Item").Preload("Reporter").
			Where("project_id = ?", project.ID).Find(&damageReports).Error; err != nil {
			continue
		}

		// Create entries for each borrow transaction
		// for _, b := range borrows {
		// 	entry := cc.createBorrowEntry(project, b, damageReports)
		// 	response = append(response, entry)
		// }

		// Create entries for each return transaction
		for _, r := range returns {
			entry := cc.createReturnEntry(project, r, damageReports)
			response = append(response, entry)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// // Helper function to create borrow entry
// func (cc *CombinedController) createBorrowEntry(project models.Project, borrow models.TransactionBorrow, damageReports []models.DamageReport) map[string]interface{} {
// 	entry := map[string]interface{}{
// 		"Project_ID":           project.ID,
// 		"Project_Name":         project.Name,
// 		"Project_Description":  project.Description,
// 		"Project_Start_Date":   project.StartDate,
// 		"Project_End_Date":     project.EndDate,
// 		"Number_of_Drone":      project.Number_of_Drone,
// 		"Project_Location":     project.Location,
// 		"Item_ID":             borrow.ItemID,
// 		"Item_Description":    borrow.Item.Description,
// 		"Item_Status":        borrow.Item.Status,
// 		"Category":           borrow.Item.Category.Name,
// 		"Remark":             borrow.Item.Remark,
// 		"Transaction_ID":     borrow.ID,
// 		"Transaction_Item":   borrow.Item.Name,
// 		"Transaction_Type":   "borrow",
// 		"Transaction_Quantity": borrow.BorrowQuantity,
// 		"Borrow_Date":        borrow.BorrowDate,
// 		"Due_Date":          borrow.DueDate,
// 		"Broken_Drone":      0,
// 		"Status":           "Approved",
// 		"User_ID":         borrow.UserID,
// 		"Reporter":       borrow.User.Username,
// 		"Created_At":    borrow.CreatedAt,
// 	}

// 	// Add damage report info if exists
// 	for _, dr := range damageReports {
// 		if dr.ItemID == borrow.ItemID {
// 			entry["Broken_Drone"] = dr.Broken_Drone
// 			entry["Damage_Reporter"] = dr.Reporter.Username
// 			entry["Damage_Report_ID"] = dr.ID
// 			break
// 		}
// 	}

// 	return entry
// }

// Helper function to create return entry
func (cc *CombinedController) createReturnEntry(project models.Project, returnTx models.TransactionReturn, damageReports []models.DamageReport) map[string]interface{} {
	entry := map[string]interface{}{
		"Project_ID":           project.ID,
		"Project_Name":         project.Name,
		"Project_Description":  project.Description,
		"Project_Start_Date":   project.StartDate,
		"Project_End_Date":     project.EndDate,
		"Number_of_Drone":      project.Number_of_Drone,
		"Project_Location":     project.Location,
		"Item_ID":             returnTx.ItemID,
		"Item_Description":    returnTx.Item.Description,
		"Item_Status":        returnTx.Item.Status,
		"Category":           returnTx.Item.Category.Name,
		"Remark":             returnTx.Item.Remark,
		"Transaction_ID":     returnTx.ID,
		"Transaction_Item":   returnTx.Item.Name,
		"Borrow_Quantity": returnTx.Borrow.BorrowQuantity,
		"Return_Quantity": returnTx.ReturnQuantity,
		"Borrow_ID":         returnTx.BorrowID,
		"Borrow_Date":      returnTx.Borrow.BorrowDate,
		"Return_Date":     returnTx.ReturnDate,
		"Broken_Drone":   0,
		"User_ID":      returnTx.UserID,
		"Reporter":    returnTx.User.Username,
		"Created_At": returnTx.CreatedAt,
	}

	// Add damage report info if exists
	for _, dr := range damageReports {
		if dr.ItemID == returnTx.ItemID {
			entry["Broken_Drone"] = dr.Broken_Drone
			entry["Damage_Reporter"] = dr.Reporter.Username
			entry["Damage_Report_ID"] = dr.ID
			entry["Damage_Report_Description"] = dr.Description
			entry["Damage_Status"] = dr.Status
			break
		}
	}

	return entry
}