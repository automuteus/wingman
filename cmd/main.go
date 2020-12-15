package main

import (
	"github.com/automuteus/wingman/pkg/broker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const DefaultWingmanPort = "8123"

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("No REDIS_ADDR specified. Exiting.")
	}

	brokerPort := os.Getenv("WINGMAN_PORT")
	if brokerPort == "" {
		log.Println("No WINGMAN_PORT provided. Defaulting to " + DefaultWingmanPort)
		brokerPort = DefaultWingmanPort
	}

	redisUser := os.Getenv("REDIS_USER")
	redisPass := os.Getenv("REDIS_PASS")
	if redisUser != "" {
		log.Println("Using REDIS_USER=" + redisUser)
	} else {
		log.Println("No REDIS_USER specified.")
	}

	if redisPass != "" {
		log.Println("Using REDIS_PASS=<redacted>")
	} else {
		log.Println("No REDIS_PASS specified.")
	}

	socketBroker := broker.NewBroker(redisAddr, redisUser, redisPass)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go socketBroker.Start(brokerPort)

	<-sc
	log.Println("Wingman received a kill/term signal and will now exit")
}
