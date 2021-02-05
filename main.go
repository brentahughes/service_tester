package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	conf "github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/brentahughes/service_tester/pkg/service"
	"github.com/brentahughes/service_tester/pkg/servicecheck"
	"github.com/brentahughes/service_tester/pkg/webserver"
	"github.com/dgraph-io/badger"
)

var (
	cpuProfile = flag.Bool("prof.cpu", false, "")
	memProfile = flag.Bool("prof.mem", false, "")
)

func main() {
	flag.Parse()

	if *cpuProfile {
		cpuF, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpuF)
		defer pprof.StopCPUProfile()
	}

	c, err := conf.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	opts := badger.DefaultOptions(".db").WithSyncWrites(false)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal("error opening database: ", err)
	}
	defer db.Close()

	go keepCurrentHostUpdated(db, c)

	s := service.NewService(c.ServicePort)
	go s.Start()

	checker, err := servicecheck.NewChecker(db, c)
	if err != nil {
		log.Fatal(err)
	}
	go checker.Start()
	defer checker.Stop()

	server := webserver.NewServer(*c, db, c.Port)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("error starting web interface", err)
		}
	}()
	defer server.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("Shutdown signal received")

	if *memProfile {
		memF, err := os.Create("mem.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(memF)
		memF.Close()
	}
}

func keepCurrentHostUpdated(db *badger.DB, c *conf.Config) {
	if err := models.UpdateCurrentHost(db, c, true); err != nil {
		log.Fatal("Error updating current host: ", err)
	}

	t := time.NewTicker(time.Hour)
	for range t.C {
		if err := models.UpdateCurrentHost(db, c, false); err != nil {
			log.Println("Error updating current host: ", err)
		}
	}
}
