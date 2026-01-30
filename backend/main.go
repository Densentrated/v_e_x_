package main

import (
	"fmt"
	"log"
	"net/http"

	"vex-backend/routes"
)

func main() {
	mux := routes.RegisterRoutes()

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
