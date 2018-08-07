package main

import (
	"log"

	"github.com/aultimus/shortly"
)

func main() {
	app := shortly.NewApp()
	err := app.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
