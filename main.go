package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"crons/internal"

	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
)

var (
	//go:embed project.yaml
	_Project []byte
)

func main() {
	var (
		release bool
		addr    string
		config  string
		err     error
		project *viper.Viper
	)

	if project, err = wrap.ConfigFromBytes(_Project, "yaml"); err != nil {
		log.Fatalln(err)
	}

	flag.StringVar(&config, "config", "configs/local.yaml", "tasks config file")
	flag.StringVar(&addr, "addr", "", "http serve address")
	flag.BoolVar(&release, "release", false, "run in release mode, work with -addr")

	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(
			output,
			"Registry: %s\nVersion: %s\n%s\n\nUsage of %s:\n",
			project.GetString("project"),
			project.GetString("version"),
			misc.BuildInfoText(),
			filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()

		fmt.Fprintf(output, "\nConfig template:\n```yaml\n%s```\n", project.GetString("config"))
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

func runCrons(config string) (err error) {
	var num int

	if num, err = internal.LoadCron(config, "jobs"); err != nil {
		return err
	}

	internal.Manager.Start()
	log.Printf(">>> Number Of cron tasks: %d, Pid: %d\n", num, os.Getpid())

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
		"-config":  config,
		"-addr":    addr,
		"-release": release,
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
		log.Printf(
			">>> HTTP server listening on %s, number Of cron tasks: %d, Pid: %d\n",
			addr, num, os.Getpid(),
		)
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
		log.Println("<<< Exit")
		err = <-errch
	}

	return err
}
