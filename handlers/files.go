package handlers

import (
	"context"
	"encoding/json"
	"file-sharing/config"
	"file-sharing/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Get user’s uploaded files ( Redis caching)

func GetFiles(c *gin.Context) {
	userEmail := c.GetString("email")

	// log.Println("mail", userEmail)

	//  Redis cache
	cacheKey := "files:" + userEmail
	cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
	// log.Println("Cached Data:", cachedData, "Error:", err)

	if err == nil && cachedData != "null" {
		var files []models.File
		if err := json.Unmarshal([]byte(cachedData), &files); err == nil {
			c.JSON(http.StatusOK, files)
			return
		}
	}
	// log.Println("Cache miss for", cacheKey)

	//DB
	rows, err := config.DB.Query("SELECT id,user_email, file_name, size, url, upload_at  FROM files WHERE user_email = $1", userEmail)

	if err != nil {
		// log.Println("Database query error:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch files"})
		return
	}
	defer rows.Close()

	var files []models.File
	// count := 0
	for rows.Next() {
		var file models.File
		// count++
		// log.Println("Processing row:", count) // Debugging

		if err := rows.Scan(&file.ID, &file.UserEmail, &file.FileName, &file.Size, &file.URL, &file.UploadAt); err != nil {
			log.Println("Error sca	nning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning files"})
			return
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No files found"})
		return
	}

	// Cache result -> put in Redis
	cacheData, _ := json.Marshal(files)
	config.RedisClient.Set(ctx, cacheKey, cacheData, redis.KeepTTL)
	c.JSON(http.StatusOK, files)
}

// public link

func ShareFile(c *gin.Context) {
	fileID := c.Param("id")
	// log.Println("Sharing file with ID:", fileID)

	// file details from DB
	var file models.File
	err := config.DB.QueryRow("SELECT id, file_name, url FROM files WHERE id = $1", fileID).
		Scan(&file.ID, &file.FileName, &file.URL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// expiring link (validity- 24 hours)
	expiration := time.Now().Add(24 * time.Hour).Unix()
	shareURL := fmt.Sprintf("http://localhost:8080/download/%d?expires=%d", file.ID, expiration)
	c.JSON(http.StatusOK, gin.H{"file_name": file.FileName, "share_url": shareURL})
}

// Download file using that public link

func DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	expiresStr := c.Query("expires")

	// check for expiry
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Link expired"})
		return
	}
	// log.Println("Downloading file with ID:", fileID)

	// Fetch from DB
	var file models.File
	err = config.DB.QueryRow("SELECT url FROM files WHERE id = $1", fileID).Scan(&file.URL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Redirect to file URL
	c.Redirect(http.StatusFound, file.URL)
}

// file search
func SearchFiles(c *gin.Context) {
	userEmail := c.GetString("email")

	fname := c.Query("fname")     // File name (partial match)
	uploadDate := c.Query("date") // Upload date (YYYY-MM-DD)
	fileType := c.Query("type")   // File type (e.g., pdf, txt)

	// Cache Key- Redis
	cacheKey := fmt.Sprintf("search:%s:%s:%s:%s", userEmail, fname, uploadDate, fileType)
	if cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result(); err == nil {
		var files []models.File
		json.Unmarshal([]byte(cachedData), &files)
		c.JSON(http.StatusOK, files)
		return
	}

	// Base Query
	sqlQuery := `SELECT id, file_name, size, upload_at, url FROM files WHERE user_email = $1`
	args := []interface{}{userEmail}
	argIndex := 2

	// Filter by

	// file name (ILIKE : case-insensitive , partial match)
	if fname != "" {
		sqlQuery += fmt.Sprintf(` AND LEFT(file_name, LENGTH(file_name) - POSITION('.' IN REVERSE(file_name))) ILIKE $%d`, argIndex)
		args = append(args, "%"+fname+"%")
		argIndex++
	}

	// upload date
	if uploadDate != "" {
		sqlQuery += fmt.Sprintf(` AND DATE(upload_at) = $%d`, argIndex)
		args = append(args, uploadDate)
		argIndex++
	}

	//  file extension
	if fileType != "" {
		sqlQuery += fmt.Sprintf(` AND file_name ILIKE $%d`, argIndex)
		args = append(args, "%."+fileType)
		argIndex++
	}

	//execute
	rows, err := config.DB.Query(sqlQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch files"})
		return
	}
	defer rows.Close()

	// Results
	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(&file.ID, &file.FileName, &file.Size, &file.UploadAt, &file.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning files"})
			return
		}
		files = append(files, file)
	}

	// Results -> Redis (5 minutes cache)
	cacheData, _ := json.Marshal(files)
	config.RedisClient.Set(ctx, cacheKey, cacheData, 5*time.Minute)
	c.JSON(http.StatusOK, files)
}
