package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mcp/cmd"
)

func main() {
	//exit process immediately upon sigterm
	handleSigTerms()

	//run
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func handleSigTerms() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("received SIGTERM, exiting")
		os.Exit(1)
	}()
}
