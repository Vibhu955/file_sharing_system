package handlers

import (
	// "fmt"
	// "io"
	"log"
	"net/http"

	// "os"
	// "path/filepath"
	"time"
	"file-sharing/config"
	"file-sharing/storage"

	"github.com/gin-gonic/gin"
)

// UploadFile handles file uploads
func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}
	defer file.Close()

	fileName := header.Filename
	fileSize := header.Size
	uploadTime := time.Now()

	// Choose storage method (local or S3)
	useS3 := false // Change this to true if using S3

	var fileURL string
	if useS3 {
		fileURL, err = storage.UploadToS3(file, fileName)
	} else {
		fileURL, err = storage.SaveLocally(file, fileName)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload failed"})
		return
	}

	// Save file metadata to DB
	log.Println("Saving file metadata to DB ", c.GetString("email"))
	_, err = config.DB.Exec("INSERT INTO files (file_name, size, url, upload_at, user_email) VALUES ($1, $2, $3, $4, $5)", fileName, fileSize, fileURL, uploadTime, c.GetString("email"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "url": fileURL})
}
