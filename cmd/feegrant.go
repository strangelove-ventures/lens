package cmd

import (
	"fmt"
	"reflect"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/spf13/cobra"
)

// flag for feegrant module
const (
	FlagExpiration  = "expiration"
	FlagPeriod      = "period"
	FlagPeriodLimit = "period-limit"
	FlagSpendLimit  = "spend-limit"
	FlagAllowedMsgs = "allowed-messages"
)

func feegrantQueryGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants [grantee]",
		Short: "query things about a grantee",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			pageReq, err := ReadPageRequest(cmd.Flags())

			if err != nil {
				return err
			}

			var address types.AccAddress
			var keyNameOrAddress string

			if len(args) == 0 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[0]
			}

			address, err = cl.AccountFromKeyOrAddress(keyNameOrAddress)

			if err != nil {
				return err
			}

			grants, err := cl.QueryFeeGrants(address, pageReq)
			if err != nil {
				return err
			}

			// FIXME: if anyone can tell me why JSON output without this doesn't work gets a 5 OSMO tip
			// FIXME: seems like there should be no side effects with using the MarshalProto function, but thats not the
			for grant := range grants {
				_, err := cl.MarshalProto(grants[grant])
				if err != nil {
					return err
				}
			}

			return cl.PrintObject(grants)
		},
	}

	return paginationFlags(cmd)
}

func feegrantQueryGrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant grantee [granter]",
		Short: "query things about a single grant",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			var (
				granterAddr      types.AccAddress
				granteeAddr      types.AccAddress
				err              error
				keyNameOrAddress string
			)

			if len(args) == 1 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[1]
			}

			granterAddr, err = cl.AccountFromKeyOrAddress(keyNameOrAddress)

			if err != nil {
				return fmt.Errorf("did not find wallet %s", keyNameOrAddress)
			}

			fmt.Println(args[0])

			granteeAddr, err = cl.DecodeBech32AccAddr(args[0])
			if err != nil {
				return err
			}

			grant, err := cl.QueryFeeGrant(granteeAddr, granterAddr)

			if err != nil {
				return err
			}

			return cl.PrintObject(grant)
		},
	}

	return cmd
}

func feegrantFeeGrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant grantee",
		Short: "Grant fee allowance to an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				granterAddr sdk.AccAddress
				granteeAddr sdk.AccAddress
				err         error
			)

			cl := config.GetDefaultClient()
			key, _ := cmd.Flags().GetString(FlagFrom)
			granterAddr, err = cl.AccountFromKeyOrAddress(key)
			if err != nil {
				return err
			}

			granteeAddr, err = cl.DecodeBech32AccAddr(args[0])
			if err != nil {
				return err
			}

			spendLimit, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}

			spendLimitCoins, err := sdk.ParseCoinsNormalized(spendLimit)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetString(FlagExpiration)
			if err != nil {
				return err
			}

			basicGrant := feegrant.BasicAllowance{
				SpendLimit: spendLimitCoins,
			}

			var expiresAtTime time.Time
			if exp != "" {
				expiresAtTime, err = time.Parse(time.RFC3339, exp)
				if err != nil {
					return err
				}

				basicGrant.Expiration = &expiresAtTime
			}

			var grant feegrant.FeeAllowanceI
			grant = &basicGrant

			periodClock, err := cmd.Flags().GetInt64(FlagPeriod)
			if err != nil {
				return err
			}

			periodLimitVal, err := cmd.Flags().GetString(FlagPeriodLimit)
			if err != nil {
				return err
			}

			if periodClock > 0 || periodLimitVal != "" {
				periodLimit, err := sdk.ParseCoinsNormalized(periodLimitVal)
				if err != nil {
					return err
				}

				if periodClock <= 0 {
					return fmt.Errorf("period clock was not set")
				}

				if periodLimit == nil {
					return fmt.Errorf("period limit was not set")
				}

				periodReset := getPeriodReset(periodClock)
				if exp != "" && periodReset.Sub(expiresAtTime) > 0 {
					return fmt.Errorf("period (%d) cannot reset after expiration (%v)", periodClock, exp)
				}

				periodic := feegrant.PeriodicAllowance{
					Basic:            basicGrant,
					Period:           getPeriod(periodClock),
					PeriodReset:      getPeriodReset(periodClock),
					PeriodSpendLimit: periodLimit,
					PeriodCanSpend:   periodLimit,
				}

				grant = &periodic
			}

			allowedMsgs, err := cmd.Flags().GetStringSlice(FlagAllowedMsgs)
			if err != nil {
				return err
			}

			if len(allowedMsgs) > 0 {
				grant, err = feegrant.NewAllowedMsgAllowance(grant, allowedMsgs)
				if err != nil {
					return err
				}

			}

			msg, err := feegrant.NewMsgGrantAllowance(grant, granterAddr, granteeAddr)
			if err != nil {
				return err
			}

			return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), msg))
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().StringSlice(FlagAllowedMsgs, []string{}, "Set of allowed messages for fee allowance")
	cmd.Flags().String(FlagExpiration, "", "The RFC 3339 timestamp after which the grant expires for the user")
	cmd.Flags().String(FlagSpendLimit, "", "Spend limit specifies the max limit can be used, if not mentioned there is no limit")
	cmd.Flags().Int64(FlagPeriod, 0, "period specifies the time duration(in seconds) in which period_limit coins can be spent before that allowance is reset (ex: 3600)")
	cmd.Flags().String(FlagPeriodLimit, "", "period limit specifies the maximum number of coins that can be spent in the period")

	return cmd
}

func feegrantRevokeFeeGrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke grantee",
		Short: "Grant fee allowance to an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("a")
			var (
				granterAddr sdk.AccAddress
				granteeAddr sdk.AccAddress
				err         error
			)

			cl := config.GetDefaultClient()
			key, _ := cmd.Flags().GetString(FlagFrom)
			granterAddr, err = cl.AccountFromKeyOrAddress(key)
			if err != nil {
				return err
			}

			granteeAddr, err = cl.DecodeBech32AccAddr(args[0])
			if err != nil {
				return err
			}

			msg := feegrant.NewMsgRevokeAllowance(granterAddr, granteeAddr)
			fmt.Println(reflect.TypeOf(msg))

			// sends tx but fails due to gas

			return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), &msg))
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getPeriodReset(duration int64) time.Time {
	return time.Now().Add(getPeriod(duration))
}

func getPeriod(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}
