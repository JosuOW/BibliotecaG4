package controllers

import (
	"net/http"
	"strconv"

	"biblioteca-backend/database"
	"biblioteca-backend/middleware"
	"biblioteca-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// @Summary		Get current user profile
// @Description	Get profile information of the authenticated user
// @Tags			users
// @Accept			json
// @Produce		json
// @Success		200	{object}	models.User
// @Failure		401	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Security		BearerAuth
// @Router			/users/profile [get]
func GetUserProfile(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	result := database.DB.Preload("Loans").
		Preload("Loans.Exemplar").
		Preload("Loans.Exemplar.Book").
		Preload("Fine").
		Where("login = ?", userLogin).
		First(&user)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// @Summary		Update user profile
// @Description	Update profile information of the authenticated user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			user	body		models.User	true	"User data to update"
// @Success		200		{object}	models.User
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		404		{object}	map[string]string
// @Security		BearerAuth
// @Router			/users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	result := database.DB.Where("login = ?", userLogin).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updateData models.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Solo permitir actualizar ciertos campos
	user.Name = updateData.Name
	user.LastName = updateData.LastName
	user.Email = updateData.Email
	user.Street = updateData.Street
	user.Number = updateData.Number
	user.Floor = updateData.Floor
	user.City = updateData.City
	user.PostalCode = updateData.PostalCode
	user.ParentsPhone = updateData.ParentsPhone
	user.DepartmentName = updateData.DepartmentName

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error updating user profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// @Summary		Get all users (Admin only)
// @Description	Get list of all users in the system
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			page	query		int	false	"Page number"		default(1)
// @Param			limit	query		int	false	"Items per page"	default(10)
// @Success		200		{array}		models.User
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Security		BearerAuth
// @Router			/admin/users [get]
func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var users []models.User
	var total int64

	database.DB.Model(&models.User{}).Count(&total)

	result := database.DB.Preload("Loans").
		Preload("Fine").
		Offset(offset).
		Limit(limit).
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching users",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// @Summary		Create new user (Admin only)
// @Description	Create a new user in the system
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			user	body		models.User	true	"User data"
// @Success		201		{object}	models.User
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Security		BearerAuth
// @Router			/admin/users [post]func CreateUser(c *gin.Context) {
func CreateUser(c *gin.Context) {
	var user models.User

	// Validar JSON entrante
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validación adicional si fuera necesaria (por ejemplo, con validator.v10)
	if err := validate.Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos", "details": err.Error()})
		return
	}

	// Verificar si el usuario con el mismo login ya existe
	var existing models.User
	if err := database.DB.Where("login = ?", user.Login).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ya existe un usuario con este login"})
		return
	}

	// Crear el usuario
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error al crear el usuario",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Usuario creado exitosamente",
		"data":    user,
	})
}
