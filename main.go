package main

import (
	"log"
	"server/server"
)

func main() {
	addr := "localhost:8090"
	s := &server.Server{
		Addr:    addr,
		DocRoot: "server/static-files",
	}
	log.Fatal(s.ListenAndServe())
}