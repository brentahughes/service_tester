package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/db"
	"github.com/brentahughes/service_tester/pkg/servicecheck"
	"github.com/brentahughes/service_tester/pkg/webserver"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "service_tester",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.LoadEnvConfig()
		if err != nil {
			log.Fatal(err)
		}

		db := db.NewInMemoryStore()

		checker := servicecheck.NewChecker(db, c.Discovery, c.CheckInterval)
		go checker.Start()
		defer checker.Stop()

		server := webserver.NewServer(db, c.Port)
		go server.Start()
		defer server.Stop()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
