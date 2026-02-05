package embed

import (
	"context"

	"vex-backend/vector"
)

// VectorEmbed defines the interface for embedding operations
// This interface is provider-agnostic and can be implemented by different embedding providers
// Implementations include VoyageEmbed, OpenAIEmbed, etc.
type VectorEmbed interface {
	// CreateChunks takes a file, turns it into chunks, and returns the chunk strings
	// The chunks are a size that the embedding model can handle and have overlap with each other
	CreateChunks(filename string) ([]string, error)

	// EmbedChunk takes a chunk, a metadata map, embeds the text, and returns the VectorData object
	EmbedChunk(ctx context.Context, chunk string, metadata map[string]string) (vector.VectorData, error)

	// EmbedText takes a plain text string and returns its embedding vector
	// This is useful for embedding queries without creating full VectorData
	EmbedText(ctx context.Context, text string) ([]float32, error)

	// EmbedFile reads a file, chunks it, embeds all chunks, and returns all VectorData objects
	// baseMetadata is copied to each chunk's metadata with chunk-specific fields added
	EmbedFile(ctx context.Context, filename string, baseMetadata map[string]string) ([]vector.VectorData, error)
}
