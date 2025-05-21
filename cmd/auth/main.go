package main

import (
	"auth/internal"
	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/redis"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := config.MustLoad()
	logger := internal.SetupLogger(config.Env)
	database.Init(config.DSN)
	redis.Init(config.Redis.Address, config.Redis.Password, config.Redis.DB)
	application := app.New(config.GRPC.Port, logger)

	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()
}
