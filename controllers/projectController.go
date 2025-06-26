package controllers

import (
	_ "math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-store/models"
	"warehouse-store/utils"
)

type ProjectController struct {
	DB *gorm.DB
}

func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{DB: db}
}

func (ctrl *ProjectController) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctrl.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (ctrl *ProjectController) GetProjects(c *gin.Context) {
	var projects []models.Project
	if err := ctrl.DB.Find(&projects).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (ctrl *ProjectController) GetProjectByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var project models.Project
	if err := ctrl.DB.First(&project, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (ctrl *ProjectController) UpdateProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var project models.Project
	if err := ctrl.DB.First(&project, id).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input models.Project
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.Name = input.Name
	project.Description = input.Description
	project.StartDate = input.StartDate
	project.EndDate = input.EndDate
	project.Number_of_Drone = input.Number_of_Drone
	project.Location = input.Location

	if err := ctrl.DB.Save(&project).Error; err != nil {
		utils.LogError("Failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (ctrl *ProjectController) DeleteProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := ctrl.DB.Delete(&models.Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (ctrl *ProjectController) GetProjectsByMonth(c *gin.Context) {
	month := c.Param("month") // Expecting format like "06" for June
	year := c.Param("year")   // Expecting format like "2025"

	var projects []models.Project
	if err := ctrl.DB.Where("SUBSTRING(start_date, 4, 2) = ? AND SUBSTRING(start_date, 7, 4) = ?", month, year).
		Find(&projects).Error; err != nil {
		utils.LogError("Failed to fetch projects by month", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	c.JSON(http.StatusOK, projects)
}
