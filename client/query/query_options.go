package query

import "github.com/cosmos/cosmos-sdk/types/query"

type QueryOptions struct {
	Pagination *query.PageRequest
	Height     int64
}

func DefaultOptions() *QueryOptions {
	return &QueryOptions{
		Pagination: &query.PageRequest{
			Key:        []byte(""),
			Offset:     0,
			Limit:      1000,
			CountTotal: true,
		},
		Height: 0,
	}
}
