package controllers

import (
	"net/http"
	"strconv"
	"time"

	"biblioteca-backend/database"
	"biblioteca-backend/models"

	"github.com/gin-gonic/gin"
)

// @Summary		Get all books
// @Description	Get list of all books with availability information
// @Tags			books
// @Accept			json
// @Produce		json
// @Success		200	{array}		models.Book
// @Failure		500	{object}	map[string]string
// @Router			/books [get]
func GetBooks(c *gin.Context) {
	var books []models.Book

	result := database.DB.Preload("Exemplars").Find(&books)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error al obtener libros",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  books,
		"count": len(books),
	})
}

// @Summary		Get book by ID
// @Description	Get detailed information about a specific book including recommendations
// @Tags			books
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Book ID"
// @Success		200	{object}	models.Book
// @Failure		400	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Router			/books/{id} [get]
func GetBookByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var book models.Book
	result := database.DB.Preload("Exemplars").
		Preload("Recommendations").
		Preload("Recommendations.TargetBook").
		First(&book, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Libro no encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": book})
}

// @Summary		Create a new book
// @Description	Create a new book (Admin only)
// @Tags			books
// @Accept			json
// @Produce		json
// @Param			book	body		models.Book	true	"Book data"
// @Success		201		{object}	models.Book
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Security		BearerAuth
// @Router			/books [post]
func CreateBook(c *gin.Context) {
	var book models.Book

	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Create(&book)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error al crear libro",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Libro creado exitosamente",
		"data":    book,
	})
}

// Health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"message":   "Biblioteca API funcionando correctamente",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
