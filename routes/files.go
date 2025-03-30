package routes

import (
	"file-sharing/handlers"
	"file-sharing/middlewares" // Import middleware package

	"github.com/gin-gonic/gin"
)

// Setup file routes
func FileRoutes(r *gin.Engine) {
	auth := r.Group("/") // Create a group for authenticated routes

	auth.Use(middlewares.AuthMiddleware())    // Apply auth middleware
	auth.POST("/upload", handlers.UploadFile) // Protect this route
	auth.GET("/files", handlers.GetFiles)     //  Protect this route
	auth.GET("/share/:id", handlers.ShareFile)
	auth.GET("/download/:id", handlers.DownloadFile) // Protect this route

	// r.GET("/files", handlers.GetFiles)          // Retrieve user files
	// r.GET("/share/:file_id", handlers.ShareFile) // Generate shareable link
	// r.GET("/download/:file_id", handlers.DownloadFile) // Secure file download
}
