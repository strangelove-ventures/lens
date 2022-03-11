package cmd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysList_EmptyKeys(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	res := sys.MustRun(t, "keys", "list")

	// Before adding any keys, listing the keys gives a helpful message on stderr.
	require.Empty(t, res.Stdout.String())
	require.Contains(t, res.Stderr.String(), "no keys found")
}

func TestKeysAdd_List(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	sys.MustRun(t, "keys", "add")

	res := sys.MustRun(t, "keys", "list")
	require.Contains(t, res.Stdout.String(), "key(default) -> cosmos1")
	require.Empty(t, res.Stderr.String())
}

func TestKeysAdd_CustomName_List(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	sys.MustRun(t, "keys", "add", "foo")

	res := sys.MustRun(t, "keys", "list")
	require.Contains(t, res.Stdout.String(), "key(foo) -> cosmos1")
	require.Empty(t, res.Stderr.String())
}

func TestKeys_Restore(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	in := strings.NewReader(ZeroMnemonic + "\n")
	res := sys.MustRunWithInput(t, in, "keys", "restore", "mykey")
	require.Equal(t, res.Stdout.String(), ZeroCosmosAddr+"\n")
	// res.Stderr can be ignored here.

	// After calling restore, the key can be retrieved through keys list.
	res = sys.MustRun(t, "keys", "list")
	require.Equal(t, res.Stdout.String(), "key(mykey) -> "+ZeroCosmosAddr+"\n")
}
