package main

import (
	"log"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"auth/internal/pkg/database"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&database.User{},
	)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}
	os.Stdout.WriteString(stmts)
}
