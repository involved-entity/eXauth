package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	application := app.New(config.MustLoad())

	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()
}
