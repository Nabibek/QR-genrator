package handlers

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginRequest - запрос на вход
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse - ответ при успешном входе
type LoginResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role,omitempty"`
	Token    string `json:"token,omitempty"`
	Error    string `json:"error,omitempty"`
}

// hashPassword - простая хеширование пароля (SHA256)
// ⚠️ В production используйте bcrypt!
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// Login - обработчик POST /api/login
// Проверяет учётные данные пользователя (имя и пароль)
func Login(c *gin.Context) {
	var req LoginRequest

	// Парси JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Error:   "Невалидные данные: " + err.Error(),
		})
		return
	}

	db := database.GetDB()
	var user models.User

	// Ищем пользователя по имени
	if err := db.First(&user, "username = ?", req.Username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, LoginResponse{
				Success: false,
				Error:   "Пользователь не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, LoginResponse{
				Success: false,
				Error:   "Ошибка при проверке учётных данных",
			})
		}
		return
	}

	// Проверяем пароль (сравниваем хеши)
	hashedPassword := hashPassword(req.Password)
	if user.PasswordHash != hashedPassword {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Неверный пароль",
		})
		return
	}

	// Успешная авторизация
	// ⚠️ В production нужна генерация JWT токена!
	c.JSON(http.StatusOK, LoginResponse{
		Success:  true,
		Message:  "Успешная авторизация",
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    fmt.Sprintf("bearer_%s", user.ID), // Упрощённый временный токен
	})
}

// CurrentUser - обработчик GET /api/me
// Возвращает информацию о текущем пользователе (требует авторизацию)
type CurrentUserResponse struct {
	Success bool         `json:"success"`
	User    *models.User `json:"user,omitempty"`
	Error   string       `json:"error,omitempty"`
}

func CurrentUser(c *gin.Context) {
	// Получаем userID из контекста (должен быть установлен middleware'ем авторизации)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, CurrentUserResponse{
			Success: false,
			Error:   "Пользователь не авторизован",
		})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, CurrentUserResponse{
				Success: false,
				Error:   "Пользователь не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, CurrentUserResponse{
				Success: false,
				Error:   "Ошибка при получении информации пользователя",
			})
		}
		return
	}

	c.JSON(http.StatusOK, CurrentUserResponse{
		Success: true,
		User:    &user,
	})
}
