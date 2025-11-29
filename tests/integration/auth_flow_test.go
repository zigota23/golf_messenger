package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/golf_messenger/internal/handler"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/repository"
	"github.com/yourusername/golf_messenger/internal/router"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestAuthFlow_Integration(t *testing.T) {
	db := setupTestDB(t)

	logger, _ := zap.NewDevelopment()

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	jwtSecret := "test-secret"
	accessDuration := 15 * time.Minute
	refreshDuration := 7 * 24 * time.Hour

	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		jwtSecret,
		accessDuration,
		refreshDuration,
	)
	userService := service.NewUserService(userRepo, nil)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	rt := router.NewRouter(
		authHandler,
		userHandler,
		logger,
		jwtSecret,
		[]string{"*"},
	)

	httpHandler := rt.SetupRoutes()

	t.Run("Complete Auth Flow", func(t *testing.T) {
		registerReq := map[string]string{
			"email":      "integration@example.com",
			"password":   "password123",
			"first_name": "Integration",
			"last_name":  "Test",
		}
		body, _ := json.Marshal(registerReq)

		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		httpHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var registerResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &registerResp)

		assert.True(t, registerResp["success"].(bool))
		assert.NotNil(t, registerResp["data"])

		data := registerResp["data"].(map[string]interface{})
		accessToken := data["access_token"].(string)
		refreshToken := data["refresh_token"].(string)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		getProfileReq := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		getProfileReq.Header.Set("Authorization", "Bearer "+accessToken)
		w = httptest.NewRecorder()

		httpHandler.ServeHTTP(w, getProfileReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var profileResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &profileResp)

		assert.True(t, profileResp["success"].(bool))
		profileData := profileResp["data"].(map[string]interface{})
		assert.Equal(t, "integration@example.com", profileData["email"])
		assert.Equal(t, "Integration", profileData["first_name"])

		refreshReq := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ = json.Marshal(refreshReq)

		refreshTokenReq := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		refreshTokenReq.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		httpHandler.ServeHTTP(w, refreshTokenReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var refreshResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &refreshResp)

		assert.True(t, refreshResp["success"].(bool))
		refreshData := refreshResp["data"].(map[string]interface{})
		newAccessToken := refreshData["access_token"].(string)
		newRefreshToken := refreshData["refresh_token"].(string)

		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, accessToken, newAccessToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)

		logoutReq := map[string]string{
			"refresh_token": newRefreshToken,
		}
		body, _ = json.Marshal(logoutReq)

		logoutTokenReq := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(body))
		logoutTokenReq.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		httpHandler.ServeHTTP(w, logoutTokenReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var logoutResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &logoutResp)

		assert.True(t, logoutResp["success"].(bool))
	})
}
