package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"service_orchestrator/internal/components/configuration"
	"service_orchestrator/internal/components/globals"
	"service_orchestrator/internal/components/logger"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// init global atomic
	globals.ProcessRunning.Store(true)

	// get bin path
	binPath := os.Args[0]

	globals.BinPath = filepath.Dir(binPath)
	globals.BinPath = filepath.Dir(globals.BinPath)

	// create logger
	loggerPath := filepath.Join(globals.BinPath, "/orchestrator_assets/logs/orch.log")
	log := logger.New(loggerPath, 10, 5, 30, true)

	// init logger
	globals.LogLevel = logger.LogInfo
	log.Init()
	defer log.Deinit()

	// read config
	_, err := configuration.ReadConfiguration(log)
	if err != nil {
		slog.Error("Bad configuration")
		return
	}

	// handle args
	slog.Debug("CMD", "value", fmt.Sprint(os.Args))

	// get mode
	mode := "unknown"
	if len(os.Args) > 1 {
		mode = os.Args[1]
		slog.Debug("Running mode", "mode", mode)
	} else {
		PrintInfo()
		return
	}

	switch mode {
	case "console":
		runConsole()
	case "daemon":
		runDaemon()
	default:
		PrintInfo()
	}

	globals.ProcessRunning.Store(false)

}

func PrintInfo() {
	fmt.Println("You should to set mode:")
	fmt.Println("	console - run in console")
	fmt.Println("	daemon - run as service via systemd")
}

func runConsole() {
	slog.Info("Press ESC to stop...")
	// turn on raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	for globals.ProcessRunning.Load() {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		if buf[0] == 27 { // ESC
			slog.Info("ESC was pressed, stopping...")
			globals.ProcessRunning.Store(false)
			break
		}
	}
}

func runDaemon() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Daemon mode start...")

	<-stop
	slog.Info("Got stop signal, stopping...")
	globals.ProcessRunning.Store(false)
}
