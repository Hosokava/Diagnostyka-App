package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"gin-quickstart/auth"
	"gin-quickstart/database"
	"gin-quickstart/models"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	var userID uint
	var passwordHash string
	switch body.Role {
	case "patient":
		var patient models.Patient
		if err := database.DB.Where("email = ?", body.Email).First(&patient).Error; err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		userID, passwordHash = patient.ID, patient.PasswordHash
	case "doctor":
		var doctor models.Doctor
		if err := database.DB.Where("email = ?", body.Email).First(&doctor).Error; err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		userID, passwordHash = doctor.ID, doctor.PasswordHash
	default:
		c.JSON(400, gin.H{"error": "invalid role"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)); err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, _ := auth.GenerateAccessToken(userID, body.Role)
	rawRefresh, hashedRefresh, _ := auth.GenerateRefreshToken()

	database.DB.Create(&models.RefreshToken{
		TokenHash: hashedRefresh,
		UserID:    userID,
		UserType:  body.Role,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	c.SetCookie("access_token", accessToken, 900, "/", "", false, true)
	c.SetCookie("refresh_token", rawRefresh, 7*24*3600, "/", "", false, true)

	c.JSON(200, gin.H{"message": "logged in"})
}

func Logout(c *gin.Context) {
	if cookieToken, err := c.Cookie("refresh_token"); err == nil {
		hash := sha256.Sum256([]byte(cookieToken))
		hashedToken := hex.EncodeToString(hash[:])
		database.DB.Where("token_hash = ?", hashedToken).Delete(&models.RefreshToken{})
	}

	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(200, gin.H{"message": "logged out"})
}

func Register(c *gin.Context) {
	var body struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		Role           string `json:"role"`
		Specialization string `json:"specialization"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}

	switch body.Role {
	case "patient":
		patient := models.Patient{
			Email:        body.Email,
			PasswordHash: string(hash),
		}
		if err := database.DB.Create(&patient).Error; err != nil {
			c.JSON(400, gin.H{"error": "email may already exist"})
			return
		}
	case "doctor":
		doctor := models.Doctor{
			Email:          body.Email,
			PasswordHash:   string(hash),
			Specialization: body.Specialization,
		}
		if err := database.DB.Create(&doctor).Error; err != nil {
			c.JSON(400, gin.H{"error": "email may already exist"})
			return
		}
	default:
		c.JSON(400, gin.H{"error": "invalid role"})
		return
	}

	c.JSON(200, gin.H{"message": "registered successfully"})
}

func Refresh(c *gin.Context) {
	_, _, err := auth.PerformTokenRefresh(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	c.JSON(200, gin.H{"message": "Tokens refreshed"})
}

func GetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	c.JSON(200, gin.H{
		"authenticated": true,
		"user_id":       userID,
		"role":          role,
		"message":       "You are currently logged in",
	})
}
