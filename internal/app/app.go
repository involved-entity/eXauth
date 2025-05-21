package app

import (
	authgrpc "auth/internal/grpc"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	GRPCServer *grpc.Server
	Port       string
	Logger     *slog.Logger
}

func New(port string, logger *slog.Logger) *App {
	grpcServer := grpc.NewServer()

	authgrpc.Register(grpcServer)

	return &App{Port: port, GRPCServer: grpcServer, Logger: logger}
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
