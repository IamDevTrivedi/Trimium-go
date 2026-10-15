package main

import (
	"fmt"
	"os"
	"server/config"
)

func main() {
	_, err := config.Get()

	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("HELLO MITRO")
}
