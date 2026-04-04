package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	serverMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("error listening and serving lol", err)
		os.Exit(1)
	}
}
