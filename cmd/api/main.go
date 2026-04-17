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
	dashboardHandler := handler.NewDashboardHandler()
	adminUserHandler := handler.NewAdminUserHandler()
	motoristaHandler := handler.NewMotoristaHandler()
	veiculoHandler := handler.NewVeiculoHandler()
	clienteHandler := handler.NewClienteHandler()
	tipoCargaHandler := handler.NewTipoCargaHandler()
	viagemHandler := handler.NewViagemHandler()
	ocorrenciaHandler := handler.NewOcorrenciaHandler()
	abastecimentoHandler := handler.NewAbastecimentoHandler()
	notificacaoHandler := handler.NewNotificacaoHandler()
	manutencaoHandler := handler.NewManutencaoHandler()
	relatorioHandler := handler.NewRelatorioHandler()

	r := gin.Default()

	authHandler.RegisterRoutes(r)

	admin := r.Group("/admin")
	admin.Use(
		authMiddleware,
		middleware.RequireActorTypes("admin"),
	)

	dashboardHandler.RegisterAdminRoutes(admin)
	adminUserHandler.RegisterAdminRoutes(admin)
	motoristaHandler.RegisterAdminRoutes(admin)
	veiculoHandler.RegisterAdminRoutes(admin)
	clienteHandler.RegisterAdminRoutes(admin)
	tipoCargaHandler.RegisterAdminRoutes(admin)
	viagemHandler.RegisterAdminRoutes(admin)
	ocorrenciaHandler.RegisterAdminRoutes(admin)
	abastecimentoHandler.RegisterAdminRoutes(admin)
	notificacaoHandler.RegisterAdminRoutes(admin)
	manutencaoHandler.RegisterAdminRoutes(admin)
	relatorioHandler.RegisterAdminRoutes(admin)

	superadmin := admin.Group("")
	superadmin.Use(middleware.RequireRoles("superadmin"))
	adminUserHandler.RegisterSuperadminRoutes(superadmin)

	motorista := r.Group("/motorista")
	motorista.Use(
		authMiddleware,
		middleware.RequireActorTypes("motorista"),
	)

	motoristaHandler.RegisterMotoristaRoutes(motorista)
	viagemHandler.RegisterMotoristaRoutes(motorista)
	ocorrenciaHandler.RegisterMotoristaRoutes(motorista)
	abastecimentoHandler.RegisterMotoristaRoutes(motorista)
	notificacaoHandler.RegisterMotoristaRoutes(motorista)

	r.Run(":8080")
}
