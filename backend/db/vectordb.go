package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/philippgille/chromem-go"

	"vex-backend/config"
	"vex-backend/vector/manager"
)

var (
	// VectorDB is the global chromem-go database instance
	VectorDB *chromem.DB
	// VectorManager is the global vector manager for storing and retrieving vectors
	VectorManager manager.VectorManager
)

// VoyageEmbeddingRequest represents the request body for Voyage AI API
type VoyageEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// VoyageEmbeddingResponse represents the response from Voyage AI API
type VoyageEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// NewVoyageEmbeddingFunc creates an embedding function that uses Voyage AI API
func NewVoyageEmbeddingFunc() chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		if config.Config.VoyageAPIKey == "" {
			return nil, fmt.Errorf("VOYAGE_API_KEY is not set")
		}

		reqBody := VoyageEmbeddingRequest{
			Input: text,
			Model: "voyage-3",
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+config.Config.VoyageAPIKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to Voyage AI: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("voyage ai api returned status %d: %s", resp.StatusCode, string(data))
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var embedResp VoyageEmbeddingResponse
		if err := json.Unmarshal(data, &embedResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(embedResp.Data) == 0 {
			return nil, fmt.Errorf("no embeddings returned from Voyage AI")
		}

		return embedResp.Data[0].Embedding, nil
	}
}

// InitVectorDB initializes the vector database and manager with Voyage AI embeddings
func InitVectorDB() error {
	ctx := context.Background()

	// Create in-memory chromem-go database
	VectorDB = chromem.NewDB()
	log.Println("Initialized chromem-go vector database")

	// Create embedding function using Voyage AI
	embeddingFunc := NewVoyageEmbeddingFunc()

	// Create or get the documents collection with Voyage AI embedding function
	collection, err := VectorDB.GetOrCreateCollection("documents", nil, embeddingFunc)
	if err != nil {
		return fmt.Errorf("failed to create/get collection: %w", err)
	}

	// Initialize the vector manager
	VectorManager = manager.NewChromeManager(collection, VectorDB)
	log.Println("Initialized vector manager with Voyage AI embeddings")

	// Test the connection
	if err := testVectorDB(ctx); err != nil {
		return fmt.Errorf("failed to test vector database: %w", err)
	}

	log.Println("Vector database initialized successfully")
	return nil
}

// testVectorDB performs a basic test to ensure the database is working
func testVectorDB(ctx context.Context) error {
	// Just verify we can create a collection
	if VectorDB == nil {
		return fmt.Errorf("vector database is nil")
	}
	if VectorManager == nil {
		return fmt.Errorf("vector manager is nil")
	}
	return nil
}
