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

// @Summary		Get my loans
// @Description	Get current loans for authenticated user
// @Tags			loans
// @Accept			json
// @Produce		json
// @Success		200	{array}		models.LoanResponse
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		BearerAuth
// @Router			/loans/my [get]
func GetMyLoans(c *gin.Context) {
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

	// Obtener préstamos activos
	var loans []models.Loan
	result := database.DB.Preload("Exemplar").
		Preload("Exemplar.Book").
		Where("user_id = ? AND due_date >= ?", user.ID, time.Now()).
		Find(&loans)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching loans",
			"details": result.Error.Error(),
		})
		return
	}

	// Convertir a formato de respuesta
	var loanResponses []models.LoanResponse
	for _, loan := range loans {
		loanResponse := models.LoanResponse{
			ID:           loan.ID,
			BookTitle:    loan.Exemplar.Book.Title,
			BookAuthor:   loan.Exemplar.Book.Author,
			ExemplarCode: loan.Exemplar.Code,
			LoanDate:     loan.LoanDate,
			DueDate:      loan.DueDate,
			ReturnedAt:   loan.ReturnedAt,
			IsOverdue:    loan.IsOverdue(),
			OverdueDays:  loan.GetOverdueDays(),
		}
		loanResponses = append(loanResponses, loanResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  loanResponses,
		"count": len(loanResponses),
	})
}

// @Summary		Create loan
// @Description	Request a book loan
// @Tags			loans
// @Accept			json
// @Produce		json
// @Param			loan	body		models.LoanRequest	true	"Loan request"
// @Success		201		{object}	models.Loan
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		409		{object}	map[string]string
// @Security		BearerAuth
// @Router			/loans [post]
func CreateLoan(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var loanRequest models.LoanRequest
	if err := c.ShouldBindJSON(&loanRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar usuario
	var user models.User
	if err := database.DB.Preload("Fine").Where("login = ?", userLogin).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verificar si el usuario puede pedir prestado
	if !user.CanBorrow() {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "User cannot borrow books",
			"reason": string(user.Status),
		})
		return
	}

// Buscar ejemplar disponible (tratando exemplar_id como book_id)
var exemplar models.Exemplar
if err := database.DB.Preload("Book").
    Where("book_id = ? AND is_available = true", loanRequest.ExemplarID).
    First(&exemplar).Error; err != nil {
	c.JSON(http.StatusNotFound, gin.H{"error": "No hay ejemplares disponibles para este libro"})
	return
}

	// Crear préstamo en transacción
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Crear préstamo
		now := time.Now()
due := now.AddDate(0, 0, 5) // siempre 5 días
loan := models.Loan{
    UserID:     user.ID,
    ExemplarID: exemplar.ID,
    LoanDate:   now,
    DueDate:    due,
    ReturnedAt: &due,
}


		if err := tx.Create(&loan).Error; err != nil {
			return err
		}

		// Marcar ejemplar como no disponible
		if err := tx.Model(&exemplar).Update("is_available", false).Error; err != nil {
			return err
		}

		// Actualizar contador de ejemplares disponibles del libro
		if err := tx.Model(&exemplar.Book).
			UpdateColumn("available_exemplars", gorm.Expr("available_exemplars - 1")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creating loan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Loan created successfully",
		"book_title":    exemplar.Book.Title,
		"exemplar_code": exemplar.Code,
		"due_date":      time.Now().AddDate(0, 0, user.GetLoanDays()),
	})
}

