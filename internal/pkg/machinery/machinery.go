package machinery

import (
	"log"

	"github.com/RichardKnop/machinery/v2"
)

var server *machinery.Server

func Init(configEmail string, configPassword string, configBroker string, configResultBackend string) {
	server = New(configEmail, configPassword, configBroker, configResultBackend)
}

func GetServer() *machinery.Server {
	if server == nil {
		log.Fatal("Machinery server not initialized")
	}
	return server
}
