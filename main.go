package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crons/internal"

	"github.com/d2jvkpn/go-web/pkg/misc"
)

func main() {
	var (
		num     int
		release bool
		addr    string
		config  string
		err     error
	)

	flag.StringVar(&config, "config", "configs/test.yaml", "tasks config file")
	flag.StringVar(&addr, "addr", ":3011", "http serve address")
	flag.BoolVar(&release, "release", false, "run in release mode")
	flag.Parse()

	okOrExit := func(err error) {
		if err == nil {
			return
		}
		log.Fatalln(err)
	}

	// err = os.Chdir(filepath.Dir(os.Args[0]))
	// okOrExit(err)
	parameters := map[string]any{
		"--config":  config,
		"--addr":    addr,
		"--release": release,
	}
	for k, v := range misc.BuildInfo() {
		parameters[k] = v
	}

	num, err = internal.LoadCron(config, "jobs")
	okOrExit(err)
	err = internal.Load(release)
	okOrExit(err)

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)

	go func() {
		var err error
		log.Printf(">>> HTTP server listening on %s, number Of cron tasks: %d\n", addr, num)
		err = internal.Serve(addr, parameters)
		errch <- err
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-errch:
	case sig := <-quit:
		fmt.Println("")
		log.Println("received signal:", sig)
		internal.Shutdown()
		err = <-errch
		log.Println("<<< Stop Cron")
	}

	if err != nil {
		log.Fatalln(err)
	}
}
