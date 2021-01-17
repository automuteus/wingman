package main

import (
	"github.com/automuteus/wingman/pkg/broker"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const DefaultWingmanPort = "8123"

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	galactusAddr := os.Getenv("GALACTUS_ADDR")
	if galactusAddr == "" {
		log.Fatal("no GALACTUS_ADDR specified; exiting")
	}

	brokerPort := os.Getenv("WINGMAN_PORT")
	if brokerPort == "" {
		log.Println("No WINGMAN_PORT provided. Defaulting to " + DefaultWingmanPort)
		brokerPort = DefaultWingmanPort
	}

	socketBroker := broker.NewBroker(galactusAddr, logger)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go socketBroker.Start(brokerPort)

	<-sc
	log.Println("Wingman received a kill/term signal and will now exit")
}
