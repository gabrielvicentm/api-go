package main

import (
	"context"

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

	r2Storage, err := service.NewR2StorageFromEnv(context.Background())
	if err != nil {
		panic(err)
	}

	dataEncryptionKey, err := security.DataEncryptionKeyFromEnv()
	if err != nil {
		panic(err)
	}

	authRepo := repository.NewAuthRepository(db)
	motoristaRepo := repository.NewMotoristaRepository(db, dataEncryptionKey)
	veiculoRepo := repository.NewVeiculoRepository(db)
	clienteRepo := repository.NewClienteRepository(db)
	tipoCargaRepo := repository.NewTipoCargaRepository(db)
<<<<<<< HEAD
	manutencaoRepo := repository.NewManutencaoRepository(db)
=======
	viagemRepo := repository.NewViagemRepository(db)
	viagemService := service.NewViagemService(viagemRepo, r2Storage)
>>>>>>> b2df43310ad46be13b75643acdcb9f4967961b73
	authService := service.NewAuthService(authRepo, tokenManager)
	authMiddleware := middleware.AuthMiddleware(tokenManager)
	authHandler := handler.NewAuthHandler(authService, authMiddleware)
	dashboardHandler := handler.NewDashboardHandler()
	adminUserHandler := handler.NewAdminUserHandler()
	motoristaHandler := handler.NewMotoristaHandler(motoristaRepo, r2Storage)
	veiculoHandler := handler.NewVeiculoHandler(veiculoRepo)
	clienteHandler := handler.NewClienteHandler(clienteRepo)
	tipoCargaHandler := handler.NewTipoCargaHandler(tipoCargaRepo)
	viagemHandler := handler.NewViagemHandler(viagemRepo, viagemService)
	ocorrenciaHandler := handler.NewOcorrenciaHandler()
	abastecimentoHandler := handler.NewAbastecimentoHandler()
	notificacaoHandler := handler.NewNotificacaoHandler()
	manutencaoHandler := handler.NewManutencaoHandler(manutencaoRepo)
	relatorioHandler := handler.NewRelatorioHandler()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Static("/uploads", "./uploads")

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
