package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"vex-backend/db"
	"vex-backend/vector"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Files to process
	files := []string{
		"/home/dense/Projects/Gitea/techronomicon/Academia/Current Issues in Cities and Suburbs/Urban Segregation.md",
		"/home/dense/Projects/Gitea/techronomicon/Analysis of Random Phenomena - Partitions and Total Probability.md",
	}

	// Step 1: Embed files and store in vector database using VectorManager
	log.Println("Starting file embedding and storage process...")
	var buf bytes.Buffer
	buf.WriteString("Step 1: Embedding files and storing in database...\n\n")

	baseMetadata := map[string]string{
		"source": "academic_notes",
	}

	// Use StoreFilesAsVectors which handles chunking, embedding, and storage
	err := db.VectorManager.StoreFilesAsVectors(ctx, files, baseMetadata)
	if err != nil {
		log.Printf("Error storing files: %v", err)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("Error storing files: %v\n", err)))
		return
	}

	buf.WriteString("Successfully embedded and stored all files!\n\n")

	// Report approximate document count tracked by the vector manager (if available).
	// We use a minimal type assertion so we don't need to import the concrete manager type here.
	if getter, ok := db.VectorManager.(interface{ GetDocCount() int64 }); ok {
		docCount := getter.GetDocCount()
		buf.WriteString(fmt.Sprintf("Collection document count (approx): %d\n\n", docCount))
	}

	// Step 2: Query the vector database for information on ethnic enclaves
	log.Println("Querying vector database for ethnic enclaves information...")
	buf.WriteString("Step 2: Querying database for ethnic enclaves information...\n\n")

	query := "ethnic enclaves"
	results, err := db.VectorManager.RetrieveVectorWithQuery(ctx, query, 5)
	if err != nil {
		// Treat chromem-go nResults / collection-size errors as empty results rather than failing.
		// This prevents the handler from returning a 500 when the collection simply has fewer documents
		// than the requested number of results.
		if strings.Contains(err.Error(), "nResults") || strings.Contains(err.Error(), "number of documents") {
			log.Printf("Query returned nResults/collection-size error; treating as no results: %v", err)
			buf.WriteString(fmt.Sprintf("No results found for query '%s' (collection may be small).\n\n", query))
			results = nil
		} else {
			log.Printf("Error querying vector database: %v", err)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Error querying vector database: %v\n", err)))
			return
		}
	}

	log.Printf("Found %d results for query '%s'", len(results), query)
	buf.WriteString(fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), query))

	// Step 3: Store query results back into the vector database
	log.Println("Storing query results back into vector database...")
	buf.WriteString("Step 3: Storing query results...\n\n")

	var storedCount int
	for i, result := range results {
		// Create a summary of the result
		contentPreview := result.Content
		if len(contentPreview) > 100 {
			contentPreview = contentPreview[:100] + "..."
		}
		summary := fmt.Sprintf("Query result for '%s': %s (similarity: %.4f)", query, contentPreview, result.Similarity)

		resultMetadata := map[string]string{
			"query":        query,
			"result_index": fmt.Sprintf("%d", i),
			"original_id":  result.ID,
			"similarity":   fmt.Sprintf("%.4f", result.Similarity),
			"source":       "query_result",
		}

		// Embed the summary using the embedder from VectorManager
		embedding, err := db.VectorManager.GetEmbedder().EmbedText(ctx, summary)
		if err != nil {
			log.Printf("Error embedding query result %d: %v", i, err)
			buf.WriteString(fmt.Sprintf("Error embedding query result %d: %v\n", i, err))
			continue
		}

		resultVectorData := vector.VectorData{
			ID:        fmt.Sprintf("query_result_%s_%d_%d", query, i, time.Now().UnixNano()),
			Data:      summary,
			MetaData:  resultMetadata,
			Embedding: embedding,
		}

		err = db.VectorManager.StoreVectorInDB(ctx, resultVectorData)
		if err != nil {
			log.Printf("Error storing query result %d: %v", i, err)
			buf.WriteString(fmt.Sprintf("Error storing query result %d: %v\n", i, err))
			continue
		}

		storedCount++
		log.Printf("Stored query result %d", i)
		buf.WriteString(fmt.Sprintf("Result %d - ID: %s, Similarity: %.4f\n", i+1, result.ID, result.Similarity))

		contentDisplay := result.Content
		if len(contentDisplay) > 150 {
			contentDisplay = contentDisplay[:150] + "..."
		}
		buf.WriteString(fmt.Sprintf("  Content: %s\n\n", contentDisplay))
	}

	summary := fmt.Sprintf("\n✅ Test completed successfully!\n\nSummary:\n- Files processed: %d\n- Query results found: %d\n- Query results stored: %d\n", len(files), len(results), storedCount)
	buf.WriteString(summary)

	// Flush buffered output once all work has completed successfully.
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())

	log.Println("Test handler completed successfully")
}
