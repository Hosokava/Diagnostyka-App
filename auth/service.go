package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"gin-quickstart/database"
	"gin-quickstart/models"
	"time"

	"github.com/gin-gonic/gin"
)

func PerformTokenRefresh(c *gin.Context) (uint, string, error) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return 0, "", err
	}

	hash := sha256.Sum256([]byte(refreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	var rf models.RefreshToken
	if err := database.DB.Where("token_hash = ?", hashedToken).First(&rf).Error; err != nil {
		return 0, "", errors.New("session not found")
	}

	if time.Now().After(rf.ExpiresAt) {
		database.DB.Delete(&rf)
		return 0, "", errors.New("session expired")
	}

	newAccessToken, err := GenerateAccessToken(rf.UserID, rf.UserType)
	if err != nil {
		return 0, "", err
	}
	newRawRefresh, newHashedRefresh, err := GenerateRefreshToken()
	if err != nil {
		return 0, "", err
	}

	rf.TokenHash = newHashedRefresh
	rf.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	database.DB.Save(&rf)

	c.SetCookie("access_token", newAccessToken, 900, "/", "", false, true)
	c.SetCookie("refresh_token", newRawRefresh, 7*24*3600, "/", "", false, true)

	return rf.UserID, rf.UserType, nil
}
