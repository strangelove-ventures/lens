package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jhump/protoreflect/desc"
)

var _ error = ChainNotFoundError{}

// ChainNotFoundError is used when a requested chain does not exist.
// Its error message includes the list of known chains.
type ChainNotFoundError struct {
	Requested string
	Config    *Config
}

func (e ChainNotFoundError) Error() string {
	available := make([]string, 0, len(e.Config.Chains))
	for chainName := range e.Config.Chains {
		available = append(available, chainName)
	}
	sort.Strings(available)

	return fmt.Sprintf(
		"no chain %q found (available chains: %s)",
		e.Requested,
		strings.Join(available, ", "),
	)
}

var _ error = GRPCServiceNotFoundError{}

// GRPCServiceNotFoundError is used when a requested gRPC service does not exist.
// Its error message includes the provided available services.
type GRPCServiceNotFoundError struct {
	Requested string
	Available []string
}

func (e GRPCServiceNotFoundError) Error() string {
	sort.Strings(e.Available)
	// TODO: would be nice to suggest close matches here.
	return fmt.Sprintf(
		"no service %q found (available services: %s)",
		e.Requested,
		strings.Join(e.Available, ", "),
	)
}

var _ error = GRPCMethodNotFoundError{}

// GRPCMethodNotFoundError is used when a requested gRPC method does not exist.
// Its error message includes the provided available services.
type GRPCMethodNotFoundError struct {
	TargetService string
	Requested     string
	Available     []*desc.MethodDescriptor
}

func (e GRPCMethodNotFoundError) Error() string {
	methodNames := make([]string, len(e.Available))
	for i, md := range e.Available {
		methodNames[i] = md.GetName()
	}
	sort.Strings(methodNames)

	return fmt.Sprintf(
		"service %q has no method with name %q (available methods: %s)",
		e.TargetService,
		e.Requested,
		strings.Join(methodNames, ", "),
	)
}
