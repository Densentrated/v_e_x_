package manager

import (
	"context"

	"vex-backend/vector"
)

// QueryResult is a database-agnostic result from a vector query
type QueryResult struct {
	ID         string
	Content    string
	Metadata   map[string]string
	Similarity float32
}

// VectorManager defines the interface for vector database operations
// This interface is database-agnostic and doesn't depend on chromem-go, Pinecone, etc.
type VectorManager interface {
	// storage methods

	// StoreVectorInDB stores a single VectorData object as a vector in the database
	StoreVectorInDB(ctx context.Context, vectorData vector.VectorData) error

	// StoreFileAsVector uses vectorembed to convert a file into vectors and stores in DB
	StoreFileAsVector(ctx context.Context, filename string) error

	// StoreFilesAsVectors uses StoreFileAsVector to store multiple files into the vector DB
	StoreFilesAsVectors(ctx context.Context, files []string) error

	// retrieval methods

	// RetrieveVectorWithMetaData retrieves vectors with specific metadata filters
	RetrieveVectorWithMetaData(ctx context.Context, metadata map[string]string) ([]QueryResult, error)

	// RetrieveVectorWithEmbedding retrieves vectors similar to a given embedding
	RetrieveVectorWithEmbedding(ctx context.Context, embedding []float32) ([]QueryResult, error)

	// RetrieveVectorWithQuery retrieves vectors using a text query (semantic search)
	RetrieveVectorWithQuery(ctx context.Context, query string, nResults int) ([]QueryResult, error)

	// db editing methods

	// DeleteVectorsWithEmbedding deletes vectors similar to a given embedding
	DeleteVectorsWithEmbedding(ctx context.Context, embedding []float32) error

	// DeleteVectorsWithMetaData deletes all vectors matching specific metadata
	DeleteVectorsWithMetaData(ctx context.Context, metadata map[string]string) error
}

// baseVectorManager provides common functionality that implementations can embed (PRIVATE)
type baseVectorManager struct {
	// Implementations can store whatever they need here
	// For chromem-go: a chromem.Collection and chromem.DB
	// For Pinecone: a client connection
	// For others: whatever makes sense
}

func newBaseVectorManager() *baseVectorManager {
	return &baseVectorManager{}
}
