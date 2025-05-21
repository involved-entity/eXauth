data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "./cmd/migrator/atlas-provider-gorm.go",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  url = "dsn var"
  dev = "docker://postgres/17-alpine/dev"
  
  migration {
    dir = "file://migrations"
  }
}
