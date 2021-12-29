package client

import (
	"context"
	"errors"
	"fmt"
)

type ContextKey string

func NewContextKey(chainid string) ContextKey {
	return ContextKey(fmt.Sprintf("chain-client-context/%s", chainid))
}

// SetChainClientToContext sets the chain client to the context
func SetChainClientOnContext(ctx context.Context, chainid string, client *ChainClient) error {
	key := NewContextKey("clients")

	fmt.Printf("Creating key %s from %#v\n", key, ctx)
	v := ctx.Value(key)
	ptr := v.(map[string]*ChainClient)
	if ptr == nil {
		return errors.New("failed to type assert")
	}

	ptr[chainid] = client
	return nil
}

// GetChainClientFromContext returns the chain client from the context
func GetChainClientFromContext(ctx context.Context, chainid string) *ChainClient {
	key := NewContextKey("clients")

	fmt.Printf("Getting key %s from %#v\n", key, ctx)

	if v, ok := ctx.Value(key).(map[string]*ChainClient); ok {
		return v[chainid]
	}
	return nil
}
