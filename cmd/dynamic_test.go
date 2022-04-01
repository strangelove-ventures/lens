package cmd_test

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func TestDynamicInspect_ChainID(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	_ = sys.MustRun(t, "chains", "edit", "cosmoshub", "grpc-addr", gRPCAddr)

	res := sys.MustRun(t, "dynamic", "inspect", "cosmoshub")
	require.Equal(t, res.Stdout.String(), "grpc.channelz.v1.Channelz\ngrpc.reflection.v1alpha.ServerReflection\n")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicInspect_AddressLiteral(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	res := sys.MustRun(t, "dynamic", "inspect", gRPCAddr)
	require.Equal(t, res.Stdout.String(), "grpc.channelz.v1.Channelz\ngrpc.reflection.v1alpha.ServerReflection\n")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicInput_SecureOnly(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	// Connection refuses with --secure-only flag set.
	res := sys.Run(zaptest.NewLogger(t), "dynamic", "inspect", "--secure-only", gRPCAddr)
	require.Error(t, res.Err)
	require.Empty(t, res.Stdout.String())
	require.Contains(t, res.Stderr.String(), "failed to dial gRPC address")
}

func TestDynamicInspectService(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	res := sys.MustRun(t, "dynamic", "inspect", gRPCAddr, "grpc.channelz.v1.Channelz")
	// Just check two arbitrary RPC definition lines.
	require.Contains(t, res.Stdout.String(), "rpc GetServer ( .grpc.channelz.v1.GetServerRequest ) returns ( .grpc.channelz.v1.GetServerResponse );\n\n")
	require.Contains(t, res.Stdout.String(), "rpc GetSocket ( .grpc.channelz.v1.GetSocketRequest ) returns ( .grpc.channelz.v1.GetSocketResponse );\n\n")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicInspectMethod(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	res := sys.MustRun(t, "dynamic", "inspect", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServer")
	// Includes the request and response messages.
	require.Contains(t, res.Stdout.String(), "message GetServerRequest {\n")
	require.Contains(t, res.Stdout.String(), "message GetServerResponse {\n")
	// Includes transitive message definitions, with a leading comment of the fully qualified name and filename.
	require.Contains(t, res.Stdout.String(), "// grpc.channelz.v1.SubchannelRef (grpc/channelz/v1/channelz.proto)\nmessage SubchannelRef {")
	// Includes enums, with a leading comment of the fully qualified name and filename.
	require.Contains(t, res.Stdout.String(), "// grpc.channelz.v1.ChannelTraceEvent.Severity (grpc/channelz/v1/channelz.proto)\nenum Severity {")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicQuery_ChainID(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	_ = sys.MustRun(t, "chains", "edit", "cosmoshub", "grpc-addr", gRPCAddr)

	// ServerSockets will be empty since this is a new gRPC server
	// that has no other connections.
	res := sys.MustRun(t, "dynamic", "query", "cosmoshub", "grpc.channelz.v1.Channelz", "GetServerSockets")
	require.Equal(t, res.Stdout.String(), `{"end":true}`+"\n")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicQuery_AddressLiteral(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	// ServerSockets will be empty since this is a new gRPC server
	// that has no other connections.
	res := sys.MustRun(t, "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServerSockets")
	require.Equal(t, res.Stdout.String(), `{"end":true}`+"\n")
	require.Empty(t, res.Stderr.String())
}

func TestDynamicQuery_SecureOnly(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	// Connection refuses with --secure-only flag set.
	res := sys.Run(zaptest.NewLogger(t), "dynamic", "query", "--secure-only", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServerSockets")
	require.Error(t, res.Err)
	require.Empty(t, res.Stdout.String())
	require.Contains(t, res.Stderr.String(), "failed to dial gRPC address")
}

func TestDynamicQuery_Response(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	// GetServers returns some details about the active server.
	res := sys.MustRun(t, "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServers")
	require.Empty(t, res.Stderr.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &resp))

	// Response should look like this:
	// {"server":[{"ref":{"serverId":"1"},"data":{"callsStarted":"12","callsSucceeded":"10","lastCallStartedTimestamp":"2022-03-30T12:23:35.608849Z"},"listenSocket":[{"socketId":"2","name":"127.0.0.1:58762"}]}],"end":true}
	// Make a couple general assertions on its shape.

	val, ok := resp["server"]
	require.True(t, ok, "expected response to have top-level key 'server'")

	_, ok = val.([]interface{})
	require.True(t, ok, "expected response's server field to be an array of objects, got %T", val)
}

func TestDynamicQuery_InputVariations(t *testing.T) {
	// This test is NOT using parallel
	// because querying the servers will pick up a server ID
	// of a different goroutine, which may disappear before this test completes.

	sys := NewSystem(t)

	gRPCAddr := runGRPCReflectionServer(t)

	// During parallel tests, the reported server ID is random,
	// so get the server ID by querying the current servers.
	res := sys.MustRun(t, "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServers")

	var serversResp struct {
		Server []struct {
			Ref struct {
				ServerID string
			}
		}
	}

	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &serversResp))
	require.NotEmpty(t, serversResp.Server)

	serverID := serversResp.Server[0].Ref.ServerID
	input := fmt.Sprintf(`{"server_id":%s}`, serverID)
	wantResp := fmt.Sprintf(`{"serverId":%q}`, serverID)

	t.Run("explicit argument", func(t *testing.T) {
		res := sys.MustRun(t, "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServer", input)
		require.Contains(t, res.Stdout.String(), wantResp)
		require.Empty(t, res.Stderr.String())
	})

	t.Run("@file argument", func(t *testing.T) {
		tmpdir := t.TempDir()
		f, err := os.CreateTemp(tmpdir, "")
		require.NoError(t, err)
		f.Close()
		require.NoError(t, os.WriteFile(f.Name(), []byte(input), 0600))

		res := sys.MustRun(t, "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServer", "@"+f.Name())
		require.Contains(t, res.Stdout.String(), wantResp)
		require.Empty(t, res.Stderr.String())
	})

	t.Run("stdin flag", func(t *testing.T) {
		res := sys.MustRunWithInput(t, strings.NewReader(input), "dynamic", "query", gRPCAddr, "grpc.channelz.v1.Channelz", "GetServer", "--stdin")
		require.Contains(t, res.Stdout.String(), wantResp)
		require.Empty(t, res.Stderr.String())
	})
}

func runGRPCReflectionServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	reflection.Register(srv)                         // Required for reflection.
	channelzsvc.RegisterChannelzServiceToServer(srv) // Arbitrary other built-in gRPC service to confirm reflection behavior.
	go func() {
		srv.Serve(ln)
	}()
	t.Cleanup(srv.Stop)

	return ln.Addr().String()
}
