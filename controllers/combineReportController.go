package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type CombinedController struct {
	DB *gorm.DB
}

func NewCombinedController(DB *gorm.DB) CombinedController {
	return CombinedController{DB}
}

func (cc *CombinedController) GetFullCombinedData(ctx *gin.Context) {
	projectID := ctx.Query("project_id")
	itemID := ctx.Query("item_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	transactionType := ctx.Query("type")

	projectQuery := cc.DB.Model(&models.Project{})
	if projectID != "" {
		projectQuery = projectQuery.Where("id = ?", projectID)
	}

	var projects []models.Project
	if err := projectQuery.Find(&projects).Error; err != nil {
		utils.LogError("Failed", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var response []map[string]interface{}

	for _, project := range projects {
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
				utils.LogError("Failed", err)
				borrowQuery = borrowQuery.Where("borrow_date <= ?", parsedDate)
			}
		}

		var borrows []models.TransactionBorrow
		if transactionType == "" || transactionType == "borrow" {
			if err := borrowQuery.Find(&borrows).Error; err != nil {
				utils.LogError("Failed", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

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
				utils.LogError("Failed", err)
				returnQuery = returnQuery.Where("return_date >= ?", parsedDate)
			}
		}
		if endDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", endDate); err == nil {
				utils.LogError("Failed", err)
				returnQuery = returnQuery.Where("return_date <= ?", parsedDate)
			}
		}

		var returns []models.TransactionReturn
		if transactionType == "" || transactionType == "return" {
			if err := returnQuery.Find(&returns).Error; err != nil {
				utils.LogError("Failed", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		var damageReports []models.DamageReport
		if err := cc.DB.Preload("Item").Preload("Reporter").
			Where("project_id = ?", project.ID).Find(&damageReports).Error; err != nil {
			utils.LogError("Failed", err)
			continue
		}

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

func (cc *CombinedController) createReturnEntry(project models.Project, returnTx models.TransactionReturn, damageReports []models.DamageReport) map[string]interface{} {
	entry := map[string]interface{}{
		"Project_ID":          project.ID,
		"Project_Name":        project.Name,
		"Project_Description": project.Description,
		"Project_Start_Date":  project.StartDate,
		"Project_End_Date":    project.EndDate,
		"Number_of_Drone":     project.Number_of_Drone,
		"Project_Location":    project.Location,
		"Item_ID":             returnTx.ItemID,
		"Item_Description":    returnTx.Item.Description,
		"Item_Status":         returnTx.Item.Status,
		"Category":            returnTx.Item.Category.Name,
		"Remark":              returnTx.Item.Remark,
		"Transaction_ID":      returnTx.ID,
		"Transaction_Item":    returnTx.Item.Name,
		"Borrow_Quantity":     returnTx.Borrow.BorrowQuantity,
		"Return_Quantity":     returnTx.ReturnQuantity,
		"Borrow_ID":           returnTx.BorrowID,
		"Borrow_Date":         returnTx.Borrow.BorrowDate,
		"Return_Date":         returnTx.ReturnDate,
		"Broken_Drone":        0,
		"User_ID":             returnTx.UserID,
		"Reporter":            returnTx.User.Username,
		"Created_At":          returnTx.CreatedAt,
	}

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
