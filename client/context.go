package client

import (
	"context"
	"errors"
	"fmt"
)

type contextKey string

func newContextKey(chainid string) contextKey {
	return contextKey(fmt.Sprintf("chain-client-context/%s", chainid))
}

// SetChainClientToContext sets the chain client to the context
func SetChainClientOnContext(ctx context.Context, chainid string, client *ChainClient) error {
	v := ctx.Value(newContextKey(chainid))
	if v == nil {
		return errors.New("chain client not found in context")
	}
	ptr := v.(*ChainClient)
	*ptr = *client
	return nil
}

// GetChainClientFromContext returns the chain client from the context
func GetChainClientFromContext(ctx context.Context, chainid string) *ChainClient {
	if v, ok := ctx.Value(newContextKey(chainid)).(*ChainClient); ok {
		return v
	}
	return nil
}
