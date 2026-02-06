package chat

import "context"

type chatter interface {
	GetResponse(ctx context.Context, query string) (string, error)
	GetResponseWithSystemPrompt(ctx context.Context, query string, systemprompt string) (string, error)
}
