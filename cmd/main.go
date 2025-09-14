package main

import (
	"service_orchestrator/internal/components/logger"
)

func main() {
	// create logger
	log := logger.New("./logs/orch.log", 10, 5, 30, true)

	// init logger
	log.Init(logger.LogDebug)

	// deinit logger
	log.Deinit()
}
