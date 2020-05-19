package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	conf "github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/brentahughes/service_tester/pkg/service"
	"github.com/brentahughes/service_tester/pkg/servicecheck"
	"github.com/brentahughes/service_tester/pkg/webserver"
	"github.com/dgraph-io/badger"
)

func main() {
	c, err := conf.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := badger.Open(badger.DefaultOptions(".db"))
	if err != nil {
		log.Fatal("error opening database: ", err)
	}
	defer db.Close()

	logger := models.NewLogger(db)

	if err := models.UpdateCurrentHost(db, c); err != nil {
		log.Fatal("Error updating current host: ", err)
	}

	s := service.NewService(logger, c.ServicePort)
	go s.Start()

	checker, err := servicecheck.NewChecker(db, logger, c)
	if err != nil {
		log.Fatal(err)
	}
	go checker.Start()
	defer checker.Stop()

	server := webserver.NewServer(*c, db, logger, c.Port)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("error starting web interface", err)
		}
	}()
	defer server.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	logger.Infof("Shutdown signal received")
}
