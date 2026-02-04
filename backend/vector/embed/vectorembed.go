package embed

import (
	"vex-backend/vector"
)

type vectorEmbed interface {
	// takes a file, turns it into chunks, and returns the chunk string, returns an error if it happens
	// tye chunks are a size that the embedding model can handle and have a 2/5 overlap with each other
	CreateChunks(filename string) ([]string, error)

	//Embed chunk takes a chunk, a metadata map, embeds the file, and then returns the vectorDaata object
	EmbedChunk(chunk string, metadata map[string]string) (vector.VectorData, error)
}
