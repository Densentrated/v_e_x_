package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"vex-backend/config"
	"vex-backend/vector"
	"vex-backend/vector/embed"

	"github.com/philippgille/chromem-go"
)

type chromemManager struct {
	DBInstance *chromem.DB
	Embedder   embed.Embedder
}

// creates a Manager object for vectors,
func NewChromemManager(e embed.Embedder) Manager {
	var db *chromem.DB
	var err error

	storagePath := config.Config.VectorStorageFolder

	db, err = chromem.NewPersistentDB(storagePath, false)
	if err != nil {
		db = chromem.NewDB()
	}

	_, err = db.GetOrCreateCollection("notes", nil, e.EmbedToVector)
	if err != nil {
		panic("error getting or creating notes collection")
	}

	return &chromemManager{
		DBInstance: db,
		Embedder:   e,
	}
}

func (cm *chromemManager) getNotesCollection() chromem.Collection {
	return *cm.DBInstance.GetCollection("notes", cm.Embedder.EmbedToVector)
}
func (cm *chromemManager) GetDBInstance() any {
	return cm.DBInstance
}
func (cm *chromemManager) GetEmbedder() embed.Embedder {
	return cm.Embedder
}

// storage functions
func (cm *chromemManager) StoreVectorInDB(ctx context.Context, v vector.VectorData) error {
	doc := chromem.Document{
		ID:        fmt.Sprintf("note-%s", time.Now().UTC().Format("20060102T150405.000000000")),
		Metadata:  v.Metadata,
		Embedding: v.Embedding,
		Content:   v.Content,
	}

	col := cm.getNotesCollection()
	return (&col).AddDocument(ctx, doc)
}
func (cm *chromemManager) StoreVectorsInDB(ctx context.Context, vs []vector.VectorData) error {
	for _, v := range vs {
		if err := cm.StoreVectorInDB(ctx, v); err != nil {
			return err
		}
	}
	return nil
}
func (cm *chromemManager) StoreFileAsVectorsInDB(ctx context.Context, filename string) error {
	// get metadata
	// read the file to obtain emtadata, title, date,
	// get file data like the actual contents
	// convert ile data to chunks using chunking function
	// convert these chunks to StoreFileAsVectorsInDBrun the storevectorsinDB function
	// return an error if it happens

	// properly unfold filepath
	filepathParsed, err := filepath.Abs(filepath.Clean(filename))
	if err != nil {
		return err
	}

	info, err := os.Stat(filepathParsed)
	if err != nil {
		return err
	}

	metadata := map[string]string{
		"filename": filepath.Base(filepathParsed),
		"filepath": filepathParsed,
		"mod_time": info.ModTime().UTC().Format(time.RFC3339),
		"size":     string(info.Size()),
	}

	vs, err := cm.Embedder.EmbedFileToVectorData(ctx, filepathParsed, metadata)
	if err != nil {
		return err
	}

	if err := cm.StoreVectorsInDB(ctx, vs); err != nil {
		return err
	}

	return nil
}

// retrieval functions
func (cm *chromemManager) RetriveVectorByMetadata(ctx context.Context, key string, data string) (vector.VectorData, error) {
	where := map[string]string{key: data}
	col := cm.getNotesCollection()
	results, err := (&col).Query(ctx, "", 1, where, nil)
	if err != nil {
		return vector.VectorData{}, err
	}
	if len(results) == 0 {
		return vector.VectorData{}, fmt.Errorf("no document found with metadata %s=%s", key, data)
	}
	r := results[0]
	return vector.VectorData{
		Content:   r.Content,
		Embedding: r.Embedding,
		Metadata:  r.Metadata,
	}, nil
}
func (cm *chromemManager) RetriveVectorWithID(ctx context.Context, id string) (vector.VectorData, error) {
	col := cm.getNotesCollection()
	doc, err := (&col).GetByID(ctx, id)
	if err != nil {
		return vector.VectorData{}, err
	}
	return vector.VectorData{
		Content:   doc.Content,
		Embedding: doc.Embedding,
		Metadata:  doc.Metadata,
	}, nil
}
func (cm *chromemManager) RetriveNVectorsByQuery(ctx context.Context, query string, n int) ([]vector.VectorData, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be > 0")
	}
	col := cm.getNotesCollection()
	results, err := (&col).Query(ctx, query, n, nil, nil)
	if err != nil {
		return nil, err
	}
	out := make([]vector.VectorData, 0, len(results))
	for _, r := range results {
		out = append(out, vector.VectorData{
			Content:   r.Content,
			Embedding: r.Embedding,
			Metadata:  r.Metadata,
		})
	}
	return out, nil
}

// deletion functions
func (cm *chromemManager) DeleteVectorWithID(ctx context.Context, id string) error {
	col := cm.getNotesCollection()
	return (&col).Delete(ctx, nil, nil, id)
}
func (cm *chromemManager) DeleteVectorsWithMetaData(ctx context.Context, key string, data string) error {
	where := map[string]string{key: data}
	col := cm.getNotesCollection()

	return (&col).Delete(ctx, where, nil)
}
