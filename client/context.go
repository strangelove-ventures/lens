package client

import (
	"context"
	"fmt"
)

type contextKey string

func newContextKey(chainid string) contextKey {
	return contextKey(fmt.Sprintf("chain-client-context/%s", chainid))
}

func SetChainClientOnContext(ctx context.Context, chainid string, client *ChainClient) context.Context {
	return context.WithValue(ctx, newContextKey(chainid), client)
}

func GetChainClientFromContext(ctx context.Context, chainid string) (*ChainClient, error) {
	if c, ok := ctx.Value(newContextKey(chainid)).(*ChainClient); ok {
		return c, nil
	}
	return nil, fmt.Errorf("chain client not found in context")
}
