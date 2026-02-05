package manager

import (
	"context"

	"vex-backend/vector"
	"vex-backend/vector/embed"
)

// QueryResult is a database-agnostic result from a vector query
type QueryResult struct {
	ID         string
	Content    string
	Metadata   map[string]string
	Similarity float32
	Embedding  []float32
}

// VectorManager defines the interface for vector database operations
// This interface is database-agnostic and doesn't depend on chromem-go, Pinecone, etc.
type VectorManager interface {
	// GetEmbedder returns the embedder used by this manager
	GetEmbedder() embed.VectorEmbed

	// SetEmbedder sets the embedder to use for embedding operations
	SetEmbedder(embedder embed.VectorEmbed)

	// storage methods

	// StoreVectorInDB stores a single VectorData object as a vector in the database
	StoreVectorInDB(ctx context.Context, vectorData vector.VectorData) error

	// StoreFileAsVector uses the embedder to convert a file into vectors and stores in DB
	StoreFileAsVector(ctx context.Context, filename string, metadata map[string]string) error

	// StoreFilesAsVectors uses StoreFileAsVector to store multiple files into the vector DB
	StoreFilesAsVectors(ctx context.Context, files []string, metadata map[string]string) error

	// retrieval methods

	// RetrieveVectorWithMetaData retrieves vectors with specific metadata filters
	RetrieveVectorWithMetaData(ctx context.Context, metadata map[string]string) ([]QueryResult, error)

	// RetrieveVectorWithEmbedding retrieves vectors similar to a given embedding
	RetrieveVectorWithEmbedding(ctx context.Context, embedding []float32, nResults int) ([]QueryResult, error)

	// RetrieveVectorWithQuery retrieves vectors using a text query (semantic search)
	// Uses the embedder to convert the query text into an embedding
	RetrieveVectorWithQuery(ctx context.Context, query string, nResults int) ([]QueryResult, error)

	// db editing methods

	// DeleteVectorsWithIDs deletes vectors with the specified IDs
	DeleteVectorsWithIDs(ctx context.Context, ids []string) error

	// DeleteVectorsWithMetaData deletes all vectors matching specific metadata
	DeleteVectorsWithMetaData(ctx context.Context, metadata map[string]string) error
}

// baseVectorManager provides common functionality that implementations can embed (PRIVATE)
type baseVectorManager struct {
	embedder embed.VectorEmbed
}

// newBaseVectorManager creates a base vector manager
func newBaseVectorManager(embedder embed.VectorEmbed) *baseVectorManager {
	return &baseVectorManager{
		embedder: embedder,
	}
}

// GetEmbedder returns the embedder used by this manager
func (b *baseVectorManager) GetEmbedder() embed.VectorEmbed {
	return b.embedder
}

// SetEmbedder sets the embedder to use for embedding operations
func (b *baseVectorManager) SetEmbedder(embedder embed.VectorEmbed) {
	b.embedder = embedder
}
