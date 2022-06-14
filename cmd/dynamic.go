package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"

)

func dynamicCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dynamic",
		Aliases: []string{"dyn"},
		Short:   "Dynamic integration with remote chains",
	}

	cmd.AddCommand(
		dynInspectCmd(a),
		dynQueryCmd(a),
	)

	return cmd
}

func dynQueryCmd(a *appState) *cobra.Command {
	const stdinFlag = "stdin"
	const heightFlag = "height"
	var height int64

	cmd := &cobra.Command{
		Use:     "query CHAIN_NAME_OR_GRPC_ADDR SERVICE_NAME METHOD_NAME [INPUT_OBJECT|@PATH_TO_INPUT_FILE]",
		Aliases: []string{"q"},
		Short:   "Use gRPC reflection to dynamically interact with arbitrary services on the chain",
		Long: fmt.Sprintf(`Execute a gRPC request against the remote server.

All gRPC requests require an input.
To determine the format of the input, use '%[1]s dynamic query'.

There are multiple ways to provide the input.
If you don't specify anything explicitly, the input defaults to an empty object.
Many gRPC services accept input without any extra arguments.

Otherwise, you can specify the input as an explicit command line argument:
    %[1]s dyn q my-chain cosmos.base.tendermint.v1beta1.Service GetBlockByHeight '{"height": 2222222}'

Or you can use the --stdin flag if you have another process to generate the input:
    account_input_arg.sh | %[1]s dyn q my-chain cosmos.bank.v1beta1.Query Balance --stdin

Finally, you can use an '@' prefix if you want to read from a file:
    %[1]s dyn q my-chain cosmos.bank.v1beta1.Query Balance @my_account.json
`,
			appName),
		Args: withUsage(cobra.RangeArgs(3, 4)),
		Example: fmt.Sprintf(`$ %[1]s dynamic query example.com:9090 cosmos.bank.v1beta1.Query TotalSupply
$ %[1]s dynamic q my-chain cosmos.base.tendermint.v1beta1.Service GetBlockByHeight '{"height": 2222222}'
$ %[1]s dynamic q my-chain cosmos.base.tendermint.v1beta1.Service GetBlockByHeight @path/to/input.json
$ echo '{"validator_address": "..."}' | %[1]s dyn q my-chain cosmos.distribution.v1beta1.Query ValidatorOutstandingRewards --stdin`,
			appName),
		RunE: func(cmd *cobra.Command, args []string) error {
			gRPCAddr, err := chooseGRPCAddr(a, args[0])
			if err != nil {
				return err
			}
			serviceName := args[1]
			if serviceName == "" {
				return fmt.Errorf("service name may not be empty")
			}
			methodName := args[2]
			if methodName == "" {
				return fmt.Errorf("method name may not be empty")
			}

			var in []byte
			if len(args) > 3 {
				if strings.HasPrefix(args[3], "@") {
					// @file format.
					name := strings.TrimPrefix(args[3], "@")
					in, err = os.ReadFile(name)
					if err != nil {
						return fmt.Errorf("failed to read input file: %w", err)
					}
				} else {
					// Provided explicit value on command line.
					in = []byte(args[3])
				}
			} else if useStdin, _ := cmd.Flags().GetBool(stdinFlag); useStdin {
				// Provided --stdin.
				in, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("error reading from stdin: %w", err)
				}
			} else {
				// Didn't provide command line argument and didn't use --stdin.
				// Default to empty object for input.
				in = []byte("{}")
			}
			return dynamicQuery(cmd, a, gRPCAddr, serviceName, methodName, in, height)
		},
	}

	cmd = gRPCFlags(cmd, a.Viper)
	cmd.Flags().Bool(stdinFlag, false, "read input from stdin instead of as command-line argument")
	cmd.Flags().Int64Var(&height, heightFlag, 0, "specify the height for the query or use latest")
	return cmd
}

