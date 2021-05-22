package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/lyulka/rbdns/server/server"
)

func main() {
	fmt.Println("RBDNS: Server starting")

	debugMode := false
	if len(os.Args) == 2 && os.Args[1] == "debug" {
		debugMode = true
		fmt.Println("Debug mode enabled")
	}

	server := server.New(debugMode)

	// Teardown server when receive an interrupt
	// signal (for example, when user inputs Ctrl+C
	// in terminal)
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		server.Teardown()

		os.Exit(0)
	}()

	if err := http.ListenAndServe("localhost:8080", server.Router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
