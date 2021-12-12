# `lens`

`lens` is your lens to view the Cosmos :atom:. `lens` packs all the best practices in golang cosmos client developemnet into one place and provides a simple and easy to use APIs provided by standard Cosmos chains. The `cmd` package implementes the `lens` command line tool while the `client` package contains all the building blocks to build your own, complex, feature rich, Cosmos client in go.

Intended use cases:
- Trading Bots
- Server side query integrations for alerting or other automation
- Indexers and search engines

This is the start of ideas around how to implement the cosmos client libraries in a seperate repo

- [ ] `lens keys` to test key integration
- [ ] `lens query` to test query integration
- [ ] `lens tx` to test tx integration
- [ ] Switch `lens` from using `var config *Config` as a global variable to a `lens.Config` struct pulled out `context` in a similar manner to the way the SDK works. This should allow for generic reuse of the `cmd` package to quickly build a new client. with standard functionality.
- [ ] How to instantiate and use the GRPC golang client? This is not not currently obvious.
- [ ] Currently we are depending on the sdk keyring libs. This is fine for now but eventually the goal is to support our own keyring implemenation that is specifically for `lens`
- [ ] Biggest TODO: is transaction generation and signing