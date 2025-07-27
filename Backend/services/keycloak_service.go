package services

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
)

type KeycloakService struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	PublicKey    *rsa.PublicKey
	client       *resty.Client
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type KeycloakClaims struct {
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	jwt.RegisteredClaims
}

var keycloakService *KeycloakService

func InitKeycloakService() {
	keycloakService = &KeycloakService{
		BaseURL:      getEnv("KEYCLOAK_URL", "http://localhost:8000"),
		Realm:        getEnv("KEYCLOAK_REALM", "biblioteca"),
		ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "biblioteca-api"),
		ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", "biblioteca-api-secret-2024"),
		client:       resty.New(),
	}

	// Obtener clave pública para validar tokens
	err := keycloakService.fetchPublicKey()
	if err != nil {
		log.Printf("Warning: Could not fetch Keycloak public key: %v", err)
	}
}

func GetKeycloakService() *KeycloakService {
	if keycloakService == nil {
		InitKeycloakService()
	}
	return keycloakService
}

func (ks *KeycloakService) fetchPublicKey() error {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid_connect/certs", ks.BaseURL, ks.Realm)

	resp, err := ks.client.R().Get(url)
	if err != nil {
		return err
	}

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	err = json.Unmarshal(resp.Body(), &jwks)
	if err != nil {
		return err
	}

	// Por simplicidad, usamos la primera clave
	// En producción, deberías buscar la clave correcta por kid
	if len(jwks.Keys) > 0 {
		// Aquí deberías convertir n y e a RSA public key
		// Por ahora, implementaremos validación básica
		log.Println("Keycloak public key fetched successfully")
	}

	return nil
}

func (ks *KeycloakService) ValidateToken(tokenString string) (*KeycloakClaims, error) {
	// Remover "Bearer " si está presente
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse del token sin verificar signature (para desarrollo)
	// En producción, DEBES verificar la signature con la clave pública
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &KeycloakClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verificar expiración
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func (ks *KeycloakService) HasRole(claims *KeycloakClaims, role string) bool {
	// Verificar roles del realm
	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}

	// Verificar roles del cliente
	if clientRoles, exists := claims.ResourceAccess[ks.ClientID]; exists {
		for _, r := range clientRoles.Roles {
			if r == role {
				return true
			}
		}
	}

	return false
}

func (ks *KeycloakService) GetUserLogin(claims *KeycloakClaims) string {
	return claims.PreferredUsername
}

func (ks *KeycloakService) IntrospectToken(token string) (bool, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid_connect/token/introspect", ks.BaseURL, ks.Realm)

	resp, err := ks.client.R().
		SetFormData(map[string]string{
			"token":         token,
			"client_id":     ks.ClientID,
			"client_secret": ks.ClientSecret,
		}).
		Post(url)

	if err != nil {
		return false, err
	}

	var result struct {
		Active bool `json:"active"`
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return false, err
	}

	return result.Active, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
