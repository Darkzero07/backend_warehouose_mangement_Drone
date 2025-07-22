package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"warehouse-store/models"
)

type WarrantyController struct {
	DB *gorm.DB
}

// NewWarrantyController creates a new warranty controller
func NewWarrantyController(db *gorm.DB) *WarrantyController {
	return &WarrantyController{DB: db}
}

// UploadXLSXResponse represents the response structure for XLSX upload
type UploadXLSXResponse struct {
	Message       string                   `json:"message"`
	TotalRecords  int                      `json:"total_records"`
	SuccessCount  int                      `json:"success_count"`
	FailureCount  int                      `json:"failure_count"`
	Errors        []string                 `json:"errors,omitempty"`
	FailedRecords []map[string]interface{} `json:"failed_records,omitempty"`
}

// UploadXLSX handles XLSX file upload and processes warranty records using excelize
func (wc *WarrantyController) UploadXLSX(c *gin.Context) {
	// Get the uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get file from request",
		})
		return
	}
	defer file.Close()

	// Check file extension
	if header.Header.Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Please upload an XLSX file",
		})
		return
	}

	// Open the XLSX file
	f, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read XLSX file: " + err.Error(),
		})
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	// Get the first sheet name
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No sheets found in XLSX file",
		})
		return
	}

	// Get all rows from the first sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get rows from XLSX file: " + err.Error(),
		})
		return
	}

	if len(rows) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No data rows found in XLSX file (only header or empty file)",
		})
		return
	}

	var warranties []models.Warranty
	var errors []string
	var failedRecords []map[string]interface{}
	successCount := 0
	failureCount := 0

	// Process each row (skip header row)
	for rowIndex, row := range rows {
		if rowIndex == 0 {
			continue // Skip header row
		}

		// Ensure we have at least 8 columns (based on your Excel structure)
		if len(row) < 8 {
			// Pad the row with empty strings if needed
			for len(row) < 8 {
				row = append(row, "")
			}
		}

		// Extract data from cells (adjust indices based on your Excel columns)
		droneIDStr := row[0]   // Column A
		serialNumber := row[1] // Column B
		buyDate := row[2]      // Column C
		timeWarranty := row[3] // Column D
		status := row[4]       // Column E
		boxIDStr := row[5]     // Column F (BOX_XX format)
		lot := row[6]          // Column G
		remark := row[7]       // Column H

		// Skip empty rows
		if droneIDStr == "" && serialNumber == "" && boxIDStr == "" {
			continue
		}

		// Validate and convert DroneID
		droneID, err := strconv.ParseUint(droneIDStr, 10, 32)
		if err != nil || droneIDStr == "" {
			failedRecord := map[string]interface{}{
				"row":           rowIndex + 1,
				"drone_id":      droneIDStr,
				"serial_number": serialNumber,
				"box_id":        boxIDStr,
				"error":         "Invalid or empty drone_id format",
			}
			failedRecords = append(failedRecords, failedRecord)
			errors = append(errors, fmt.Sprintf("Row %d: Invalid or empty drone_id format", rowIndex+1))
			failureCount++
			continue
		}

		// Validate serial number
		if serialNumber == "" {
			failedRecord := map[string]interface{}{
				"row":           rowIndex + 1,
				"drone_id":      droneIDStr,
				"serial_number": serialNumber,
				"box_id":        boxIDStr,
				"error":         "Serial number is required",
			}
			failedRecords = append(failedRecords, failedRecord)
			errors = append(errors, fmt.Sprintf("Row %d: Serial number is required", rowIndex+1))
			failureCount++
			continue
		}

		// Extract numeric part from box_id (BOX_XX format)
		var boxID uint = 0
		if strings.HasPrefix(boxIDStr, "BOX_") {
			boxNumStr := strings.TrimPrefix(boxIDStr, "BOX_")
			boxNum, err := strconv.ParseUint(boxNumStr, 10, 32)
			if err == nil {
				boxID = uint(boxNum)
			}
		}

		// If boxID is still 0, it means the format was invalid
		if boxID == 0 && boxIDStr != "" {
			failedRecord := map[string]interface{}{
				"row":           rowIndex + 1,
				"drone_id":      droneIDStr,
				"serial_number": serialNumber,
				"box_id":        boxIDStr,
				"error":         "Invalid box_id format - expected BOX_XX where XX is a number",
			}
			failedRecords = append(failedRecords, failedRecord)
			errors = append(errors, fmt.Sprintf("Row %d: Invalid box_id format - expected BOX_XX where XX is a number", rowIndex+1))
			failureCount++
			continue
		}

		// Set default values if empty
		if buyDate == "" {
			buyDate = time.Now().Format("2006-01-02")
		}
		if timeWarranty == "" {
			timeWarranty = "12 months"
		}
		if status == "" {
			status = "active"
		}

		// Check if warranty with this serial number already exists
        var existingWarranty models.Warranty
        if err := wc.DB.Where("serial_number = ?", serialNumber).First(&existingWarranty).Error; err == nil {
            failedRecord := map[string]interface{}{
                "row":           rowIndex + 1,
                "drone_id":      droneIDStr,
                "serial_number": serialNumber,
                "box_id":        boxIDStr,
                "error":         "Warranty with this serial number already exists",
            }
            failedRecords = append(failedRecords, failedRecord)
            errors = append(errors, fmt.Sprintf("Row %d: Warranty with serial number %s already exists", rowIndex+1, serialNumber))
            failureCount++
            continue
        }

        // Create warranty record
        warranty := models.Warranty{
            DroneID:      uint(droneID),
            SerialNumber: serialNumber,
            BuyDate:      buyDate,
            TimeWarranty: timeWarranty,
            Status:       status,
            BoxID:        boxID,
            Lot:          lot,
            Remark:       remark,
        }

        warranties = append(warranties, warranty)
        successCount++
    }

	// Batch insert warranties
	if len(warranties) > 0 {
		if err := wc.DB.Create(&warranties).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to insert warranty records: " + err.Error(),
			})
			return
		}
	}

	// Prepare response
	response := UploadXLSXResponse{
		Message:       "XLSX file processed successfully",
		TotalRecords:  len(rows) - 1, // Exclude header row
		SuccessCount:  successCount,
		FailureCount:  failureCount,
		Errors:        errors,
		FailedRecords: failedRecords,
	}

	c.JSON(http.StatusOK, response)
}

