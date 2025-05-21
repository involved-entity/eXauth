package main

import (
	conf "auth/internal/config"
	"auth/internal/machinery"
	"log"
)

func main() {
	config := conf.MustLoad()
	server := machinery.New(config.Mail.Email, config.Mail.Password, config.Machinery.Broker, config.Machinery.ResultBackend)
	worker := server.NewWorker("email_worker", 10)
	err := worker.Launch()
	if err != nil {
		log.Fatal(err)
	}
}
