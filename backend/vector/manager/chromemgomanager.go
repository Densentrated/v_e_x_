package manager

import (
	"context"
	"fmt"

	"github.com/philippgille/chromem-go"

	"vex-backend/vector"
	"vex-backend/vector/embed"
)

// ChromaManager implements VectorManager using chromem-go
type ChromaManager struct {
	*baseVectorManager
	collection *chromem.Collection
	db         *chromem.DB
}

// NewChromeManager creates a new ChromaManager with the given embedder
func NewChromeManager(collection *chromem.Collection, db *chromem.DB, embedder embed.VectorEmbed) *ChromaManager {
	return &ChromaManager{
		baseVectorManager: newBaseVectorManager(embedder),
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
		Embedding: vectorData.Embedding, // Can be nil if embedder will generate it
	}

	// Add document to collection with single thread (for single document)
	err := cm.collection.AddDocuments(ctx, []chromem.Document{doc}, 1)
	if err != nil {
		return fmt.Errorf("failed to store vector in database: %w", err)
	}

	return nil
}

// StoreFileAsVector uses the embedder to convert a file into vectors and stores in DB
func (cm *ChromaManager) StoreFileAsVector(ctx context.Context, filename string, metadata map[string]string) error {
	if cm.embedder == nil {
		return fmt.Errorf("embedder is not set")
	}

	// Use the embedder to embed the entire file
	vectorDataList, err := cm.embedder.EmbedFile(ctx, filename, metadata)
	if err != nil {
		return fmt.Errorf("failed to embed file %s: %w", filename, err)
	}

	// Store each vector in the database
	for i, vectorData := range vectorDataList {
		if err := cm.StoreVectorInDB(ctx, vectorData); err != nil {
			return fmt.Errorf("failed to store chunk %d from file %s: %w", i, filename, err)
		}
	}

	return nil
}

// StoreFilesAsVectors stores multiple files as vectors
func (cm *ChromaManager) StoreFilesAsVectors(ctx context.Context, files []string, metadata map[string]string) error {
	for _, file := range files {
		// Create file-specific metadata
		fileMetadata := make(map[string]string)
		for k, v := range metadata {
			fileMetadata[k] = v
		}
		fileMetadata["filename"] = file

		if err := cm.StoreFileAsVector(ctx, file, fileMetadata); err != nil {
			return fmt.Errorf("failed to store file %s: %w", file, err)
		}
	}
	return nil
}

// RetrieveVectorWithMetaData retrieves vectors with specific metadata filters
func (cm *ChromaManager) RetrieveVectorWithMetaData(ctx context.Context, metadata map[string]string) ([]QueryResult, error) {
	// chromem-go doesn't have a direct "get by metadata" method, so we use Query with a dummy query
	// We'll retrieve all and filter, or use the where clause in Query
	results, err := cm.collection.Query(ctx, "", 1000, metadata, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve by metadata: %w", err)
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:         r.ID,
			Content:    r.Content,
			Metadata:   r.Metadata,
			Embedding:  r.Embedding,
			Similarity: r.Similarity,
		}
	}

	return queryResults, nil
}

// RetrieveVectorWithEmbedding retrieves vectors similar to a given embedding
func (cm *ChromaManager) RetrieveVectorWithEmbedding(ctx context.Context, embedding []float32, nResults int) ([]QueryResult, error) {
	if nResults <= 0 {
		nResults = 10
	}

	// Use QueryEmbedding for embedding-based queries
	results, err := cm.collection.QueryEmbedding(ctx, embedding, nResults, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query by embedding: %w", err)
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:         r.ID,
			Content:    r.Content,
			Metadata:   r.Metadata,
			Embedding:  r.Embedding,
			Similarity: r.Similarity,
		}
	}

	return queryResults, nil
}

// RetrieveVectorWithQuery retrieves vectors using a text query (semantic search)
// Uses the embedder to convert the query text into an embedding
func (cm *ChromaManager) RetrieveVectorWithQuery(ctx context.Context, query string, nResults int) ([]QueryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if nResults <= 0 {
		nResults = 10
	}

	if cm.embedder == nil {
		return nil, fmt.Errorf("embedder is not set, cannot embed query")
	}

	// Use the embedder to convert the query to an embedding
	queryEmbedding, err := cm.embedder.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Use the embedding to query the database
	return cm.RetrieveVectorWithEmbedding(ctx, queryEmbedding, nResults)
}

// DeleteVectorsWithIDs deletes vectors with the specified IDs
func (cm *ChromaManager) DeleteVectorsWithIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// chromem-go's Delete signature: Delete(_ context.Context, where, whereDocument map[string]string, ids ...string) error
	err := cm.collection.Delete(ctx, nil, nil, ids...)
	if err != nil {
		return fmt.Errorf("failed to delete vectors by IDs: %w", err)
	}

	return nil
}

// DeleteVectorsWithMetaData deletes all vectors matching specific metadata
func (cm *ChromaManager) DeleteVectorsWithMetaData(ctx context.Context, metadata map[string]string) error {
	if len(metadata) == 0 {
		return fmt.Errorf("metadata filter cannot be empty")
	}

	// chromem-go's Delete can filter by metadata using the "where" parameter
	err := cm.collection.Delete(ctx, metadata, nil)
	if err != nil {
		return fmt.Errorf("failed to delete vectors by metadata: %w", err)
	}

	return nil
}

// Ensure ChromaManager implements VectorManager interface
var _ VectorManager = (*ChromaManager)(nil)
