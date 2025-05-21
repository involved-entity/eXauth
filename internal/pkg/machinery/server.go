package machinery

import (
	"auth/internal/tasks"
	"os"

	"github.com/RichardKnop/machinery/v2"
	backendsAmqp "github.com/RichardKnop/machinery/v2/backends/amqp"
	brokersAmqp "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"

	"log"
)

func New(configEmail string, configPassword string, configBroker string, configResultBackend string) *machinery.Server {
	cnf := &config.Config{
		DefaultQueue:    "tasks",
		ResultsExpireIn: 3600,
		Broker:          configBroker,
		ResultBackend:   configResultBackend,
		AMQP: &config.AMQPConfig{
			Exchange:      "machinery_exchange",
			ExchangeType:  "direct",
			BindingKey:    "machinery_task",
			PrefetchCount: 3,
		},
	}

	broker := brokersAmqp.New(cnf)
	backend := backendsAmqp.New(cnf)
	lock := eager.New()

	server := machinery.NewServer(cnf, broker, backend, lock)

	t := make(map[string]interface{})
	t["send_email"] = tasks.SendVerificationEmail(configEmail, configPassword)
	t["reset_password"] = tasks.SendResetPasswordEmail(configEmail, configPassword)
	err := server.RegisterTasks(t)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return server
}