func dynamicQuery(cmd *cobra.Command, a *appState, gRPCAddr, serviceName, methodName string, input []byte, height int64) error {
	conn, err := dialGRPC(cmd, a, gRPCAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	stub := rpb.NewServerReflectionClient(conn)
	c := grpcreflect.NewClient(cmd.Context(), stub)
	defer c.Reset()

	svcDesc, err := c.ResolveService(serviceName)
	if err != nil {
		if grpcreflect.IsElementNotFoundError(err) {
			// If we can list the available services, return a more useful error.
			services, svcErr := c.ListServices()
			if svcErr == nil {
				return GRPCServiceNotFoundError{
					Requested: serviceName,
					Available: services,
				}
			}
		}

		return fmt.Errorf("failed to resolve service %q: %w", serviceName, err)
	}

	methodDesc := svcDesc.FindMethodByName(methodName)
	if methodDesc == nil {
		return GRPCMethodNotFoundError{
			TargetService: serviceName,
			Requested:     methodName,
			Available:     svcDesc.GetMethods(),
		}
	}

	inMsgDesc := methodDesc.GetInputType() // TODO: check for nil input type?
	inputMsg := dynamic.NewMessage(inMsgDesc)

	if err := inputMsg.UnmarshalJSON(input); err != nil {
		return fmt.Errorf("failed to marshal input into message of type %s: %w", inMsgDesc.GetFullyQualifiedName(), err)
	}

	dynClient := grpcdynamic.NewStub(conn)
	if methodDesc.IsClientStreaming() || methodDesc.IsServerStreaming() {
		return fmt.Errorf("TODO: handle client/server streaming")
	}

	md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
	ctx := metadata.NewOutgoingContext(cmd.Context(), md)
	output, err := dynClient.InvokeRpc(ctx, methodDesc, inputMsg)
	if err != nil {
		return fmt.Errorf("failed to invoke rpc: %w", err)
	}

	// Convert to a dynamic message, so that we can use the AnyResolver
	// based on the client that can resolve not-yet-known messages.
	dynOutput, err := dynamic.AsDynamicMessage(output)
	if err != nil {
		return fmt.Errorf("failed to convert output to dynamic message: %w", err)
	}
	j, err := dynOutput.MarshalJSONPB(&jsonpb.Marshaler{
		// For Any fields, resolve through the client.
		AnyResolver: reflectClientAnyResolver{c: c},
	})
	if err != nil {
		return fmt.Errorf("failed to serialize output message: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(j))
	return nil
}

func dynInspectCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect CHAIN_NAME_OR_GRPC_ADDR [SERVICE_NAME [METHOD_NAME]]",
		Aliases: []string{"i"},
		Short:   "Use gRPC reflection to see protobuf definitions of services or methods",
		Args:    withUsage(cobra.RangeArgs(1, 3)),
		Example: fmt.Sprintf(`$ %s dynamic inspect example.com:9090
$ %s dynamic i my-chain
$ %s dyn i my-chain cosmos.bank.v1beta1.Query TotalSupply`,
			appName, appName, appName),
		RunE: func(cmd *cobra.Command, args []string) error {
			gRPCAddr, err := chooseGRPCAddr(a, args[0])
			if err != nil {
				return err
			}

			var serviceName, methodName string
			if len(args) > 1 {
				serviceName = args[1]
			}
			if len(args) > 2 {
				methodName = args[2]
			}

			a.Log.Debug("Inspecting server", zap.String("addr", gRPCAddr))

			return dynamicInspect(cmd, a, gRPCAddr, serviceName, methodName)
		},
	}

	return gRPCFlags(cmd, a.Viper)
}

func dynamicInspect(cmd *cobra.Command, a *appState, gRPCAddr, serviceName, methodName string) error {
	conn, err := dialGRPC(cmd, a, gRPCAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	stub := rpb.NewServerReflectionClient(conn)
	c := grpcreflect.NewClient(cmd.Context(), stub)
	defer c.Reset()

	pp := &protoprint.Printer{
		SortElements:             true,
		ForceFullyQualifiedNames: true,
	}

	if serviceName == "" {
		a.Log.Debug("Listing all services")

		services, err := c.ListServices()
		if err != nil {
			return fmt.Errorf("failed to list remote services: %w", err)
		}

		for _, svc := range services {
			svcDesc, err := c.ResolveService(svc)
			if err != nil {
				a.Log.Info(
					"Error resolving service",
					zap.String("service_name", svc),
					zap.Error(err),
				)
				continue
			}
			fmt.Fprintln(cmd.OutOrStdout(), svcDesc.GetFullyQualifiedName())
		}

		return nil
	}

	a.Log.Debug("Resolving requested service", zap.String("service_name", serviceName))
	svcDesc, err := c.ResolveService(serviceName)
	if err != nil {
		if grpcreflect.IsElementNotFoundError(err) {
			// Return a descriptive error if we can't resolve the given service
			// but can list the other services.
			services, err := c.ListServices()
			if err == nil {
				return GRPCServiceNotFoundError{
					Requested: serviceName,
					Available: services,
				}
			}
		}

		a.Log.Info(
			"Error resolving service",
			zap.Error(err),
		)
		return err
	}

	if methodName == "" {
		proto, err := pp.PrintProtoToString(svcDesc)
		if err != nil {
			a.Log.Info(
				"Error converting to proto string",
				zap.String("service_name", svcDesc.GetFullyQualifiedName()),
				zap.Error(err),
			)
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), proto)

		return nil
	}

	a.Log.Debug("Resolving requested method", zap.String("service_name", serviceName), zap.String("method_name", methodName))
	mDesc := svcDesc.FindMethodByName(methodName)
	if mDesc == nil {
		return GRPCMethodNotFoundError{
			TargetService: serviceName,
			Requested:     methodName,
			Available:     svcDesc.GetMethods(),
		}
	}

	proto, err := pp.PrintProtoToString(mDesc)
	if err != nil {
		a.Log.Info(
			"Error converting to proto string",
			zap.String("service_name", svcDesc.GetFullyQualifiedName()),
			zap.String("method_name", mDesc.GetFullyQualifiedName()),
			zap.Error(err),
		)
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), proto)

	var s sources
	if inType := mDesc.GetInputType(); inType != nil {
		proto, err := pp.PrintProtoToString(inType)
		if err != nil {
			a.Log.Info(
				"Error converting method input type to string",
				zap.String("service_name", svcDesc.GetFullyQualifiedName()),
				zap.String("method_name", mDesc.GetFullyQualifiedName()),
				zap.Error(err),
			)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "// "+inType.GetFile().GetFullyQualifiedName())
			fmt.Fprintln(cmd.OutOrStdout(), proto)

			s = walkMessageType(inType, s)
		}
	}

	if outType := mDesc.GetOutputType(); outType != nil {
		proto, err := pp.PrintProtoToString(outType)
		if err != nil {
			a.Log.Info(
				"Error converting method output type to string",
				zap.String("service_name", svcDesc.GetFullyQualifiedName()),
				zap.String("method_name", mDesc.GetFullyQualifiedName()),
				zap.Error(err),
			)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "// "+outType.GetFile().GetFullyQualifiedName())
			fmt.Fprintln(cmd.OutOrStdout(), proto)

			s = walkMessageType(outType, s)
		}
	}

	s.Print(a.Log, cmd.OutOrStdout(), pp)

	return nil
}

