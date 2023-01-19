package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/term"
)

const (
	flagCoinType           = "coin-type"
	defaultCoinType uint32 = sdk.CoinType
)

// keysCmd represents the keys command
func keysCmd(a *appState) *cobra.Command {
	// flagAccountPrefix allows the user to override the prefix for a given account.
	// Scoped to keysCmd so that concurrent test runs don't encounter a data race.
	var flagAccountPrefix string

	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "manage keys held by the relayer for each chain",
	}

	cmd.AddCommand(
		keysAddCmd(a),
		keysRestoreCmd(a),
		keysDeleteCmd(a),
		keysListCmd(a),
		keysShowCmd(a, &flagAccountPrefix),
		keysEnumerateCmd(a),
		keysExportCmd(a),
	)

	return cmd
}

// keysAddCmd respresents the `keys add` command
func keysAddCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [name]",
		Aliases: []string{"a"},
		Short:   "adds a key to the keychain associated with a particular chain",
		Long:    "if no name is passed, 'default' is used",
		Args:    cobra.RangeArgs(0, 1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys add
$ %s keys add test_key
$ %s k a osmo_key --chain osmosis`, appName, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			var keyName string
			if len(args) == 0 {
				keyName = cl.Config.Key
			} else {
				keyName = args[0]
			}
			if cl.KeyExists(keyName) {
				return errKeyExists(keyName)
			}

			coinType, err := cmd.Flags().GetUint32(flagCoinType)
			if err != nil {
				return err
			}

			ko, err := cl.AddKey(keyName, coinType)
			if err != nil {
				return err
			}

			// Not calling writeJSON because this is one case that does not use indentation.
			// (Was that intentional?)
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetEscapeHTML(false)
			if err := enc.Encode(&ko); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().Uint32(flagCoinType, defaultCoinType, "coin type number for HD derivation")
	return cmd
}

// keysRestoreCmd respresents the `keys add` command
func keysRestoreCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore [name]",
		Aliases: []string{"r"},
		Short:   "restores a mnemonic to the keychain associated with a particular chain",
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys restore --chain ibc-0 testkey
$ %s k r --chain ibc-1 faucet-key`, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			keyName := args[0]
			if cl.KeyExists(keyName) {
				return errKeyExists(keyName)
			}

			mnemonic, err := readMnemonic(cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				// Can happen when there is an issue with the terminal.
				return fmt.Errorf("failed to read mnemonic: %w", err)
			}

			coinType, err := cmd.Flags().GetUint32(flagCoinType)
			if err != nil {
				return err
			}

			address, err := cl.RestoreKey(keyName, string(mnemonic), coinType)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), address)
			return nil
		},
	}
	cmd.Flags().Uint32(flagCoinType, defaultCoinType, "coin type number for HD derivation")
	return cmd
}

// readMnemonic reads a password in terminal mode if stdin is a terminal,
// otherwise it returns all of stdin with the trailing newline removed.
func readMnemonic(stdin io.Reader, stderr io.Writer) ([]byte, error) {
	type fder interface {
		Fd() uintptr
	}

	if f, ok := stdin.(fder); ok {
		fmt.Fprint(stderr, "Enter mnemonic 🔑: ")
		mnemonic, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(stderr)
		return mnemonic, err
	}

	in, err := io.ReadAll(stdin)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(in, []byte("\n")), nil
}

