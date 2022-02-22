package query

import "github.com/cosmos/cosmos-sdk/types/query"

type Options struct {
	Pagination query.PageRequest
	Height     int64
}

func DefaultOptions() *Options {
	return &Options{
		Pagination: query.PageRequest{
			Key:        []byte(""),
			Offset:     0,
			Limit:      1000,
			CountTotal: true,
		},
		Height: 0,
	}
}
