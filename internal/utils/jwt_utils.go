package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken - генерирует JWT токен, содержащий хэш пароля и срок действия
func GenerateToken(password string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // Создание нового токена с HMAC-SHA256

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user"                               // Установка субъекта токена
	claims["exp"] = time.Now().Add(8 * time.Hour).Unix() // Установка срока действия токена (8 часов)
	claims["hash"] = GeneratePasswordHash(password)      // Хэш пароля для дальнейшей проверки

	// Подписание токена с использованием пароля
	tokenString, err := token.SignedString([]byte(password))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken - парсит JWT токен и проверяет его валидность
func ParseToken(tokenString, password string) (*jwt.Token, error) {
	// Парсинг токена и проверка метода подписи
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(password), nil // Возврат ключа (пароль)
	})
	if err != nil || !token.Valid {
		return nil, err // Ошибка, если токен недействителен
	}

	// Проверка claims (утверждений) токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("недопустимые утверждения токена")
	}

	// Сравнение хэша пароля для подтверждения подлинности
	if claims["hash"] != GeneratePasswordHash(password) {
		return nil, fmt.Errorf("недопустимый хэш токена")
	}

	return token, nil
}

// GeneratePasswordHash - генерирует хэш пароля с использованием SHA-256
func GeneratePasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password)) // Хэширование пароля
	return hex.EncodeToString(hash[:])      // Преобразование хэша в строку
}
