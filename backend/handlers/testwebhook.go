package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	w.Write([]byte("Step 1: Embedding files and storing in database...\n\n"))

	baseMetadata := map[string]string{
		"source": "academic_notes",
	}

	// Use StoreFilesAsVectors which handles chunking, embedding, and storage
	err := db.VectorManager.StoreFilesAsVectors(ctx, files, baseMetadata)
	if err != nil {
		log.Printf("Error storing files: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error storing files: %v\n", err)))
		return
	}

	w.Write([]byte("Successfully embedded and stored all files!\n\n"))

	// Step 2: Query the vector database for information on ethnic enclaves
	log.Println("Querying vector database for ethnic enclaves information...")
	w.Write([]byte("Step 2: Querying database for ethnic enclaves information...\n\n"))

	query := "ethnic enclaves"
	results, err := db.VectorManager.RetrieveVectorWithQuery(ctx, query, 5)
	if err != nil {
		log.Printf("Error querying vector database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error querying vector database: %v\n", err)))
		return
	}

	log.Printf("Found %d results for query '%s'", len(results), query)
	w.Write([]byte(fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), query)))

	// Step 3: Store query results back into the vector database
	log.Println("Storing query results back into vector database...")
	w.Write([]byte("Step 3: Storing query results...\n\n"))

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
			w.Write([]byte(fmt.Sprintf("Error embedding query result %d: %v\n", i, err)))
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
			w.Write([]byte(fmt.Sprintf("Error storing query result %d: %v\n", i, err)))
			continue
		}

		storedCount++
		log.Printf("Stored query result %d", i)
		w.Write([]byte(fmt.Sprintf("Result %d - ID: %s, Similarity: %.4f\n", i+1, result.ID, result.Similarity)))

		contentDisplay := result.Content
		if len(contentDisplay) > 150 {
			contentDisplay = contentDisplay[:150] + "..."
		}
		w.Write([]byte(fmt.Sprintf("  Content: %s\n\n", contentDisplay)))
	}

	summary := fmt.Sprintf("\n✅ Test completed successfully!\n\nSummary:\n- Files processed: %d\n- Query results found: %d\n- Query results stored: %d\n", len(files), len(results), storedCount)
	w.Write([]byte(summary))
	log.Println("Test handler completed successfully")
}
