package app

import (
	cfg "auth/internal/config"
	authgrpc "auth/internal/grpc"
	"log/slog"
	"net"

	"auth/internal/database"
	"auth/internal/logger"
	"auth/internal/machinery"
	"auth/internal/redis"

	"google.golang.org/grpc"
)

type App struct {
	GRPCServer *grpc.Server
	Port       string
	Logger     *slog.Logger
}

func New(config *cfg.Config) *App {
	logger := logger.SetupLogger(config.Env)
	database.Init(config.DSN)
	redis.Init(config.Redis.Address, config.Redis.Password, config.Redis.DB)
	machinery.Init(config.Mail.Email, config.Mail.Password, config.Machinery.Broker, config.Machinery.ResultBackend)

	grpcServer := grpc.NewServer()

	authgrpc.Register(grpcServer)

	return &App{Port: config.Port, GRPCServer: grpcServer, Logger: logger}
}

func (a *App) Run() error {
	listener, _ := net.Listen("tcp", ":"+a.Port)

	a.Logger.Info("server started", "port", a.Port)

	if err := a.GRPCServer.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.GRPCServer.GracefulStop()

	a.Logger.Info("server stopped")
}
