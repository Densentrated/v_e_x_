package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/philippgille/chromem-go"

	"vex-backend/config"
	"vex-backend/handlers"
	"vex-backend/routes"
	"vex-backend/vector/embed"
	"vex-backend/vector/manager"
)

func main() {
	// Initialize config ONCE at startup
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded config - Git User: %s, Clone Folder: %s\n", config.Config.GitUser, config.Config.CloneFolder)

	// Initialize vector database and manager inline (avoid db.InitVectorDB globals)

	// Create the Voyage AI embedder
	embedder := embed.NewVoyageEmbed()
	log.Println("Initialized Voyage AI embedder")

	// Create in-memory chromem-go database
	vectorDB := chromem.NewDB()
	log.Println("Initialized chromem-go vector database")

	// Create or get the documents collection
	collection, err := vectorDB.GetOrCreateCollection("documents", nil, nil)
	if err != nil {
		log.Fatal("Failed to create/get collection:", err)
	}

	// Initialize the vector manager with the embedder
	vm := manager.NewChromeManager(collection, vectorDB, embedder)
	log.Println("Initialized chromem-go vector manager")

	// Inject the vector manager into HTTP handlers (so handlers rely on the manager + its embedder)
	handlers.SetVectorManager(vm)

	// Basic sanity check
	if vm == nil || vectorDB == nil || embedder == nil {
		log.Fatal("vector initialization failed")
	}

	mux := routes.RegisterRoutes()

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
