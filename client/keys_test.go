package client_test

import (
	"testing"

	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap/zaptest"
)

// TestKeyRestore restores a test mnemonic
func TestKeyRestore(t *testing.T) {
	keyName := "test_key"
	mnemonic := "blind master acoustic speak victory lend kiss grab glad help demand hood roast zone lend sponsor level cheap truck kingdom apology token hover reunion"
	expectedAddress := "cosmos15cw268ckjj2hgq8q3jf68slwjjcjlvxy57je2u"
	var coinType uint32
	coinType = 118 // Cosmos coin type used in address derivation

	homepath := t.TempDir()
	cl, err := client.NewChainClient(
		zaptest.NewLogger(t),
		client.GetCosmosHubConfig(homepath, true),
		homepath, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = cl.DeleteKey(keyName) // Delete if test is being run again
	address, err := cl.RestoreKey(keyName, mnemonic, coinType)
	if err != nil {
		t.Fatalf("Error while restoring mnemonic: %v", err)
	}
	if address != expectedAddress {
		t.Fatalf("Restored address: %s does not match expected: %s", address, expectedAddress)
	}
}
