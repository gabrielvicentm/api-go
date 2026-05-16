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
	funcionarioRepo := repository.NewFuncionarioRepository(db, dataEncryptionKey)
	motoristaRepo := repository.NewMotoristaRepository(db, dataEncryptionKey)
	veiculoRepo := repository.NewVeiculoRepository(db)
	clienteRepo := repository.NewClienteRepository(db)
	tipoCargaRepo := repository.NewTipoCargaRepository(db)
	manutencaoRepo := repository.NewManutencaoRepository(db)
	viagemRepo := repository.NewViagemRepository(db)
	abastecimentoRepo := repository.NewAbastecimentoRepository(db)
	ocorrenciaRepo := repository.NewOcorrenciaRepository(db)
	notificacaoRepo := repository.NewNotificacaoRepository(db)
	viagemService := service.NewViagemService(viagemRepo, r2Storage)
	abastecimentoService := service.NewAbastecimentoService(abastecimentoRepo, viagemRepo, r2Storage)
	ocorrenciaService := service.NewOcorrenciaService(ocorrenciaRepo, viagemRepo, r2Storage)
	expoPushClient := service.NewExpoPushClientFromEnv()
	notificacaoService := service.NewNotificacaoService(notificacaoRepo, expoPushClient)
	authService := service.NewAuthService(authRepo, tokenManager)
	authMiddleware := middleware.AuthMiddleware(tokenManager)
	authHandler := handler.NewAuthHandler(authService, authMiddleware)
	dashboardHandler := handler.NewDashboardHandler()
	adminUserHandler := handler.NewAdminUserHandler(authService)
	funcionarioHandler := handler.NewFuncionarioHandler(funcionarioRepo, r2Storage)
	motoristaHandler := handler.NewMotoristaHandler(motoristaRepo, r2Storage)
	veiculoHandler := handler.NewVeiculoHandler(veiculoRepo)
	clienteHandler := handler.NewClienteHandler(clienteRepo)
	tipoCargaHandler := handler.NewTipoCargaHandler(tipoCargaRepo)
	viagemHandler := handler.NewViagemHandler(viagemRepo, viagemService)
	ocorrenciaHandler := handler.NewOcorrenciaHandler(ocorrenciaService)
	abastecimentoHandler := handler.NewAbastecimentoHandler(abastecimentoService)
	notificacaoHandler := handler.NewNotificacaoHandler(notificacaoService)
	manutencaoHandler := handler.NewManutencaoHandler(manutencaoRepo)
	relatorioHandler := handler.NewRelatorioHandler()
	historicoAlteracoesHandler := handler.NewHistoricoAlteracoesHandler(viagemRepo)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Static("/uploads", "./uploads")

	authHandler.RegisterRoutes(r)

	internal := r.Group("/internal")
	internal.Use(middleware.InternalAuthMiddlewareFromEnv())
	notificacaoHandler.RegisterInternalRoutes(internal)

	admin := r.Group("/admin")
	admin.Use(
		authMiddleware,
		middleware.RequireActorTypes("admin"),
	)

	dashboardHandler.RegisterAdminRoutes(admin)
	adminUserHandler.RegisterAdminRoutes(admin)
	funcionarioHandler.RegisterAdminRoutes(admin)
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
	historicoAlteracoesHandler.RegisterAdminRoutes(admin)

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
