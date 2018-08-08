package main

import (
	"fmt"
	"log"

	"github.com/aultimus/shortly/db"

	"github.com/aultimus/shortly"
)

func main() {
	fmt.Println("shortly started")
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
