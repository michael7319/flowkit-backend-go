package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token
func GenerateToken(userID primitive.ObjectID) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key"
	}

	expirationTime := time.Now().Add(720 * time.Hour) // 30 days
	claims := &Claims{
		UserID: userID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "default-secret-key"
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get user from database
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		var user models.User
		err = config.UsersCollection.FindOne(c.Request.Context(), bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not found",
			})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User account is deactivated",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userId", userID)
		c.Next()
	}
}

// AuthorizeRoles checks if user has required role
func AuthorizeRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not found in context",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(models.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid user data",
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "User role '" + user.Role + "' is not authorized to access this route",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser gets user from context
func GetCurrentUser(c *gin.Context) (*models.User, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, jwt.ErrTokenMalformed
	}

	user, ok := userInterface.(models.User)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return &user, nil
}

// GetCurrentUserID gets user ID from context
func GetCurrentUserID(c *gin.Context) (primitive.ObjectID, error) {
	userIDInterface, exists := c.Get("userId")
	if !exists {
		return primitive.NilObjectID, jwt.ErrTokenMalformed
	}

	userID, ok := userIDInterface.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, jwt.ErrTokenMalformed
	}

	return userID, nil
}
