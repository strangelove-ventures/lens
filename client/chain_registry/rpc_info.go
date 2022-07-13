package chain_registry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/strangelove-ventures/lens/client"
	"golang.org/x/sync/errgroup"
)

// NewChainInfo returns a ChainInfo that is uninitialized other than the provided zap.Logger.
// Typically, the caller will unmarshal JSON content into the ChainInfo after initialization.
func NewChainInfo() ChainJson {
	return ChainJson{}
}

func GetAllRPCEndpoints(rpcs []struct {
	Address  string "json:\"address\""
	Provider string "json:\"provider\""
}) (out []string, err error) {
	for _, endpoint := range rpcs {
		u, err := url.Parse(endpoint.Address)
		if err != nil {
			return nil, err
		}

		var port string
		if u.Port() == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return nil, fmt.Errorf("invalid or unsupported url scheme: %v", u.Scheme)
			}
		} else {
			port = u.Port()
		}

		out = append(out, fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path))
	}

	return out, nil
}

func IsHealthyRPC(ctx context.Context, endpoint string) error {
	cl, err := client.NewRPCClient(endpoint, 5*time.Second)
	if err != nil {
		return err
	}
	stat, err := cl.Status(ctx)
	if err != nil {
		return err
	}

	if stat.SyncInfo.CatchingUp {
		return errors.New("still catching up")
	}

	return nil
}

func CheckRPCEndpoints(ctx context.Context, allRPCEndpoints []string) (out []string, err error) {
	var eg errgroup.Group
	var endpoints []string
	for _, endpoint := range allRPCEndpoints {
		endpoint := endpoint
		eg.Go(func() error {
			err := IsHealthyRPC(ctx, endpoint)
			if err != nil {
				fmt.Println("Ignoring endpoint due to error")
				fmt.Println("endpoint", endpoint)
				fmt.Println(err)
				// 	"Ignoring endpoint due to error",
				// 	zap.String("endpoint", endpoint),
				// 	zap.Error(err),
				// )
				return nil
			}

			fmt.Println("Verified healthy endpoint")
			fmt.Println("endpoint", endpoint)
			// c.log.Info("Verified healthy endpoint", zap.String("endpoint", endpoint))
			endpoints = append(endpoints, endpoint)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func GetRandomRPCEndpoint(ctx context.Context, rpcEndpoints []struct {
	Address  string "json:\"address\""
	Provider string "json:\"provider\""
}) (string, error) {
	allRPCEndpoints, err := GetAllRPCEndpoints(rpcEndpoints)
	if err != nil {
		fmt.Println("error: ", err)
	}
	rpcs, err := CheckRPCEndpoints(ctx, allRPCEndpoints)
	if err != nil {
		return "", err
	}

	if len(rpcs) == 0 {
		return "", fmt.Errorf("no working RPCs found")
	}

	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rpcs[randomGenerator.Intn(len(rpcs))], nil
}
