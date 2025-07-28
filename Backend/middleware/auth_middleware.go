package middleware

import (
	"errors"
	"net/http"
	"strings"

	"biblioteca-backend/database"
	"biblioteca-backend/models"
	"biblioteca-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Validar token con Keycloak
		keycloakService := services.GetKeycloakService()
		claims, err := keycloakService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Obtener login desde el token
		userLogin := keycloakService.GetUserLogin(claims)

		// Buscar usuario en base de datos por login
		var user models.User
		err = database.DB.Where("login = ?", userLogin).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Buscar por email primero
			var existingByEmail models.User
			errEmail := database.DB.Where("email = ?", claims.Email).First(&existingByEmail).Error
			if errEmail == nil {
				user = existingByEmail
				userLogin = existingByEmail.Login
			} else if errors.Is(errEmail, gorm.ErrRecordNotFound) {
				// Crear nuevo usuario
				user = models.User{
					Login:  userLogin,
					Name:   claims.GivenName + " " + claims.FamilyName,
					Email:  claims.Email,
					Status: models.ACTIVE,
				}
				if err := database.DB.Create(&user).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Error creating user",
						"details": err.Error(),
					})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error checking user email",
					"details": errEmail.Error(),
				})
				c.Abort()
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error fetching user",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Guardar en el contexto
		c.Set("user_claims", claims)
		c.Set("user_login", userLogin)
		c.Set("token", tokenString)

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No user claims found",
				"code":  "NO_USER_CLAIMS",
			})
			c.Abort()
			return
		}

		keycloakClaims, ok := claims.(*services.KeycloakClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user claims",
				"code":  "INVALID_USER_CLAIMS",
			})
			c.Abort()
			return
		}

		keycloakService := services.GetKeycloakService()
		if !keycloakService.HasRole(keycloakClaims, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"code":          "INSUFFICIENT_PERMISSIONS",
				"required_role": role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to get current user login from context
func GetCurrentUserLogin(c *gin.Context) (string, bool) {
	userLogin, exists := c.Get("user_login")
	if !exists {
		return "", false
	}

	login, ok := userLogin.(string)
	return login, ok
}

// Helper function to get current user claims from context
func GetCurrentUserClaims(c *gin.Context) (*services.KeycloakClaims, bool) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}

	keycloakClaims, ok := claims.(*services.KeycloakClaims)
	return keycloakClaims, ok
}
