package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"vex-backend/config"
	"vex-backend/vector"
)

type voyageEmbed struct {
	Model string
}

func NewVoyageEmbed(model string) Embedder {
	return &voyageEmbed{
		Model: model,
	}
}

func (ve voyageEmbed) CreateChunks(ctx context.Context, content string) []string {
	const maxChunkRunes = 10000 // tune as needed for model token limits
	overlapRunes := maxChunkRunes / 5

	content = strings.TrimSpace(content)
	if content == "" {
		return []string{}
	}

	words := strings.Fields(content)
	if len(words) == 0 {
		return []string{content}
	}

	var chunks []string

	for start := 0; start < len(words); {
		cur := 0
		end := start

		// build chunk from start..end (exclusive) not exceeding maxChunkRunes (measured in bytes)
		for end < len(words) {
			wlen := len(words[end])
			add := wlen
			if end > start {
				add += 1 // space
			}
			if cur+add > maxChunkRunes {
				// if no progress (single word larger than limit), include it anyway
				if end == start {
					end++
				}
				break
			}
			cur += add
			end++
		}

		// create chunk string
		chunk := strings.Join(words[start:end], " ")
		chunks = append(chunks, strings.TrimSpace(chunk))

		// if we've reached the end, break
		if end >= len(words) {
			break
		}

		// determine how many words to overlap to reach approximately overlapRunes
		ovAccum := 0
		overlapCount := 0
		for k := end - 1; k >= start; k-- {
			if overlapCount == 0 {
				ovAccum += len(words[k])
			} else {
				ovAccum += 1 + len(words[k]) // space + word
			}
			overlapCount++
			if ovAccum >= overlapRunes {
				break
			}
		}

		newStart := end - overlapCount
		// ensure progress; if overlap would not move forward, advance to end
		if newStart <= start {
			newStart = end
		}
		start = newStart
	}

	return chunks
}

func (ve voyageEmbed) EmbedToVector(ctx context.Context, content string) ([]float32, error) {
	voyageAPIKey := config.Config.VoyageAPIKey

	// assume that the string here is of appropriate size
	reqBody := map[string]any{
		"input":      []string{content},
		"model":      ve.Model,
		"inpyt_type": "document",
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.voyageai.com/v1/embeddings", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+voyageAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("voyage API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	type dataItem struct {
		Object    string    `json:"object`
		Embedding []float32 `json"embedding"`
		Index     int       `json: "index"`
	}

	type voyageResp struct {
		Object string     `json:"object"`
		Data   []dataItem `json:"data"`
		Model  string     `json:"model"`
		Usage  any        `json:"usage"`
	}

	var vr voyageResp
	if err := json.Unmarshal(respBytes, &vr); err == nil {
		// if the API returned float32 arrays directly into the struct, we're done
		if len(vr.Data) > 0 && len(vr.Data[0].Embedding) > 0 {
			return vr.Data[0].Embedding, nil
		}
		// If Data is empty, fall through to tolerant decode below
	} else {
		// try tolerant decoding (in case embedding numbers are float64 or mixed)
		var alt map[string]any
		if err2 := json.Unmarshal(respBytes, &alt); err2 == nil {
			if rawData, ok := alt["data"].([]any); ok && len(rawData) > 0 {
				if first, ok := rawData[0].(map[string]any); ok {
					if embRaw, ok := first["embedding"].([]any); ok {
						emb := make([]float32, 0, len(embRaw))
						for _, v := range embRaw {
							switch vv := v.(type) {
							case float32:
								emb = append(emb, vv)
							case float64:
								emb = append(emb, float32(vv))
							case int:
								emb = append(emb, float32(vv))
							case int64:
								emb = append(emb, float32(vv))
							default:
								// skip non-numeric entries
							}
						}
						if len(emb) > 0 {
							return emb, nil
						}
					}
				}
			}
		}
		// final fallback: return original unmarshal error
		return nil, fmt.Errorf("failed to parse voyage response: %w", err)
	}

	// If we reached here, vr was unmarshaled but no usable embedding found
	return nil, fmt.Errorf("voyage response did not contain an embedding")
}

func (ve voyageEmbed) EmbedStringToVectorData(ctx context.Context, content string, metadata map[string]string) ([]vector.VectorData, error) {
	chunks := ve.CreateChunks(ctx, content)
	vectors := []vector.VectorData{}
	for _, chunk := range chunks {
		embedding, err := ve.EmbedToVector(ctx, chunk)
		if err != nil {
			return nil, err
		}

		short := chunk
		if len(short) > 32 {
			short = short[:32]
		}

		chunkVectorData := vector.VectorData{
			Content:   chunk,
			Embedding: embedding,
			Metadata:  metadata,
			// create a reasonably unique ID using a short prefix of the chunk, the chunk pointer and embedding length
			Id: fmt.Sprintf("voyage-%x-%p-%d", short, &chunk, len(embedding)),
		}
		vectors = append(vectors, chunkVectorData)
	}
	return vectors, nil
}

func (ve voyageEmbed) EmbedFileToVectorData(ctx context.Context, filename string, metadata map[string]string) ([]vector.VectorData, error) {
	// Read the entire file content
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Delegate to EmbedStringToVectorData
	return ve.EmbedStringToVectorData(ctx, string(b), metadata)
}
