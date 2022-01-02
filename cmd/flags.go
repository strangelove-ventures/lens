package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func peersFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("peers", false, "Comma-delimited list of peers to connect to for syncing")
	viper.BindPFlag("peers", cmd.Flags().Lookup("peers"))
	return cmd
}

func proveFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("prove", false, "return the proof as well as the result")
	viper.BindPFlag("prove", cmd.Flags().Lookup("prove"))
	return cmd
}

func limitFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int("limit", 100, "limit the number of things to fetch")
	viper.BindPFlag("limit", cmd.Flags().Lookup("limit"))
	return cmd
}

func skipConfirm(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().BoolP("skip", "y", false, "output using yaml")
	viper.BindPFlag("skip", cmd.Flags().Lookup("skip"))
	return cmd
}

var (
	FlagFrom = "from"
)

// AddTxFlagsToCmd defines common flags to be reused across cmds
func AddTxFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign, if left empty, the default key will be used")
}

// AddPaginationFlagsToCmd adds common pagination flags to cmd
func paginationFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Uint64("page", 1, "pagination page of objects to query for. This sets offset to a multiple of limit")
	viper.BindPFlag("page", cmd.Flags().Lookup("page"))

	cmd.Flags().String("page-key", "", "pagination page-key of objects to query for")
	viper.BindPFlag("page-key", cmd.Flags().Lookup("page-key"))

	cmd.Flags().Uint64("limit", 100, "pagination limit of objects to query for")
	viper.BindPFlag("limit", cmd.Flags().Lookup("limit"))

	cmd.Flags().Uint64("offset", 0, "pagination offset of objects to query for")
	viper.BindPFlag("offset", cmd.Flags().Lookup("offset"))

	cmd.Flags().Bool("count-total", true, "count total number of records in objects to query for")
	viper.BindPFlag("count-total", cmd.Flags().Lookup("count-total"))

	cmd.Flags().Bool("reverse", false, "results are sorted in descending order")
	viper.BindPFlag("reverse", cmd.Flags().Lookup("reverse"))
	return cmd
}

// ReadPageRequest reads and builds the necessary page request flags for pagination.
func ReadPageRequest(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	pageKey, _ := flagSet.GetString(flags.FlagPageKey)
	offset, _ := flagSet.GetUint64(flags.FlagOffset)
	limit, _ := flagSet.GetUint64(flags.FlagLimit)
	countTotal, _ := flagSet.GetBool(flags.FlagCountTotal)
	page, _ := flagSet.GetUint64(flags.FlagPage)
	reverse, _ := flagSet.GetBool(flags.FlagReverse)

	if page > 1 && offset > 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot specify both --page and --offset")
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	return &query.PageRequest{
		Key:        []byte(pageKey),
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}
