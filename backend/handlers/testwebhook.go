package handlers

import (
	"fmt"
	"log"
	"net/http"
	"vex-backend/storage"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	embedder := storage.VoyageEmbedder{}
	filename := "/home/dense/Projects/Gitea/v_e_x_/backend/~/Projects/Gitea/v_e_x_/NotesDir/techronomicon.git/Academia/Current Issues in Cities and Suburbs/Urban Segregation.md"
	chunks, err := embedder.CreateChunks(filename)

	if err != nil {
		log.Printf("error creating chunks: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating chunk", err)))
		return
	}

	log.Printf("created %d chunks", len(chunks))

	if len(chunks) > 0 {
		embedding, err := embedder.EmbedChunk(chunks[0])
		if err != nil {
			log.Printf("Error embedding chunk: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error embedding chunk: %v", err)))
			return
		}

		output := fmt.Sprintf("Successfully embedded chunk! Vector size: %d dimensions\nFirst 5 values: %v\n", len(embedding), embedding[:5])
		log.Print(output)
		w.Write([]byte(output))
	}
}
