package manager

import (
	"context"
	"vex-backend/vector"
	"vex-backend/vector/embed"
)

type Manager interface {
	// can be a link, can be an embedded vector db, just needs to be the consistent throughout the manager's lifetime
	GetDBInstance() any
	GetEmbedder() embed.Embedder

	StoreVectorInDB(ctx context.Context, v vector.VectorData) error
	StoreVectorsInDB(ctx context.Context, vs []vector.VectorData) error
	StoreFileAsVectorsInDB(ctx context.Context, filename string) error

	RetriveVectorByMetadata(ctx context.Context, key string, data string) (vector.VectorData, error)
	RetriveVectorWithID(ctx context.Context, id string) (vector.VectorData, error)
	RetriveNVectorsByQuery(ctx context.Context, query string, n int) ([]vector.VectorData, error)

	DeleteVectorWithID(ctx context.Context, id string) error
	DeleteVectorsWithMetaData(ctx context.Context, key string, data string) error
}
