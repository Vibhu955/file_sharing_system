package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"file-sharing/config"
	"file-sharing/routes"
	"file-sharing/middlewares"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtSecret = []byte("your_secret_key")

// User struct
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// JWT Claims
type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Register endpoint
func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	if config.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}
	_, err = config.DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", user.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login endpoint
func login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if config.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var storedPassword string
	err := config.DB.QueryRow("SELECT password FROM users WHERE email=$1", user.Email).Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	_, err = config.DB.Exec("UPDATE users SET token=$1 WHERE email=$2", tokenString, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}



func main() {
	config.ConnectDB()    // Initialize PostgreSQL connection
	config.ConnectRedis() // Redis
	db = config.DB

	log.Println("Database connected successfully")
	r := gin.Default()

	r.POST("/register", register)
	r.POST("/login", login)

	r.GET("/protected", middlewares.AuthMiddleware(), func(c *gin.Context) {
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{"message": "Protected content", "user": email})
	})
	r.Static("/uploads", "./uploads") // Serve uploads folder

	routes.SetupRoutes(r)

	routes.FileRoutes(r)

	r.Run(":8080")
}
