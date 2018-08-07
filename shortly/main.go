package main

import (
	"log"

	"github.com/aultimus/shortly/db"

	"github.com/aultimus/shortly"
)

func main() {
	app := shortly.NewApp()
	err := app.Init(db.NewMapDB())
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
