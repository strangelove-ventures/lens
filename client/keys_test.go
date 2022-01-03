package client

import (
	"testing"
)

// TestKeyRestore restores a test mnemonic 
// for a valid return value.
func TestKeyRestore(t *testing.T) {
	keyName := "test_key"
	mnemonic := "blind master acoustic speak victory lend kiss grab glad help demand hood roast zone lend sponsor level cheap truck kingdom apology token hover reunion"
	expectedAddress := "cosmos15cw268ckjj2hgq8q3jf68slwjjcjlvxy57je2u"

	cl, _ := NewChainClient(GetCosmosHubConfig("/tmp", true), nil, nil)
	_ = cl.DeleteKey(keyName) // Delete if test is being run again
	address, err := cl.RestoreKey(keyName, mnemonic)
	if err != nil {
		t.Fatalf("Error while restoring mnemonic: %v", err)
	}
	if address != expectedAddress {
		t.Fatalf("Restored address: %s does not match expected: %s", address, expectedAddress)
	}
}
