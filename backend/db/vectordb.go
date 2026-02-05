package db

import (
	"context"
	"fmt"
	"log"

	"github.com/philippgille/chromem-go"

	"vex-backend/vector/embed"
	"vex-backend/vector/manager"
)

var (
	// VectorDB is the global chromem-go database instance
	VectorDB *chromem.DB
	// VectorManager is the global vector manager for storing and retrieving vectors
	VectorManager manager.VectorManager
	// Embedder is the global embedder used for creating embeddings
	Embedder embed.VectorEmbed
)

// InitVectorDB initializes the vector database and manager with the given embedder
func InitVectorDB() error {
	ctx := context.Background()

	// Create the Voyage AI embedder
	Embedder = embed.NewVoyageEmbed()
	log.Println("Initialized Voyage AI embedder")

	// Create in-memory chromem-go database
	VectorDB = chromem.NewDB()
	log.Println("Initialized chromem-go vector database")

	// Create or get the documents collection
	// We pass nil for embeddingFunc since we handle embeddings ourselves via VectorEmbed
	collection, err := VectorDB.GetOrCreateCollection("documents", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create/get collection: %w", err)
	}

	// Initialize the vector manager with the embedder
	VectorManager = manager.NewChromeManager(collection, VectorDB, Embedder)
	log.Println("Initialized vector manager with Voyage AI embedder")

	// Test the connection
	if err := testVectorDB(ctx); err != nil {
		return fmt.Errorf("failed to test vector database: %w", err)
	}

	log.Println("Vector database initialized successfully")
	return nil
}

// InitVectorDBWithEmbedder initializes the vector database with a custom embedder
func InitVectorDBWithEmbedder(embedder embed.VectorEmbed) error {
	ctx := context.Background()

	// Use the provided embedder
	Embedder = embedder
	log.Println("Initialized custom embedder")

	// Create in-memory chromem-go database
	VectorDB = chromem.NewDB()
	log.Println("Initialized chromem-go vector database")

	// Create or get the documents collection
	collection, err := VectorDB.GetOrCreateCollection("documents", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create/get collection: %w", err)
	}

	// Initialize the vector manager with the embedder
	VectorManager = manager.NewChromeManager(collection, VectorDB, Embedder)
	log.Println("Initialized vector manager with custom embedder")

	// Test the connection
	if err := testVectorDB(ctx); err != nil {
		return fmt.Errorf("failed to test vector database: %w", err)
	}

	log.Println("Vector database initialized successfully")
	return nil
}

// testVectorDB performs a basic test to ensure the database is working
func testVectorDB(ctx context.Context) error {
	if VectorDB == nil {
		return fmt.Errorf("vector database is nil")
	}
	if VectorManager == nil {
		return fmt.Errorf("vector manager is nil")
	}
	if Embedder == nil {
		return fmt.Errorf("embedder is nil")
	}
	return nil
}
