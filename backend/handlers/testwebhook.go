package handlers

import (
	"fmt"
	"log"
	"net/http"
	"vex-backend/vector/embed"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	embedder := embed.VoyageEmbed{}
	filename := "/home/dense/Projects/Gitea/v_e_x_/NotesDir/techronomicon.git/Academia/Current Issues in Cities and Suburbs/Urban Segregation.md"
	chunks, err := embedder.CreateChunks(filename)

	if err != nil {
		log.Printf("error creating chunks: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating chunk: %v", err)))
		return
	}

	log.Printf("created %d chunks", len(chunks))

	if len(chunks) > 0 {
		metadata := map[string]string{
			"filename": filename,
		}
		vectorData, err := embedder.EmbedChunk(chunks[0], metadata)
		if err != nil {
			log.Printf("Error embedding chunk: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error embedding chunk: %v", err)))
			return
		}

		output := fmt.Sprintf("Successfully embedded chunk! Vector size: %d dimensions\nFirst 5 values: %v\n", len(vectorData.Embedding), vectorData.Embedding[:5])
		log.Print(output)
		w.Write([]byte(output))
	}
}
