package main

//go:generate bash scripts/go_build.sh

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/d2jvkpn/crons/internal"

	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	//go:embed project.yaml
	_Project    []byte
	_NotWindows bool
)

func init() {
	misc.RegisterLogPrinter()
	_NotWindows = runtime.GOOS != "windows"
}

func main() {
	var (
		release bool
		go2self bool
		addr    string
		config  string
		err     error
		project *viper.Viper
	)

	if project, err = wrap.ConfigFromBytes(_Project, "yaml"); err != nil {
		log.Fatalln(err)
	}
	meta := misc.BuildInfo()
	meta["project"] = project.GetString("project")
	meta["version"] = project.GetString("version")

	flag.StringVar(&config, "config", "configs/local.yaml", "tasks config file")
	flag.StringVar(&addr, "addr", "", "http serve address")
	flag.BoolVar(&go2self, "go2self", false, "change to directory of program")
	flag.BoolVar(&release, "release", false, "run in release mode, work with -addr")

	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(
			output, "%s\n\nUsage of %s:\n",
			misc.BuildInfoText(meta), filepath.Base(os.Args[0]),
		)

		flag.PrintDefaults()

		fmt.Fprintf(output, "\nConfig template:\n```yaml\n%s```\n", project.GetString("config"))
	}
	flag.Parse()

	if go2self {
		p := filepath.Dir(os.Args[0])
		if err = os.Chdir(p); err != nil {
			log.Fatalln(err)
		}
	}

	if meta["working_dir"], err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	meta["config"] = config
	meta["addr"] = addr
	meta["release"] = release
	meta["pid"] = os.Getpid()

	level := wrap.LogLevelFromStr("debug")
	if release {
		level = wrap.LogLevelFromStr("info")
	}
	internal.Logger = wrap.NewLogger("logs/crons.log", level, 256, nil)

	if addr != "" {
		err = runServer(config, addr, release, meta)
	} else {
		err = runCrons(config, meta)
	}

	internal.Logger.Down()
	if _NotWindows {
		if err != nil {
			log.Fatalln(err)
		} else {
			log.Println("<<< Exit")
		}
	}
}

func runCrons(config string, meta map[string]any) (err error) {
	var num int

	if num, err = internal.LoadCron(config, "jobs", meta); err != nil {
		return err
	}

	internal.Manager.Start()
	log.Printf(">>> Number Of cron tasks: %d, Config: %s, Pid: %v\n", num, config, meta["pid"])

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		if sig == os.Interrupt && _NotWindows {
			fmt.Println("... received:", sig)
		}
		internal.Manager.Shutdown()
	}

	return err
}

func runServer(config, addr string, release bool, meta map[string]any) (err error) {
	var num int

	if num, err = internal.LoadCron(config, "jobs", meta); err != nil {
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
			">>> HTTP server listening on %s, Number Of cron tasks: %d, Config: %s, Pid: %v\n",
			addr, num, config, meta["pid"],
		)
		err = internal.Serve(addr, meta)
		errch <- err
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-errch:
	case sig := <-quit:
		if sig == os.Interrupt && _NotWindows {
			fmt.Println("")
		}
		internal.Logger.Warn("received signal", zap.Any("signal", sig))

		internal.Shutdown()
		err = <-errch
	}

	internal.Manager.Shutdown()

	return err
}
