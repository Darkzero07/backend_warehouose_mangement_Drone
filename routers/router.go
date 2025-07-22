package routers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/controllers"
	"warehouse-store/middlewares"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS (if frontend and backend are on different origins)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize controllers
	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	projectController := controllers.NewProjectController(db)
	itemController := controllers.NewItemController(db)
	damageReportController := controllers.NewDamageReportController(db)
	categoryController := controllers.NewCategoryController(db)
	passwordResetController := controllers.NewPasswordResetController(db)
	combinedReportController := controllers.NewCombinedController(db)
	transactionBorrowController := controllers.NewTransactionBorrowController(db)
	transactionReturnController := controllers.NewTransactionReturnController(db)
	warantyController := controllers.NewWarrantyController(db)

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

		// Category routes (accessible to all authenticated users for viewing)
		authorized.GET("/categories", categoryController.GetCategories)
		authorized.GET("/categories/:id", categoryController.GetCategoryByID)

		// New Borrow/Return routes with separate controllers
		authorized.POST("/transactions/borrow", transactionBorrowController.BorrowItem)
		authorized.POST("/transactions/return", transactionReturnController.ReturnItem)
		authorized.GET("/transactions/borrows", transactionBorrowController.GetAllBorrowTransactions)
		authorized.GET("/transactions/returns", transactionReturnController.GetAllReturnTransactions)

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

			// Category Management (Admin only)
			admin.POST("/categories", categoryController.CreateCategory)
			admin.PUT("/categories/:id", categoryController.UpdateCategory)
			admin.DELETE("/categories/:id", categoryController.DeleteCategory)

			// Transaction Reporting (Admin specific)
			admin.GET("/admin/transactions/borrows", transactionBorrowController.GetAllBorrowTransactions)
			admin.GET("/admin/transactions/borrows/project/:projectId", transactionBorrowController.GetBorrowTransactionsByProject)
			admin.GET("/admin/transactions/returns", transactionReturnController.GetAllReturnTransactions)
			admin.GET("/admin/summary-table", combinedReportController.GetFullCombinedData)

			// Damage Report Management (Admin specific)
			// These are now available to all authenticated users above
			// admin.GET("/admin/damage-reports", damageReportController.GetDamageReports)
			admin.PUT("/admin/damage-reports/:id/status", damageReportController.UpdateDamageReportStatus)
			admin.GET("/admin/warranty", warantyController.GetAllWarranties)
			admin.GET("/admin/warranty/:id", warantyController.GetWarrantyByID)
			admin.POST("/admin/warranty", warantyController.CreateWarranty)
			admin.PUT("/admin/warranty/:id", warantyController.UpdateWarranty)
			admin.DELETE("/admin/warranty/:id", warantyController.DeleteWarranty)
			admin.POST("/admin/warranty/upload", warantyController.UploadXLSX)
		}
	}

	return r
}
