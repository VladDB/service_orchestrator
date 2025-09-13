package main

import (
	"service_orhestrator/internal/logger"
)

func main() {
	// create logger
	log := logger.New("./logs/orch.log", 10, 5, 30, true)

	// init logger
	log.Init(logger.LogDebug)

	// deinit logger
	log.Deinit()
}
