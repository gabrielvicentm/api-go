package main

import (
	"github.com/gabrielvicentm/api-go.git/config"
	"github.com/gabrielvicentm/api-go.git/internal/handler"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gabrielvicentm/api-go.git/internal/security"
	"github.com/gabrielvicentm/api-go.git/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db := config.NewDBConnection()

	defer db.Close()

	tokenManager, err := security.NewTokenManagerFromEnv()
	if err != nil {
		panic(err)
	}

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, tokenManager)
	authMiddleware := middleware.AuthMiddleware(tokenManager)
	authHandler := handler.NewAuthHandler(authService, authMiddleware)
	protectedHandler := handler.NewProtectedHandler()

	r := gin.Default()

	authHandler.RegisterRoutes(r)
	protectedHandler.RegisterAdminRoutes(r, authMiddleware)
	protectedHandler.RegisterMotoristaRoutes(r, authMiddleware)

	r.Run(":8080")
}
