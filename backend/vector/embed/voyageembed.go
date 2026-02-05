package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"vex-backend/config"
	"vex-backend/vector"
)

// VoyageEmbed implements VectorEmbed using the Voyage AI API
type VoyageEmbed struct {
	Model string
}

// NewVoyageEmbed creates a new VoyageEmbed with default settings
func NewVoyageEmbed() *VoyageEmbed {
	return &VoyageEmbed{
		Model: "voyage-4",
	}
}

// NewVoyageEmbedWithModel creates a new VoyageEmbed with a specific model
func NewVoyageEmbedWithModel(model string) *VoyageEmbed {
	return &VoyageEmbed{
		Model: model,
	}
}

// EmbedRequest represents the request body for Voyage AI API
type EmbedRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbedResponse represents the response from Voyage AI API
type EmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CreateChunks reads a file and splits it into chunks with overlap
func (ve *VoyageEmbed) CreateChunks(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	content := string(data)
	var chunks []string
	chunkSize := 12000
	overlapSize := 2400

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

// EmbedText takes raw text and returns just the embedding vector
// This is the core embedding function used by other methods
func (ve *VoyageEmbed) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if config.Config.VoyageAPIKey == "" {
		return nil, fmt.Errorf("VOYAGE_API_KEY is not set in config")
	}

	reqBody := EmbedRequest{
		Input: text,
		Model: ve.Model,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.Config.VoyageAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Voyage AI: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage ai api returned status %d: %s", resp.StatusCode, string(respData))
	}

	var embedResp EmbedResponse
	if err := json.Unmarshal(respData, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embedResp.Error != nil {
		return nil, fmt.Errorf("voyage ai api error: %s", embedResp.Error.Message)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned from Voyage AI")
	}

	return embedResp.Data[0].Embedding, nil
}

// EmbedChunk takes a chunk and metadata, embeds the text, and returns the VectorData object
func (ve *VoyageEmbed) EmbedChunk(ctx context.Context, chunk string, metadata map[string]string) (vector.VectorData, error) {
	embedding, err := ve.EmbedText(ctx, chunk)
	if err != nil {
		return vector.VectorData{}, fmt.Errorf("failed to embed chunk: %w", err)
	}

	return vector.VectorData{
		Data:      chunk,
		MetaData:  metadata,
		Embedding: embedding,
	}, nil
}

// EmbedFile reads a file, chunks it, embeds all chunks, and returns all VectorData objects
func (ve *VoyageEmbed) EmbedFile(ctx context.Context, filename string, baseMetadata map[string]string) ([]vector.VectorData, error) {
	chunks, err := ve.CreateChunks(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks: %w", err)
	}

	var results []vector.VectorData
	for i, chunk := range chunks {
		// Create metadata for this chunk
		chunkMetadata := make(map[string]string)
		for k, v := range baseMetadata {
			chunkMetadata[k] = v
		}
		chunkMetadata["chunk_index"] = fmt.Sprintf("%d", i)
		chunkMetadata["total_chunks"] = fmt.Sprintf("%d", len(chunks))

		vectorData, err := ve.EmbedChunk(ctx, chunk, chunkMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		// Generate ID for the vector
		vectorData.ID = fmt.Sprintf("%s_chunk_%d_%d", filename, i, time.Now().UnixNano())

		results = append(results, vectorData)
	}

	return results, nil
}
