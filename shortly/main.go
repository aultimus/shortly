package main

import (
	"fmt"
	"log"

	"github.com/aultimus/shortly/db"

	"net/http"
	_ "net/http/pprof"

	"github.com/aultimus/shortly"
)

// gitSHA represents the SHA that this application is built from, injected at compile time
var gitSHA string

func main() {
	fmt.Printf("shortly started. Git SHA [%s]\n", gitSHA)

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
