package chat

import (
	"context"
	"fmt"
	"vex-backend/vector/manager"
)

func ProcessQuery(ctx context.Context, vm manager.Manager, query string) (string, error) {
	chat_platform := newOpenAIChatter()

	// Step 1: Use the chatter to translate the query into a better vector database query
	queryOptimizationPrompt := `You are a search query optimizer. Your job is to take a user's question and convert it into the best possible search terms for a vector database containing notes and documentation.

Rules:
- Focus on key concepts, not question words
- Remove filler words like "how", "what", "can you", etc.
- Include synonyms and related terms
- Keep it concise but comprehensive
- Return only the optimized search terms, no explanation

Convert this user question into optimized search terms:`

	optimizedQuery, err := chat_platform.GetResponseWithSystemPrompt(ctx, query, queryOptimizationPrompt)
	if err != nil {
		// Fallback to original query if optimization fails
		optimizedQuery = query
	}

	// Step 2: Query the vector database for top 4 relevant results
	results, err := vm.RetriveNVectorsByQuery(ctx, optimizedQuery, 4)
	if err != nil {
		return "", err
	}

	// Step 3: Build context from the retrieved results
	var context string
	if len(results) == 0 {
		context = "No relevant information found in the knowledge base."
	} else {
		context = "Relevant information from the knowledge base:\n\n"
		for i, result := range results {
			context += fmt.Sprintf("--- Document %d ---\n%s\n\n", i+1, result.Content)
		}
	}

	// Step 4: Use the chatter with system prompt to generate final answer
	answerPrompt := `You are a helpful assistant that answers questions using the provided knowledge base information.

Instructions:
- Use the provided context to answer the user's question
- If the context contains relevant information, use it to provide a comprehensive answer
- If the context doesn't contain enough information, say so clearly
- Be accurate and don't make up information not present in the context
- Format your response clearly and helpfully
- You should always specify specific documents if possible
- If you are going to use math equations, make sure to put like so $${math}$$ or ${math}$, this way the formatting will be done correctly

Context:
` + context

	response, err := chat_platform.GetResponseWithSystemPrompt(ctx, query, answerPrompt)
	if err != nil {
		return "", err
	}

	return response, nil
}
