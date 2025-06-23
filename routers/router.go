package routers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/controllers" // Make sure this path is correct based on your project structure
	"warehouse-store/middlewares" // Make sure this path is correct based on your project structure
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS (if frontend and backend are on different origins)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend origin in production
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	projectController := controllers.NewProjectController(db)
	itemController := controllers.NewItemController(db)
	transactionController := controllers.NewTransactionController(db)
	damageReportController := controllers.NewDamageReportController(db)
	categoryController := controllers.NewCategoryController(db)
	passwordResetController := controllers.NewPasswordResetController(db)

	// Public routes
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)
	r.POST("/request-password-reset", passwordResetController.RequestReset)
	r.POST("/reset-password", passwordResetController.ResetPassword)

	// Authenticated routes
	authorized := r.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		// User & Admin Routes for their own data or common operations
		authorized.GET("/items", itemController.GetItems)
		authorized.GET("/items/:id", itemController.GetItemByID)
		authorized.GET("/projects", projectController.GetProjects)
		authorized.GET("/projects/:id", projectController.GetProjectByID)
		authorized.GET("/projects/filter-month/:year/:month", projectController.GetProjectsByMonth)

		// New Category routes (accessible to all authenticated users for viewing)
		authorized.GET("/categories", categoryController.GetCategories)
		authorized.GET("/categories/:id", categoryController.GetCategoryByID)

		// Borrow/Return by any authenticated user
		authorized.POST("/transactions/borrow", transactionController.BorrowItem)
		authorized.POST("/transactions/return", transactionController.ReturnItem)

		// Report damage by any authenticated user
		authorized.POST("/damage-reports", damageReportController.CreateDamageReport)
		authorized.GET("/damage-reports", damageReportController.GetDamageReports)
		authorized.PUT("/damage-reports/:id/status", damageReportController.UpdateDamageReportStatus)

		// Add refresh token route
		authorized.POST("/refresh-token", authController.RefreshToken)

		// Admin routes
		admin := authorized.Group("/")
		admin.Use(middlewares.AuthorizeRole("admin"))
		{
			// User Management
			admin.GET("/users", userController.GetUsers)
			admin.GET("/users/:id", userController.GetUserByID)
			admin.PUT("/users/:id", userController.UpdateUser)
			admin.DELETE("/users/:id", userController.DeleteUser)

			// Project Management
			admin.POST("/projects", projectController.CreateProject)
			admin.PUT("/projects/:id", projectController.UpdateProject)
			admin.DELETE("/projects/:id", projectController.DeleteProject)

			// Item Management
			admin.POST("/items", itemController.CreateItem)
			admin.PUT("/items/:id", itemController.UpdateItem)
			admin.DELETE("/items/:id", itemController.DeleteItem)

			// New: Category Management (Admin only)
			admin.POST("/categories", categoryController.CreateCategory)
			admin.PUT("/categories/:id", categoryController.UpdateCategory)
			admin.DELETE("/categories/:id", categoryController.DeleteCategory)

			// Transaction Reporting (Admin specific)
			admin.GET("/admin/transactions", transactionController.GetAllTransactions)
			admin.GET("/admin/transactions/project/:projectId", transactionController.GetTransactionsByProject)
			admin.GET("/admin/inventory-summary", transactionController.GetInventorySummary)

			// Damage Report Management (Admin specific)
			// admin.GET("/admin/damage-reports", damageReportController.GetDamageReports)
			// admin.PUT("/admin/damage-reports/:id/status", damageReportController.UpdateDamageReportStatus)
		}
	}

	return r
}
