package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"perlerbeads/internal/handler"
	"perlerbeads/internal/service"
)

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	uploadDir := filepath.Join(workDir, "uploads")
	outputDir := filepath.Join(workDir, "output")

	patternService := service.NewPatternService(uploadDir, outputDir)

	h := handler.NewHandler(patternService, outputDir)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api/v1")
	{
		api.POST("/generate", h.Generate)
		api.GET("/export", h.Export)
		api.GET("/health", h.Health)
	}

	r.Static("/static", "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Upload directory: %s", uploadDir)
	log.Printf("Output directory: %s", outputDir)
	log.Printf("API endpoints:")
	log.Printf("  POST /api/v1/generate - Generate pattern")
	log.Printf("  GET  /api/v1/export?id=xxx - Export pattern image")
	log.Printf("  GET  /api/v1/health - Health check")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
