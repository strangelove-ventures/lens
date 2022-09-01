package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tmquery "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client/query"
)

const (
	gRPCSecureOnlyFlag = "secure-only"
	flagMemo           = "memo"
)

func peersFlag(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().Bool("peers", false, "Comma-delimited list of peers to connect to for syncing")
	v.BindPFlag("peers", cmd.Flags().Lookup("peers"))
	return cmd
}

func proveFlag(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().Bool("prove", false, "return the proof as well as the result")
	v.BindPFlag("prove", cmd.Flags().Lookup("prove"))
	return cmd
}

func limitFlag(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().Int("limit", 100, "limit the number of things to fetch")
	v.BindPFlag("limit", cmd.Flags().Lookup("limit"))
	return cmd
}

func skipConfirm(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().BoolP("skip", "y", false, "output using yaml")
	v.BindPFlag("skip", cmd.Flags().Lookup("skip"))
	return cmd
}

func gRPCFlags(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().Bool(gRPCSecureOnlyFlag, false, "do not fall back to skipping TLS verification when connecting to server")
	if err := v.BindPFlag(gRPCSecureOnlyFlag, cmd.Flags().Lookup(gRPCSecureOnlyFlag)); err != nil {
		panic(err)
	}

	return cmd
}

func memoFlag(v *viper.Viper, cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String(flagMemo, "", "a memo to include in transaction")
	if err := v.BindPFlag(flagMemo, cmd.Flags().Lookup(flagMemo)); err != nil {
		panic(err)
	}
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
func paginationFlags(cmd *cobra.Command, v *viper.Viper) *cobra.Command {
	cmd.Flags().Uint64("page", 1, "pagination page of objects to query for. This sets offset to a multiple of limit")
	v.BindPFlag("page", cmd.Flags().Lookup("page"))

	cmd.Flags().String("page-key", "", "pagination page-key of objects to query for")
	v.BindPFlag("page-key", cmd.Flags().Lookup("page-key"))

	cmd.Flags().Uint64("limit", 100, "pagination limit of objects to query for")
	v.BindPFlag("limit", cmd.Flags().Lookup("limit"))

	cmd.Flags().Uint64("offset", 0, "pagination offset of objects to query for")
	v.BindPFlag("offset", cmd.Flags().Lookup("offset"))

	cmd.Flags().Bool("count-total", true, "count total number of records in objects to query for")
	v.BindPFlag("count-total", cmd.Flags().Lookup("count-total"))

	cmd.Flags().Bool("reverse", false, "results are sorted in descending order")
	v.BindPFlag("reverse", cmd.Flags().Lookup("reverse"))
	return cmd
}

// ReadPageRequest reads and builds the necessary page request flags for pagination.
func ReadPageRequest(flagSet *pflag.FlagSet) (*tmquery.PageRequest, error) {
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

	return &tmquery.PageRequest{
		Key:        []byte(pageKey),
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}

// ReadHeight reads the height flag.
func ReadHeight(flagSet *pflag.FlagSet) (int64, error) {
	if flagSet.Changed(flags.FlagHeight) {
		height, err := flagSet.GetInt64(flags.FlagHeight)
		if err != nil {
			return 0, err
		}
		return height, nil
	} else {
		return 0, nil
	}
}

func queryOptionsFromFlags(flags *pflag.FlagSet) (*query.QueryOptions, error) {
	// Query options
	pr, err := ReadPageRequest(flags)
	if err != nil {
		return nil, err
	}
	height, err := ReadHeight(flags)
	if err != nil {
		return nil, err
	}

	return &query.QueryOptions{Pagination: pr, Height: height}, nil
}