// keysDeleteCmd represents the `keys delete` command.
func keysDeleteCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "deletes a key from the keychain associated with a particular chain",
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys delete ibc-0 -y
$ %s keys delete ibc-1 key2 -y
$ %s k d ibc-2 testkey`, appName, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			chainName := cl.Config.ChainID
			keyName := args[0]
			if !cl.KeyExists(keyName) {
				return errKeyDoesntExist(keyName)
			}

			if skip, _ := cmd.Flags().GetBool("skip"); !skip {
				fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete key(%s) from chain(%s)? (Y/n)\n", keyName, chainName)
				if !askForConfirmation(a.Log, cmd) {
					return nil
				}
			}

			if err := cl.DeleteKey(keyName); err != nil {
				panic(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "key %s deleted\n", keyName)
			return nil
		},
	}

	return skipConfirm(cmd, a.Viper)
}

func askForConfirmation(log *zap.Logger, cmd *cobra.Command) bool {
	var response string

	_, err := fmt.Fscanln(cmd.InOrStdin(), &response)
	if err != nil {
		log.Fatal("Failed to read input", zap.Error(err))
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Fprintln(cmd.OutOrStdout(), "please type (y)es or (n)o and then press enter")
		return askForConfirmation(log, cmd)
	}
}

// keysListCmd respresents the `keys list` command
func keysListCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "lists keys from the keychain associated with a particular chain",
		Args:    cobra.NoArgs,
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys list ibc-0
$ %s k l ibc-1`, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			info, err := cl.ListAddresses()
			if err != nil {
				return err
			}

			if len(info) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: no keys found")
				return nil
			}

			for key, val := range info {
				fmt.Fprintf(cmd.OutOrStdout(), "key(%s) -> %s\n", key, val)
			}

			return nil
		},
	}

	return cmd
}

// keysShowCmd respresents the `keys show` command
func keysShowCmd(a *appState, flagAccountPrefix *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "shows a key from the keychain associated with a particular chain",
		Long:    "if no name is passed, name in config is used",
		Args:    cobra.RangeArgs(0, 1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys show ibc-0
$ %s keys show ibc-1 key2
$ %s k s ibc-2 testkey`, appName, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			var keyName string
			if len(args) == 0 {
				keyName = cl.Config.Key
			} else {
				keyName = args[0]
			}
			if !cl.KeyExists(keyName) {
				return errKeyDoesntExist(keyName)
			}

			if *flagAccountPrefix != "" {
				cl.Config.AccountPrefix = *flagAccountPrefix
			}

			address, err := cl.ShowAddress(keyName)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), address)
			return nil
		},
	}

	cmd.Flags().StringVar(flagAccountPrefix, "prefix", "", "Encode the key with the user specified prefix")

	return cmd
}

type KeyEnumeration struct {
	KeyName   string            `json:"key_name"`
	Addresses map[string]string `json:"addresses"`
}

// keysEnumerateCmd respresents the `keys enumerate` command
func keysEnumerateCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enumerate [key-or-address]",
		Aliases: []string{"e"},
		Short:   "enumerates the address for a given key across all configured chains",
		Long:    "if no key or address is passed, the key on the default chain is used to enumerate through all configured chains",
		Args:    cobra.RangeArgs(0, 1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys enumerate
$ %s keys enumerate key2
$ %s k e key2`, appName, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			var keyName string
			if len(args) == 0 {
				keyName = cl.Config.Key
			} else {
				keyName = args[0]
			}
			account, err := cl.AccountFromKeyOrAddress(keyName)
			if err != nil {
				return err
			}

			chains := make([]string, 0, len(a.Config.Chains))
			for chain := range a.Config.Chains {
				chains = append(chains, chain)
			}
			sort.Strings(chains)

			addresses := make(map[string]string)
			for _, chain := range chains {
				client := a.Config.GetClient(chain)
				address, err := client.EncodeBech32AccAddr(account)
				if err != nil {
					return err
				}
				addresses[chain] = address
			}

			return cl.PrintObject(addresses)
		},
	}

	// cmd.Flags().StringVar(&FlagAccountPrefix, "prefix", "", "Encode the key with the user specified prefix")

	return cmd
}

// keysExportCmd respresents the `keys export` command
func keysExportCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [name]",
		Aliases: []string{"e"},
		Short:   "exports a privkey from the keychain associated with a particular chain",
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s keys export ibc-0 testkey
$ %s k e ibc-2 testkey`, appName, appName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			keyName := args[0]
			if !cl.KeyExists(keyName) {
				return errKeyDoesntExist(keyName)
			}

			info, err := cl.ExportPrivKeyArmor(keyName)
			if err != nil {
				return err
			}

			fmt.Println(info)
			return nil
		},
	}

	return cmd
}

func errKeyExists(name string) error {
	return fmt.Errorf("a key with name %s already exists", name)
}

func errKeyDoesntExist(name string) error {
	return fmt.Errorf("a key with name %s doesn't exist", name)
}
