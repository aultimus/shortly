package main

import (
	"flag"
	"log"

	"github.com/aultimus/shortly/db"
	"github.com/cocoonlife/timber"

	"net/http"
	_ "net/http/pprof"

	"github.com/aultimus/shortly"
)

// gitSHA represents the SHA that this application is built from, injected at compile time
var gitSHA string

func main() {
	timber.AddLogger(timber.ConfigLogger{
		LogWriter: new(timber.ConsoleWriter),
		Level:     timber.DEBUG,
		Formatter: timber.NewPatFormatter("[%D %T] [%L] %s %M"),
	})

	timber.Infof("shortly started. Git SHA [%s]", gitSHA)

	go func() {
		timber.Errorf(http.ListenAndServe(":6060", nil))
	}()
	portNum := flag.String("port", "8080", "specify port number")
	flag.Parse()
	app := shortly.NewApp()
	err := app.Init(db.NewDynamoService(), *portNum)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