// @Summary		Return loan
// @Description	Return a borrowed book
// @Tags			loans
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Loan ID"
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Security		BearerAuth
// @Router			/loans/{id}/return [put]
func ReturnLoan(c *gin.Context) {
	userLogin, exists := middleware.GetCurrentUserLogin(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loanID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	// Buscar usuario
	var user models.User
	if err := database.DB.Where("login = ?", userLogin).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Buscar préstamo
	var loan models.Loan
	result := database.DB.Preload("Exemplar").
		Preload("Exemplar.Book").
		Where("id = ? AND user_id = ? AND returned_at IS NULL", loanID, user.ID).
		First(&loan)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found or already returned"})
		return
	}

	now := time.Now()
	wasOverdue := loan.IsOverdue()
	overdueDays := loan.GetOverdueDays()

	// Procesar devolución en transacción
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Marcar préstamo como devuelto
		loan.ReturnedAt = &now
		if err := tx.Save(&loan).Error; err != nil {
			return err
		}

		// Marcar ejemplar como disponible
		if err := tx.Model(&loan.Exemplar).Update("is_available", true).Error; err != nil {
			return err
		}

		// Actualizar contador de ejemplares disponibles del libro
		if err := tx.Model(&loan.Exemplar.Book).
			UpdateColumn("available_exemplars", gorm.Expr("available_exemplars + 1")).Error; err != nil {
			return err
		}

		// Crear registro en historial
		loanHistory := models.LoanHistory{
			UserID:       loan.UserID,
			ExemplarID:   loan.ExemplarID,
			BookID:       loan.Exemplar.BookID,
			LoanDate:     loan.LoanDate,
			DueDate:      loan.DueDate,
			ReturnedDate: now,
			WasOverdue:   wasOverdue,
			OverdueDays:  overdueDays,
		}
		if err := tx.Create(&loanHistory).Error; err != nil {
			return err
		}

		// Si hubo retraso, manejar multa
		if wasOverdue {
			if err := handleOverdueLoan(tx, &user, overdueDays); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error returning loan",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"message":     "Book returned successfully",
		"book_title":  loan.Exemplar.Book.Title,
		"returned_at": now,
	}

	if wasOverdue {
		response["was_overdue"] = true
		response["overdue_days"] = overdueDays
		response["penalty_applied"] = true
	}

	c.JSON(http.StatusOK, response)
}

// @Summary		Get loan history
// @Description	Get loan history for authenticated user
// @Tags			loans
// @Accept			json
// @Produce		json
// @Param			page	query		int	false	"Page number"		default(1)
// @Param			limit	query		int	false	"Items per page"	default(10)
// @Success		200		{array}		models.LoanHistory
// @Failure		401		{object}	map[string]string
// @Security		BearerAuth
// @Router			/loans/history [get]
func GetLoanHistory(c *gin.Context) {
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

	var history []models.LoanHistory
	var total int64

	database.DB.Model(&models.LoanHistory{}).Where("user_id = ?", user.ID).Count(&total)

	result := database.DB.Preload("Book").
		Preload("Exemplar").
		Where("user_id = ?", user.ID).
		Order("returned_date DESC").
		Offset(offset).
		Limit(limit).
		Find(&history)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching loan history",
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

// Helper function para manejar préstamos con retraso
func handleOverdueLoan(tx *gorm.DB, user *models.User, overdueDays int) error {
	// Buscar multa activa existente
	var fine models.Fine
	result := tx.Where("user_id = ? AND is_active = true", user.ID).First(&fine)

	if result.Error != nil {
		// Crear nueva multa
		fine = models.Fine{
			UserID:          user.ID,
			StartDate:       time.Now(),
			AccumulatedDays: overdueDays,
			IsActive:        true,
		}
		if err := tx.Create(&fine).Error; err != nil {
			return err
		}
	} else {
		// Acumular días a la multa existente
		fine.AccumulatedDays += overdueDays
		if err := tx.Save(&fine).Error; err != nil {
			return err
		}
	}

	// Cambiar estado del usuario a multado
	if err := tx.Model(user).Update("status", models.FINED).Error; err != nil {
		return err
	}

	return nil
}

// @Summary		Get all loans (Admin only)
// @Description	Get list of all loans in the system
// @Tags			loans
// @Accept			json
// @Produce		json
// @Param			page	query		int		false	"Page number"		default(1)
// @Param			limit	query		int		false	"Items per page"	default(10)
// @Param			status	query		string	false	"Loan status (active, returned, overdue)"
// @Success		200		{array}		models.Loan
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Security		BearerAuth
// @Router			/admin/loans [get]
func GetAllLoans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Loan{}).
		Preload("User").
		Preload("Exemplar").
		Preload("Exemplar.Book")

	// Filtrar por estado si se especifica
	switch status {
	case "active":
		query = query.Where("returned_at IS NULL")
	case "returned":
		query = query.Where("returned_at IS NOT NULL")
	case "overdue":
		query = query.Where("returned_at IS NULL AND due_date < ?", time.Now())
	}

	var loans []models.Loan
	var total int64

	query.Count(&total)
	result := query.Order("loan_date DESC").Offset(offset).Limit(limit).Find(&loans)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching loans",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": loans,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
