package main

import (
	"auth/internal/pkg/config"
	"auth/internal/pkg/database"
)

func main() {
	config.MustLoad()
	conf := config.GetConfig()
	database.Init(conf.DSN)
	db := database.GetDB()
	db.Exec("DELETE FROM users;")
	db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")
}