// GetAllWarranties retrieves all warranty records with pagination
func (wc *WarrantyController) GetAllWarranties(c *gin.Context) {
	var warranties []models.Warranty
	var total int64

	// Get page and limit from query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageNum, _ := strconv.Atoi(page)
	limitNum, _ := strconv.Atoi(limit)

	if pageNum < 1 {
		pageNum = 1
	}
	if limitNum < 1 {
		limitNum = 10
	}

	offset := (pageNum - 1) * limitNum

	// Count total records
	wc.DB.Model(&models.Warranty{}).Count(&total)

	// Get warranties with pagination
	if err := wc.DB.Offset(offset).Limit(limitNum).Find(&warranties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve warranties",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warranties,
		"pagination": gin.H{
			"page":        pageNum,
			"limit":       limitNum,
			"total":       total,
			"total_pages": (total + int64(limitNum) - 1) / int64(limitNum),
		},
	})
}

// GetWarrantyByID retrieves a specific warranty by ID
func (wc *WarrantyController) GetWarrantyByID(c *gin.Context) {
	id := c.Param("id")
	var warranty models.Warranty

	if err := wc.DB.First(&warranty, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Warranty not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve warranty",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warranty,
	})
}

// CreateWarranty creates a new warranty record
func (wc *WarrantyController) CreateWarranty(c *gin.Context) {
	var warranty models.Warranty

	if err := c.ShouldBindJSON(&warranty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if warranty with this serial number already exists
	var existingWarranty models.Warranty
	if err := wc.DB.Where("serial_number = ?", warranty.SerialNumber).First(&existingWarranty).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Warranty with this serial number already exists",
		})
		return
	}

	if err := wc.DB.Create(&warranty).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create warranty",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Warranty created successfully",
		"data":    warranty,
	})
}

// UpdateWarranty updates an existing warranty record
func (wc *WarrantyController) UpdateWarranty(c *gin.Context) {
	id := c.Param("id")
	var warranty models.Warranty

	// Check if warranty exists
	if err := wc.DB.First(&warranty, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Warranty not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve warranty",
		})
		return
	}

	// Bind updated data
	var updatedWarranty models.Warranty
	if err := c.ShouldBindJSON(&updatedWarranty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if serial number is being changed and if it already exists
	if updatedWarranty.SerialNumber != warranty.SerialNumber {
		var existingWarranty models.Warranty
		if err := wc.DB.Where("serial_number = ? AND id != ?", updatedWarranty.SerialNumber, id).First(&existingWarranty).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Warranty with this serial number already exists",
			})
			return
		}
	}

	// Update the warranty
	if err := wc.DB.Model(&warranty).Updates(updatedWarranty).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update warranty",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warranty updated successfully",
		"data":    warranty,
	})
}

// DeleteWarranty deletes a warranty record
func (wc *WarrantyController) DeleteWarranty(c *gin.Context) {
	id := c.Param("id")
	var warranty models.Warranty

	// Check if warranty exists
	if err := wc.DB.First(&warranty, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Warranty not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve warranty",
		})
		return
	}

	// Delete the warranty
	if err := wc.DB.Delete(&warranty).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete warranty",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warranty deleted successfully",
	})
}

// SearchWarranties searches warranties by serial number or status
func (wc *WarrantyController) SearchWarranties(c *gin.Context) {
	var warranties []models.Warranty
	query := c.Query("q")
	status := c.Query("status")

	// Get page and limit from query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageNum, _ := strconv.Atoi(page)
	limitNum, _ := strconv.Atoi(limit)

	if pageNum < 1 {
		pageNum = 1
	}
	if limitNum < 1 {
		limitNum = 10
	}

	offset := (pageNum - 1) * limitNum

	// Build the database query
	dbQuery := wc.DB.Model(&models.Warranty{})

	if query != "" {
		dbQuery = dbQuery.Where("serial_number LIKE ?", "%"+query+"%")
	}

	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	// Count total records
	var total int64
	dbQuery.Count(&total)

	// Get warranties with pagination
	if err := dbQuery.Offset(offset).Limit(limitNum).Find(&warranties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search warranties",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warranties,
		"pagination": gin.H{
			"page":        pageNum,
			"limit":       limitNum,
			"total":       total,
			"total_pages": (total + int64(limitNum) - 1) / int64(limitNum),
		},
	})
}
