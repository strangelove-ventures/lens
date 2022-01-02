package chain_info

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAllRPCEndpoints(t *testing.T) {
	testCases := map[string]struct {
		chainInfo         ChainInfo
		expectedEndpoints []string
		expectedError     error
	}{
		"endpoint with TLS": {
			chainInfo:         ChainInfoWithRPCEndpoint("https://test.com"),
			expectedEndpoints: []string{"https://test.com:443"},
			expectedError:     nil,
		},
		"endpoint without TLS": {
			chainInfo:         ChainInfoWithRPCEndpoint("http://test.com:26657"),
			expectedEndpoints: []string{"http://test.com:26657"},
			expectedError:     nil,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			endpoints, err := tc.chainInfo.GetAllRPCEndpoints()
			require.Equal(t, tc.expectedError, err)
			require.Equal(t, endpoints, tc.expectedEndpoints)
		})
	}
}

// too lazy to mock out an RPC instance that can simulate
// healthy/non-healthy behavior but that would be dope.
// one day...

func ChainInfoWithRPCEndpoint(endpoint string) ChainInfo {
	return ChainInfo{
		Apis: struct {
			RPC []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rpc"`
			Rest []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rest"`
		}{
			RPC: []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			}{
				{
					Address:  endpoint,
					Provider: "test",
				},
			},
		},
	}
}
