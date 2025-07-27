package controllers

import (
	"net/http"
	"strconv"
	"time"

	"biblioteca-backend/database"
	"biblioteca-backend/middleware"
	"biblioteca-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		Get my current fine
// @Description	Get current fine for authenticated user
// @Tags			fines
// @Accept			json
// @Produce		json
// @Success		200	{object}	models.Fine
// @Failure		401	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Security		BearerAuth
// @Router			/fines/my [get]
func GetMyFine(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Buscar usuario
	var user models.User
	if err := database.DB.Where("login = ?", userLogin).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Buscar multa activa
	var fine models.Fine
	result := database.DB.Where("user_id = ? AND is_active = true", user.ID).First(&fine)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No active fine found"})
		return
	}

	// Calcular información adicional
	endDate := fine.CalculateEndDate()
	fine.EndDate = &endDate

	response := gin.H{
		"data":           fine,
		"remaining_days": fine.GetRemainingDays(),
		"is_expired":     fine.IsExpired(),
	}

	c.JSON(http.StatusOK, response)
}

// @Summary		Get fine history
// @Description	Get fine history for authenticated user
// @Tags			fines
// @Accept			json
// @Produce		json
// @Param			page	query		int	false	"Page number"		default(1)
// @Param			limit	query		int	false	"Items per page"	default(10)
// @Success		200		{array}		models.FineHistory
// @Failure		401		{object}	map[string]string
// @Security		BearerAuth
// @Router			/fines/history [get]
func GetFineHistory(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Buscar usuario
	var user models.User
	if err := database.DB.Where("login = ?", userLogin).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var history []models.FineHistory
	var total int64

	database.DB.Model(&models.FineHistory{}).Where("user_id = ?", user.ID).Count(&total)

	result := database.DB.Where("user_id = ?", user.ID).
		Order("end_date DESC").
		Offset(offset).
		Limit(limit).
		Find(&history)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching fine history",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// @Summary		Process expired fines
// @Description	Process and close expired fines (System job)
// @Tags			fines
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]int
// @Failure		500	{object}	map[string]string
// @Security		BearerAuth
// @Router			/admin/fines/process-expired [post]
func ProcessExpiredFines(c *gin.Context) {
	var expiredFines []models.Fine
	now := time.Now()

	// Buscar multas expiradas
	result := database.DB.Where("is_active = true").Find(&expiredFines)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching fines",
			"details": result.Error.Error(),
		})
		return
	}

	processedCount := 0

	for _, fine := range expiredFines {
		endDate := fine.CalculateEndDate()
		if now.After(endDate) {
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				// Marcar multa como inactiva
				fine.EndDate = &endDate
				fine.IsActive = false
				if err := tx.Save(&fine).Error; err != nil {
					return err
				}

				// Crear registro en historial
				fineHistory := models.FineHistory{
					UserID:           fine.UserID,
					StartDate:        fine.StartDate,
					EndDate:          endDate,
					AccumulatedDays:  fine.AccumulatedDays,
					TotalPenaltyDays: fine.AccumulatedDays * 2,
				}
				if err := tx.Create(&fineHistory).Error; err != nil {
					return err
				}

				// Cambiar estado del usuario a activo
				if err := tx.Model(&models.User{}).
					Where("id = ?", fine.UserID).
					Update("status", models.ACTIVE).Error; err != nil {
					return err
				}

				return nil
			})

			if err == nil {
				processedCount++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Expired fines processed successfully",
		"processed_count": processedCount,
		"total_checked":   len(expiredFines),
	})
}

// @Summary		Get all fines (Admin only)
// @Description	Get list of all fines in the system
// @Tags			fines
// @Accept			json
// @Produce		json
// @Param			page	query		int		false	"Page number"		default(1)
// @Param			limit	query		int		false	"Items per page"	default(10)
// @Param			status	query		string	false	"Fine status (active, expired)"
// @Success		200		{array}		models.Fine
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Security		BearerAuth
// @Router			/admin/fines [get]
func GetAllFines(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Fine{}).Preload("User")

	// Filtrar por estado si se especifica
	switch status {
	case "active":
		query = query.Where("is_active = true")
	case "expired":
		query = query.Where("is_active = false")
	}

	var fines []models.Fine
	var total int64

	query.Count(&total)
	result := query.Order("start_date DESC").Offset(offset).Limit(limit).Find(&fines)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching fines",
			"details": result.Error.Error(),
		})
		return
	}

	// Enriquecer datos de multas
	for i := range fines {
		if fines[i].IsActive {
			endDate := fines[i].CalculateEndDate()
			fines[i].EndDate = &endDate
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": fines,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

//	@Summary		Manually close fine (Admin only)
//	@Description	Manually close an active fine
//	@Tags			fines
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Fine ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map
