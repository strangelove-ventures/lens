package cmd_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/strangelove-ventures/lens/client"
	"github.com/stretchr/testify/require"
)

func TestChainsShow_MissingArg(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.Run("chains", "show")
	require.Error(t, res.Err)
	require.Contains(t, res.Stderr.String(), "available names are: cosmoshub, osmosis")
	require.Empty(t, res.Stdout.String())
}

func TestChainsShowDefault(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.MustRun(t, "chains", "show-default")

	// Output must be the plain string, not a double-quoted JSON value string.
	require.Equal(t, res.Stdout.String(), "cosmoshub\n")
	require.Empty(t, res.Stderr.String())
}

func TestChainsSetDefault_ShowDefault(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.MustRun(t, "chains", "set-default", "osmosis")
	require.Empty(t, res.Stdout.String())
	require.Empty(t, res.Stderr.String())

	res = sys.MustRun(t, "chains", "show-default")
	require.Equal(t, res.Stdout.String(), "osmosis\n")
	require.Empty(t, res.Stderr.String())
}

func TestChainsSetDefault_Invalid(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.Run("chains", "set-default", "not_a_valid_chain_name")
	require.Error(t, res.Err)
	require.Empty(t, res.Stdout.String())
	require.Contains(t, res.Stderr.String(), "chain not_a_valid_chain_name not found")
}

func TestChainsDelete_Default(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.MustRun(t, "chains", "delete", "cosmoshub")
	require.Empty(t, res.Stdout.String())
	require.Contains(t, res.Stderr.String(), "Ignoring delete request for cosmoshub")

	// Confirm that it did, in fact, not change the default.
	res = sys.MustRun(t, "chains", "show-default")
	require.Equal(t, res.Stdout.String(), "cosmoshub\n")
	require.Empty(t, res.Stderr.String())
}

func TestChainEdit_Show(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	var before client.ChainClientConfig
	res := sys.MustRun(t, "chains", "show", "cosmoshub")
	require.Empty(t, res.Stderr.String())
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &before))

	res = sys.MustRun(t, "chains", "edit", "cosmoshub", "timeout", "1234s")
	require.Empty(t, res.Stdout.String())
	require.Empty(t, res.Stderr.String())

	var after client.ChainClientConfig
	res = sys.MustRun(t, "chains", "show", "cosmoshub")
	require.Empty(t, res.Stderr.String())
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &after))

	require.Equal(t, after.Timeout, "1234s")

	// Ensure nothing changed besides the Timeout field.
	require.Empty(
		t,
		cmp.Diff(before, after, cmpopts.IgnoreFields(client.ChainClientConfig{}, "Timeout")),
	)
}
