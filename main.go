package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"crons/internal"

	"github.com/d2jvkpn/go-web/pkg/misc"
)

func main() {
	var (
		release bool
		addr    string
		config  string
		err     error
	)

	flag.StringVar(&config, "config", "local.yaml", "tasks config file")
	flag.StringVar(&addr, "addr", "", "http serve address")
	flag.BoolVar(&release, "release", false, "run in release mode")

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(), "%s\n\nUsage of %s:\n",
			buildInfo(),
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if addr != "" {
		err = server(config, addr, release)
	} else {
		err = runCrons(config)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func buildInfo() string {
	data := misc.BuildInfo()
	strs := make([]string, 0, len(data))
	for k, v := range data {
		strs = append(strs, fmt.Sprintf("%s: %s", strings.Title(k), v))
	}

	sort.Strings(strs)
	return strings.Join(strs, "\n")
}

func runCrons(config string) (err error) {
	var num int

	if num, err = internal.LoadCron(config, "jobs"); err != nil {
		return err
	}

	internal.Manager.Start()
	log.Printf(">>> Number Of cron tasks: %d\n", num)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		fmt.Println("")
		log.Println("received signal:", sig)

		internal.Manager.Shutdown()
		log.Println("<<< Stop Cron")
	}

	return err
}

func server(config, addr string, release bool) (err error) {
	var num int

	parameters := map[string]any{
		"--config":  config,
		"--addr":    addr,
		"--release": release,
	}
	for k, v := range misc.BuildInfo() {
		parameters[k] = v
	}

	if num, err = internal.LoadCron(config, "jobs"); err != nil {
		return err
	}
	if err = internal.Load(release); err != nil {
		return err
	}

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)

	internal.Manager.Start()

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

		internal.Manager.Shutdown()
		log.Println("<<< Stop Cron")
		internal.Shutdown()
		err = <-errch
	}

	return err
}
