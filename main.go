package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/brentahughes/service_tester/pkg/servicecheck"
	"github.com/brentahughes/service_tester/pkg/webserver"
)

func main() {
	c, err := config.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := storm.Open("service_test.db")
	if err != nil {
		log.Fatal("error opening database: ", err)
	}
	defer db.Close()

	logger := models.NewLogger(db)

	if err := models.UpdateCurrentHost(db, c); err != nil {
		log.Fatal(err)
	}

	checker := servicecheck.NewChecker(db, c.Discovery, c.Port, logger, c.CheckInterval)
	go checker.Start()
	defer checker.Stop()

	server := webserver.NewServer(*c, db, logger, c.Port)
	go server.Start()
	defer server.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	logger.Infof("Shutdown signal received")
}
