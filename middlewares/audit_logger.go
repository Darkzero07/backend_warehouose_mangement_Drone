package middlewares

import (
	"bytes"
	_"encoding/json"
	"io"
	_"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
)

func AuditLogger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for certain routes
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/register" {
			c.Next()
			return
		}

		// Get user ID if authenticated
		userID, exists := c.Get("userID")
		if !exists {
			userID = uint(0) // Anonymous user
		}

		// Get request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process request
		c.Next()

		// Log after request is processed
		if c.Writer.Status() < 400 { // Only log successful actions
			action := c.Request.Method + " " + c.Request.URL.Path
			recordID := uint(0) // You'll need to extract this from the response or request

			log := models.AuditLog{
				UserID:    userID.(uint),
				Action:    action,
				TableName: "", // You'll need to determine this based on the route
				RecordID:  recordID,
				OldValue:  "", // You can capture this for PUT/PATCH requests
				NewValue:  string(bodyBytes),
				IPAddress: c.ClientIP(),
			}

			db.Create(&log)
		}
	}
}