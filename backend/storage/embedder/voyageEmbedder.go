package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"vex-backend/config"
)

type VoyageEmbedder struct{}

func (ve VoyageEmbedder) CreateChunks(filename string) ([]string, error) {
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

func (ve VoyageEmbedder) EmbedChunk(chunk string) ([]float32, error) {
	reqBody := EmbedRequest{
		Input: chunk,
		Model: "voyage-4",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+config.Config.VoyageAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var embedResp EmbedResponse
	if err := json.Unmarshal(data, &embedResp); err != nil {
		return nil, err
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Data[0].Embedding, nil
}
