package vector

type VectorData struct {
	Content   string
	Embedding []float32
	Metadata  map[string]string
	id        string
}
