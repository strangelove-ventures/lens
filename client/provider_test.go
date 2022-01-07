package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTrustingPeriod(t *testing.T) {
	hubClient := GetTestClient()
	tp, err := hubClient.TrustingPeriod()
	fmt.Println(tp)
	require.NoError(t, err)
}
