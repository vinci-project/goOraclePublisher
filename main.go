package main

import (
	"goOraclePublisher/publisher"
	"log"
	"net/http"
)

func main() {
	//

	p := publisher.NewPublisher()
	go p.Listen()
	http.HandleFunc("/ws/oracle", p.Serve)
	http.HandleFunc("/status", p.Status)
	log.Fatal(http.ListenAndServe(":5050", nil))
}
