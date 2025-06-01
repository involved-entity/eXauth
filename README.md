# eXauth
__eXauth__ - gRPC сервис авторизации и аутентификации на `Golang`, написанный с использованием `Atlas`, `GORM`, `Redis`, `RabbitMQ`, `Machinery`, `PostgreSQL`, готовый к использованию в небольших проектах.

## Установка
```bash
git clone https://github.com/involved-entity/eXauth
cd eXauth
```

### Production запуск
Перед __Production__ запуском убедитесь, что у вас доступна `docker compose` команда (не путать с `docker-compose`), а так же создайте файл конфигурации `config/prod.yml` и настройте по примеру `config/prod.template.yml`.

```bash
make run-prod
```

### Функционал
С полным набором доступных методов можно ознакомиться в `.proto` файлах в папке `api/`. 
