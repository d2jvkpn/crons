package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"crons/models"

	"github.com/d2jvkpn/go-web/pkg/wrap"
)

func main() {
	var (
		config   string
		logLevel string
		num      int
		dryrun   bool
		err      error
		logger   *wrap.Logger
		manager  *models.Manager
	)

	flag.StringVar(&config, "config", "configs/test.yaml", "tasks config file")
	flag.BoolVar(&dryrun, "dryrun", false, "dry run")
	flag.StringVar(&logLevel, "logLevel", "info", "log level")
	flag.Parse()

	okOrExit := func(err error) {
		logger.Down()
		if err == nil {
			return
		}
		fmt.Println(err)
		os.Exit(1)
	}

	// err = os.Chdir(filepath.Dir(os.Args[0]))
	// okOrExit(err)

	logger = wrap.NewLogger("logs/crons.log", wrap.LogLevelFromStr(logLevel), 256, nil)

	manager = models.NewManager(logger.Named("manager"))
	num, err = manager.LoadTasksFronConfig(config, "jobs")
	okOrExit(err)

	fmt.Printf(">>> Start Cron: pid=%d, numberOfTasks=%d, dryrun=%t\n", manager.Pid, num, dryrun)

	if !dryrun {
		manager.Start()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		fmt.Println("")
		manager.Shutdown()
		fmt.Println("<<< Stop Cron")
	}
}
