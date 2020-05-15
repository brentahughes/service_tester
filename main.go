package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/brentahughes/service_tester/pkg/service"
	"github.com/brentahughes/service_tester/pkg/servicecheck"
	"github.com/brentahughes/service_tester/pkg/webserver"
)

func main() {
	c, err := config.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := storm.Open("service_test.db", storm.Batch())
	if err != nil {
		log.Fatal("error opening database: ", err)
	}
	defer db.Close()

	if err := initDatabase(db); err != nil {
		log.Fatal(err)
	}

	logger := models.NewLogger(db)

	if err := models.UpdateCurrentHost(db, c); err != nil {
		log.Fatal(err)
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

func initDatabase(db *storm.DB) error {
	dbModels := []interface{}{
		&models.Check{},
		&models.Host{},
		&models.Log{},
	}

	for _, model := range dbModels {
		if err := db.Init(model); err != nil {
			return err
		}

		if err := db.ReIndex(model); err != nil {
			return err
		}
	}

	return nil
}
