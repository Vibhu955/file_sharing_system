package routes

import (
	"file-sharing/handlers"
	"file-sharing/middlewares" // Import middleware package

	"github.com/gin-gonic/gin"
)

// Setup file routes
func FileRoutes(r *gin.Engine) {
	auth := r.Group("/") // Create a group for authenticated routes

	auth.Use(middlewares.AuthMiddleware())   
	auth.POST("/upload", handlers.UploadFile) // upload only when logged in 
	auth.GET("/files", handlers.GetFiles)     //  get files of logged in user
	auth.GET("/share/:id", handlers.ShareFile) // files shared by logged in user
	auth.GET("/download/:id", handlers.DownloadFile) // logged in user can download files
	auth.GET("/search", handlers.SearchFiles) // logged in user can search files
	// create delete file route similarly
	// auth.DELETE("/delete/:id", handlers.DeleteFile) // logged in user can delete its own files
}
