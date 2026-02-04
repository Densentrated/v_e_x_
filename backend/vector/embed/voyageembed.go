package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"vex-backend/config"
	"vex-backend/vector"
)

type VoyageEmbed struct{}

func (ve VoyageEmbed) CreateChunks(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := string(data)
	var chunks []string
	chunkSize := 3000
	overlapSize := 1000

	for i := 0; i < len(content); i += chunkSize - overlapSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunks = append(chunks, content[i:end])

		if end == len(content) {
			break
		}
	}

	return chunks, nil
}

type EmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

type EmbedRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

func (ve VoyageEmbed) EmbedChunk(chunk string, metadata map[string]string) (vector.VectorData, error) {
	reqBody := EmbedRequest{
		Input: chunk,
		Model: "voyage-4",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return vector.VectorData{}, err
	}

	req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return vector.VectorData{}, err
	}

	req.Header.Set("Authorization", "Bearer "+config.Config.VoyageAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return vector.VectorData{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return vector.VectorData{}, err
	}

	var embedResp EmbedResponse
	if err := json.Unmarshal(data, &embedResp); err != nil {
		return vector.VectorData{}, err
	}

	if len(embedResp.Data) == 0 {
		return vector.VectorData{}, fmt.Errorf("no embeddings returned")
	}

	// Convert float32 to int32 for the embedding field
	embedding := make([]int32, len(embedResp.Data[0].Embedding))
	for i, v := range embedResp.Data[0].Embedding {
		embedding[i] = int32(v)
	}

	return vector.VectorData{
		Data:      chunk,
		MetaData:  metadata,
		Embedding: embedding,
	}, nil
}
