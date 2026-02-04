package storage

// Embed is an interface for embedding text chunks into vectors
type Embedder interface {
	// CreateChunks take a file, and turns it into a chunk of a size that
	// the model can handle, and chunk overlap of about 1/10 of that size
	CreateChunks(filename string) ([]string, error)

	// EmbedChunk takes a chunk and returns the embedded vector as a []float32
	// logs fatal error if the chunk is greater than the max character amount
	// if amount is fine, return the embedded vector in a []float32 format
	EmbedChunk(chunk string) ([]float32, error)
}
