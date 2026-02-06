package embed

import (
	"context"
	"vex-backend/vector"
)

type Embedder interface {
	EmbedToVector(ctx context.Context, content string) ([]float32, error)
	CreateChunks(ctx context.Context, content string) []string
	EmbedStringToVectorData(ctx context.Context, content string, metadata map[string]string) ([]vector.VectorData, error)
	EmbedFileToVectorData(ctx context.Context, filename string, metadat map[string]string) ([]vector.VectorData, error)
}
