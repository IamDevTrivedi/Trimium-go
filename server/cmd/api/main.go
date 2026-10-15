package main

import (
	"fmt"
	"os"
	"server/config"
	"server/logger"
)

func main() {
	_, err := config.Get()

	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	logger := logger.Get()
	logger.Info("Environment is valid")
}
