package main

import (
	"log"
	"net/http"
)

func startServer() {
	log.Print(http.ListenAndServe(":8080", http.FileServer(http.Dir("wwwroot"))))
}
