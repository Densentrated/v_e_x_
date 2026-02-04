package manager

import "vex-backend/vector"

type vectorManager interface {
	// storage methods
	// stores a VectorData object as a vector in the database
	storeVectorInDB(vector vector.VectorData)
	// uses a vectorembed to convert a file into a a vector, and stores in the db
	storeFileAsVector(filename string) error
	// uses the storefileasvector method to store a bunch of files into the vector db
	storeFilesAsVectors(files []string) error

	// retrieval methods
	// gets a vectorStruct from the db using specifc metadata
	retrieveVectorWithMetaData(metadata map[string]string) error
	// gets a vectorStrct from the db using the specific embedding
	retrieveVectorWithEmbedding(embedding []int32) error
	// gets n vectorStructs from the db using the query
	retrieveVectorWithQuery(query string) error

	// db editing methods
	// deletesVectorsWithSpecifcEmbedding
	deleteVectorsWithEmbedding(embedding []int32) error
	// deletesVectorsWithSpecifcMetadata
	deleteVectorsWithMetaData(metadata map[string]string) error
}
