package main

import (
	"fmt"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
	"github.com/AlexandrKudryavtsev/go-load-balancer/internal/app"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yaml")

	if err != nil {
		fmt.Println("failed load config")
		return
	}

	if err = cfg.Validate(); err != nil {
		fmt.Printf("invalid config: %v", err)
		return
	}

	app.Run(cfg)
}
