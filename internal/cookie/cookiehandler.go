package cookie

import (
	"fmt"
	"net/http"
	"slices"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	cookieName = "Authorization"
	userIDKey  = "user_id"

	registerURL = "/api/user/register"
	loginURL    = "/api/user/login"
)

var (
	cookieExpiry = 7 * 24 * time.Hour

	noAuthRoutes = []string{
		registerURL,
		loginURL,
	}
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func CookieHandler(signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(noAuthRoutes, c.Request.RequestURI) {
			c.Next()
			fmt.Println("############3", c.Keys)
			setCookie(c, signingKey)
			return
		}

		var userID string
		authHeader := c.GetHeader(cookieName)
		fmt.Println("HHHHH", authHeader)
		if authHeader != "" {
			token, err := jwt.ParseWithClaims(authHeader, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(signingKey), nil
			})

			fmt.Println("token", token)

			if err == nil && token.Valid {
				fmt.Println(token.Claims)
				claims := token.Claims.(*Claims)
				fmt.Println("UUU", claims.UserID)
				userID = claims.UserID
				c.Set(GetUserKey(), userID)
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func GetUserKey() string {
	return userIDKey
}

func setCookie(c *gin.Context, signingKey string) {
	rawUserID, exist := c.Get(GetUserKey())
	fmt.Println(exist)
	if !exist {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID := rawUserID.(string)
	fmt.Println("userID", userID)
	if userID == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(signingKey))
	fmt.Println(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.SetCookie(
		cookieName,
		tokenString,
		int(cookieExpiry.Seconds()),
		"/",
		"",
		true,
		true,
	)
	c.Header(cookieName, tokenString)
}
