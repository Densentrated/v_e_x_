package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"vex-backend/config"
	"vex-backend/routes"
	"vex-backend/vector/embed"
	vectormgr "vex-backend/vector/manager"
)

func main() {
	// Initialize config ONCE at startup
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded config - Git User: %s, Clone Folder: %s\n", config.Config.GitUser, config.Config.CloneFolder)

	embedder := embed.NewVoyageEmbed("voyage-4-large")
	manager := vectormgr.NewChromemManager(embedder)

	mux := routes.RegisterRoutes(manager)

	port := config.Config.ServerPort
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = ":" + port
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] Server starting on port %s\n", currentTime, port)
	log.Fatal(http.ListenAndServe(port, mux))
}
