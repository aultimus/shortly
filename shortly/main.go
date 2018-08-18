package main

import (
	"fmt"
	"log"

	"github.com/aultimus/shortly/db"

	"net/http"
	_ "net/http/pprof"

	"github.com/aultimus/shortly"
)

func main() {
	fmt.Println("shortly started")

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	app := shortly.NewApp()
	err := app.Init(db.NewDynamoService())
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