// sources is a collection of descriptors.
// It is a slice so iteration order when printing is maintained.
type sources []desc.Descriptor

// Contains iterates through the existing sources
// and reports whether it already contains a descriptor matching d's fully qualified name.
func (s sources) Contains(d desc.Descriptor) bool {
	want := d.GetFullyQualifiedName()
	for _, have := range s {
		if have.GetFullyQualifiedName() == want {
			return true
		}
	}

	return false
}

func (s sources) Print(log *zap.Logger, out io.Writer, pp *protoprint.Printer) {
	for _, desc := range s {
		proto, err := pp.PrintProtoToString(desc)
		if err != nil {
			log.Info(
				"Error converting descriptor to string",
				zap.String("fully_qualified_name", desc.GetFullyQualifiedName()),
				zap.Error(err),
			)
			continue
		}

		fmt.Fprintf(out, "// %s (%s)\n", desc.GetFullyQualifiedName(), desc.GetFile().GetFullyQualifiedName())
		fmt.Fprintln(out, proto)
	}
}

func walkMessageType(msgDesc *desc.MessageDescriptor, s sources) sources {
	for _, fDesc := range msgDesc.GetFields() {
		if mDesc := fDesc.GetMessageType(); mDesc != nil {
			if !s.Contains(mDesc) {
				s = append(s, mDesc)
				s = walkMessageType(mDesc, s)
			}

			continue
		}

		if eDesc := fDesc.GetEnumType(); eDesc != nil {
			if !s.Contains(eDesc) {
				s = append(s, eDesc)
				// Enums are just lists of constants, so no need to descend into them.
			}

			continue
		}
	}

	return s
}

func dialGRPC(cmd *cobra.Command, a *appState, addr string) (*grpc.ClientConn, error) {
	requireSecure, err := cmd.Flags().GetBool(gRPCSecureOnlyFlag)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	if !requireSecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	a.Log.Debug("Opening remote gRPC connection", zap.String("addr", addr))
	conn, err := grpc.DialContext(cmd.Context(), addr, dialOpts...)
	if err != nil {
		if requireSecure && strings.Contains(err.Error(), "grpc: no transport security set") {
			// Have to use string matching for unexported grpc.errNoTransportSecurity error value.
			a.Log.Warn("Refusing to connect to non-TLS server when --" + gRPCSecureOnlyFlag + " flag set")
		}
		return nil, fmt.Errorf("failed to dial gRPC address %q: %w", addr, err)
	}

	return conn, nil
}

type reflectClientAnyResolver struct {
	c *grpcreflect.Client
}

var _ jsonpb.AnyResolver = reflectClientAnyResolver{}

func (r reflectClientAnyResolver) Resolve(typeURL string) (proto.Message, error) {
	// Unclear if it is always safe to trim the leading slash here.
	typeURL = strings.TrimPrefix(typeURL, "/")
	messageDesc, err := r.c.ResolveMessage(typeURL)
	if err != nil {
		return nil, err
	}

	return dynamic.NewMessage(messageDesc), nil
}

func chooseGRPCAddr(a *appState, addrOrChainName string) (string, error) {
	if _, _, err := net.SplitHostPort(addrOrChainName); err == nil {
		// Argument looks like a host:port, so just return that value.
		return addrOrChainName, nil
	}

	chain, ok := a.Config.Chains[addrOrChainName]
	if !ok {
		return "", fmt.Errorf("%q did not look like host:port and no chain exists by that name", addrOrChainName)
	}

	gRPCAddr := chain.GRPCAddr
	if gRPCAddr == "" {
		return "", fmt.Errorf("no gRPC address set for chain %q", addrOrChainName)
	}

	return gRPCAddr, nil
}
