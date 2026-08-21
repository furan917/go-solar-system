package main

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/furan917/go-solar-system/internal/app"
)

func main() {
	bundled, err := fs.Sub(bundledSystems, "systems")
	if err != nil {
		log.Fatal(err)
	}

	solarSystem, err := app.NewSolarSystem(bundled)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🌌 Welcome to the Interactive Solar System!")
	if err := solarSystem.Run(); err != nil {
		log.Fatal(err)
	}
}
