package main

import (
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"log"
)

func main() {
	log.Print("main")
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("config loaded")
	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(a.Run())
}
