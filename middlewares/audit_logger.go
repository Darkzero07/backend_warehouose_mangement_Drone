package middlewares

import (
	"bytes"
	"io"
	_ "time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
)

func AuditLogger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		// start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		var userID uint
		if id, exists := c.Get("userID"); exists {
			userID = id.(uint)
		}

		log := models.AuditLog{
			UserID:    userID,
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			TableName: "users",
			IPAddress: c.ClientIP(),
			NewValue:  string(bodyBytes),
		}

		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/register" {
			log.Action = c.Request.URL.Path 
			if c.Writer.Status() >= 400 {
				log.Action = log.Action + "_failed"
			}
		}

		db.Create(&log)
	}
}
