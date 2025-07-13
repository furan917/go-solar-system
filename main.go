package main

import (
	"fmt"
	"log"

	"github.com/furan917/go-solar-system/internal/app"
)

func main() {
	solarSystem, err := app.NewSolarSystem()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🌌 Welcome to the Interactive Solar System!")
	if err := solarSystem.Run(); err != nil {
		log.Fatal(err)
	}
}
