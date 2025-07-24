package controllers

import "github.com/gin-gonic/gin"

// @Summary Create loan
// @Description Request a book loan
// @Tags loans
// @Accept json
// @Produce json
// @Param loan body models.LoanRequest true "Loan request"
// @Success 201 {object} models.Loan
// @Router /loans [post]
func CreateLoan(c *gin.Context) {
	// Implementar lógica de préstamo
	// Verificar disponibilidad, límites, multas, etc.
}

// @Summary Get user loans
// @Description Get current loans for authenticated user
// @Tags loans
// @Accept json
// @Produce json
// @Success 200 {array} models.Loan
// @Router /loans/my [get]
func GetMyLoans(c *gin.Context) {
	// Implementar obtención de préstamos del usuario
}
