package manager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/philippgille/chromem-go"

	"vex-backend/vector"
	"vex-backend/vector/embed"
)

// ChromaManager implements VectorManager using chromem-go
type ChromaManager struct {
	*baseVectorManager
	collection *chromem.Collection
	db         *chromem.DB

	// docCount tracks the approximate number of documents in the collection.
	// We maintain it atomically to avoid querying the collection count on every request.
	docCount int64
}

// NewChromeManager creates a new ChromaManager with the given embedder
func NewChromeManager(collection *chromem.Collection, db *chromem.DB, embedder embed.VectorEmbed) *ChromaManager {
	cm := &ChromaManager{
		baseVectorManager: newBaseVectorManager(embedder),
		collection:        collection,
		db:                db,
		docCount:          0,
	}

	// Try to initialize an approximate document count; failure is non-fatal.
	// Use a background context for initialization.
	ctx := context.Background()
	if docs, err := collection.Query(ctx, "", 100000, nil, nil); err == nil {
		atomic.StoreInt64(&cm.docCount, int64(len(docs)))
	} else {
		// If probe fails, leave docCount at 0 and allow runtime adjustments on adds/deletes.
		log.Printf("warning: could not initialize collection docCount: %v", err)
	}

	return cm
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

	// Increment our atomic document count to reflect the newly added document.
	// Use a conservative +1 because AddDocuments succeeded for one document.
	atomic.AddInt64(&cm.docCount, 1)

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
// Simplified behavior: try calling QueryEmbedding starting at the requested nResults
// and decrement down to 1 until a call succeeds. This avoids relying on collection.Query
// probes and prevents propagating the chromem-go "nResults must be <= the number of documents"
// error to callers when fewer documents exist than requested.
// RetrieveVectorWithEmbedding retrieves vectors similar to a given embedding
// Strategy: call QueryEmbedding starting at the requested nResults and decrement until
// it succeeds or reaches 0. If chromem-go reports an nResults/collection-size error,
// return an empty result set silently (no logging) so callers don't receive the raw chromem error.
func (cm *ChromaManager) RetrieveVectorWithEmbedding(ctx context.Context, embedding []float32, nResults int) ([]QueryResult, error) {
	if nResults <= 0 {
		nResults = 10
	}

	// Log the tracked document count and requested nResults for debugging.
	currentCount := int(atomic.LoadInt64(&cm.docCount))
	log.Printf("chroma manager: RetrieveVectorWithEmbedding called with requested nResults=%d, tracked docCount=%d", nResults, currentCount)

	// If we have no documents tracked, return empty results immediately.
	if currentCount == 0 {
		return []QueryResult{}, nil
	}
	// Clamp nResults to our tracked count to avoid chromem-go errors.
	if nResults > currentCount {
		nResults = currentCount
	}

	var lastErr error

	// Try QueryEmbedding from requested nResults down to 1 until a call succeeds.
	for trial := nResults; trial >= 1; trial-- {
		results, err := cm.collection.QueryEmbedding(ctx, embedding, trial, nil, nil)
		if err == nil {
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

		lastErr = err

		// If the error indicates the requested nResults is larger than the collection,
		// treat that as "no results" and return an empty slice silently (no chromem error).
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "nResults") || strings.Contains(msg, "number of documents") {
				// Silent empty result to hide chromem-go constraint details from callers.
				return []QueryResult{}, nil
			}
		}

		// Otherwise continue trying smaller sizes. We intentionally avoid logging the specific
		// chromem nResults error here to keep behavior quiet as requested.
	}

	// If all attempts failed for other reasons, return the last observed error.
	if lastErr != nil {
		return nil, fmt.Errorf("failed to query by embedding after retries: %w", lastErr)
	}

	// No attempts succeeded and no error to return: return empty results.
	return []QueryResult{}, nil
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

	// Decrement our atomic docCount by the number of IDs requested for deletion.
	// If we end up negative, clamp to zero.
	if len(ids) > 0 {
		newCount := atomic.AddInt64(&cm.docCount, -int64(len(ids)))
		if newCount < 0 {
			atomic.StoreInt64(&cm.docCount, 0)
		}
	}

	return nil
}

// DeleteVectorsWithMetaData deletes all vectors matching specific metadata
func (cm *ChromaManager) DeleteVectorsWithMetaData(ctx context.Context, metadata map[string]string) error {
	if len(metadata) == 0 {
		return fmt.Errorf("metadata filter cannot be empty")
	}

	// chromem-go's Delete can filter by metadata using the "where" parameter
	// Before deleting by metadata, attempt to count matching documents so we can adjust docCount.
	// Query up to a reasonable cap to determine how many docs will be deleted.
	toDeleteCount := 0
	if docs, qErr := cm.collection.Query(ctx, "", 10000, metadata, nil); qErr == nil {
		toDeleteCount = len(docs)
	}

	err := cm.collection.Delete(ctx, metadata, nil)
	if err != nil {
		return fmt.Errorf("failed to delete vectors by metadata: %w", err)
	}

	// Adjust the atomic docCount by the deleted count. Clamp to zero if necessary.
	if toDeleteCount > 0 {
		newCount := atomic.AddInt64(&cm.docCount, -int64(toDeleteCount))
		if newCount < 0 {
			atomic.StoreInt64(&cm.docCount, 0)
		}
	}

	return nil
}

// GetDocCount returns the manager's tracked document count (atomic).
func (cm *ChromaManager) GetDocCount() int64 {
	return atomic.LoadInt64(&cm.docCount)
}

// Ensure ChromaManager implements VectorManager interface
var _ VectorManager = (*ChromaManager)(nil)
