package manager

import (
	"context"
	"fmt"

	"vex-backend/vector"

	"github.com/philippgille/chromem-go"
)

// ChromaManager implements VectorManager using chromem-go
type ChromaManager struct {
	*baseVectorManager
	collection *chromem.Collection
	db         *chromem.DB
}

// NewChromeManager creates a new ChromaManager (PUBLIC)
func NewChromeManager(collection *chromem.Collection, db *chromem.DB) *ChromaManager {
	return &ChromaManager{
		baseVectorManager: newBaseVectorManager(),
		collection:        collection,
		db:                db,
	}
}

// StoreVectorInDB stores a single VectorData object as a vector in the database
func (cm *ChromaManager) StoreVectorInDB(ctx context.Context, vectorData vector.VectorData) error {
	// Validate input
	if vectorData.ID == "" {
		return fmt.Errorf("vector ID cannot be empty")
	}
	if vectorData.Data == "" {
		return fmt.Errorf("vector data/content cannot be empty")
	}

	// Create a chromem document from the VectorData
	doc := chromem.Document{
		ID:        vectorData.ID,
		Content:   vectorData.Data,
		Metadata:  vectorData.MetaData,
		Embedding: vectorData.Embedding, // Can be nil, chromem-go will generate it
	}

	// Add document to collection with single thread (for single document)
	err := cm.collection.AddDocuments(ctx, []chromem.Document{doc}, 1)
	if err != nil {
		return fmt.Errorf("failed to store vector in database: %w", err)
	}

	return nil
}

// StoreFileAsVector uses vectorembed to convert a file into vectors and stores in DB
func (cm *ChromaManager) StoreFileAsVector(ctx context.Context, filename string) error {
	// TODO: Implement using your embedding logic from voyageembed.go
	// 1. Read file
	// 2. Create chunks using vectorEmbed interface
	// 3. Embed each chunk
	// 4. Store using StoreVectorInDB
	return fmt.Errorf("not implemented")
}

// StoreFilesAsVectors stores multiple files as vectors
func (cm *ChromaManager) StoreFilesAsVectors(ctx context.Context, files []string) error {
	for _, file := range files {
		if err := cm.StoreFileAsVector(ctx, file); err != nil {
			return fmt.Errorf("failed to store file %s: %w", file, err)
		}
	}
	return nil
}

// RetrieveVectorWithMetaData retrieves vectors with specific metadata filters
func (cm *ChromaManager) RetrieveVectorWithMetaData(ctx context.Context, metadata map[string]string) ([]QueryResult, error) {
	results, err := cm.collection.Query(ctx, "", 1000, metadata, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve by metadata: %w", err)
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:       r.ID,
			Content:  r.Content,
			Metadata: r.Metadata,
		}
	}

	return queryResults, nil
}

// RetrieveVectorWithEmbedding retrieves vectors similar to a given embedding
func (cm *ChromaManager) RetrieveVectorWithEmbedding(ctx context.Context, embedding []float32) ([]QueryResult, error) {
	// Use QueryEmbedding for embedding-based queries
	results, err := cm.collection.QueryEmbedding(ctx, embedding, 10, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query by embedding: %w", err)
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:         r.ID,
			Content:    r.Content,
			Metadata:   r.Metadata,
			Similarity: r.Similarity,
		}
	}

	return queryResults, nil
}

// RetrieveVectorWithQuery retrieves vectors using a text query (semantic search)
func (cm *ChromaManager) RetrieveVectorWithQuery(ctx context.Context, query string, nResults int) ([]QueryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if nResults <= 0 {
		return nil, fmt.Errorf("nResults must be greater than 0")
	}

	results, err := cm.collection.Query(ctx, query, nResults, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:         r.ID,
			Content:    r.Content,
			Metadata:   r.Metadata,
			Similarity: r.Similarity,
		}
	}

	return queryResults, nil
}

// DeleteVectorsWithEmbedding deletes vectors similar to a given embedding
func (cm *ChromaManager) DeleteVectorsWithEmbedding(ctx context.Context, embedding []float32) error {
	return fmt.Errorf("not implemented for chromem-go: use DeleteVectorsWithMetaData instead")
}

// DeleteVectorsWithMetaData deletes all vectors matching specific metadata
func (cm *ChromaManager) DeleteVectorsWithMetaData(ctx context.Context, metadata map[string]string) error {
	if len(metadata) == 0 {
		return fmt.Errorf("metadata filter cannot be empty")
	}

	// Delete documents with the matching metadata
	// chromem-go's Delete signature: Delete(_ context.Context, where, whereDocument map[string]string, ids ...string) error
	err := cm.collection.Delete(ctx, metadata, nil)
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	return nil
}
