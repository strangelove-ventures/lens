# `lens`

`lens` is your lens to view the Cosmos :atom:. `lens` packs all the best practices in golang cosmos client development into one place and provides a simple and easy to use APIs provided by standard Cosmos chains. The `cmd` package implements the `lens` command line tool while the `client` package contains all the building blocks to build your own, complex, feature rich, Cosmos client in go.

Intended use cases:
- Trading Bots
- Server side query integrations for alerting or other automation
- Indexers and search engines
- Transaction automation (`x/authz` for security) for farming applications
- ...:atom::rocket::moon:

This is the start of ideas around how to implement the cosmos client libraries in a seperate repo.

## Tutorial

### CMD Line

Lens provides a cmd line tool to  interact with any cosmos chain supporting the core [Cosmos-SDK modules](https://github.com/cosmos/cosmos-sdk/tree/master/x).

#### Install

```
git clone https://github.com/strangelove-ventures/lens.git

cd lens

make install
```

After running the above commands, when running `lens` you should see

```
❯ lens            
This is my lens, there are many like it, but this one is mine.

Usage:
  lens [command]

Available Commands:
  chains      manage local chain configurations
  keys        manage keys held by the relayer for each chain
  query       query things about a chain
  tendermint  all tendermint query commands
  tx          query things about a chain
  help        Help about any command

Flags:
      --chain string   override default chain
  -d, --debug          debug output
  -h, --help           help for lens
      --home string    set home directory (default "/Users/lenscrafters/.lens")

Use "lens [command] --help" for more information about a command.
```
#### Chains

Lens comes with two defaulted configs, Cosmos Hub and Osmosis. Located at `~/.lens/config.toml` 

```
default_chain: osmosis
 chains:
   cosmoshub:
     key: default
     chain-id: cosmoshub-4
     rpc-addr: https://cosmoshub-4.technofractal.com:443
     grpc-addr: https://gprc.cosmoshub-4.technofractal.com:443
     account-prefix: cosmos
     keyring-backend: test
     gas-adjustment: 1.2
     gas-prices: 0.01uatom
     key-directory: /Users/lenscrafters/.lens/keys
     debug: false
     timeout: 20s
     output-format: json
     sign-mode: direct
   osmosis:
     key: default
     chain-id: osmosis-1
     rpc-addr: https://osmosis-1.technofractal.com:443
     grpc-addr: https://gprc.osmosis-1.technofractal.com:443
     account-prefix: osmo
     keyring-backend: test
     gas-adjustment: 1.2
     gas-prices: 0.01uosmo
     key-directory: /Users/lenscrafters/.lens/keys
     debug: false
     timeout: 20s
     output-format: json
     sign-mode: direct
	```

To add more to your config you can run. 

```
lens chains add juno
```

This command checks the chain registry located [here](https://github.com/cosmos/chain-registry), for the requested chain. After running the command your `config.toml` should look like

```
default_chain: osmosis
 chains:
   cosmoshub:
     key: default
     chain-id: cosmoshub-4
     rpc-addr: https://cosmoshub-4.technofractal.com:443
     grpc-addr: https://gprc.cosmoshub-4.technofractal.com:443
     account-prefix: cosmos
     keyring-backend: test
     gas-adjustment: 1.2
     gas-prices: 0.01uatom
     key-directory: /Users/lenscrafters/.lens/keys
     debug: false
     timeout: 20s
     output-format: json
     sign-mode: direct
   juno:
     key: default
     chain-id: juno-1
     rpc-addr: https://rpc-juno.itastakers.com:443
     grpc-addr: ""
     account-prefix: juno
     keyring-backend: test
     gas-adjustment: 1.2
     gas-prices: 0.01ujuno
     key-directory: /Users/lenscrafters/.lens/keys
     debug: false
     timeout: 20s
     output-format: json
     sign-mode: direct
   osmosis:
     key: default
     chain-id: osmosis-1
     rpc-addr: https://osmosis-1.technofractal.com:443
     grpc-addr: https://gprc.osmosis-1.technofractal.com:443
     account-prefix: osmo
     keyring-backend: test
     gas-adjustment: 1.2
     gas-prices: 0.01uosmo
     key-directory: /Users/lenscrafters/.lens/keys
     debug: false
     timeout: 20s
     output-format: json
     sign-mode: direct
	```

When running a command, it will run the command for the defaulted chain. The defaulted chain can be found at the top of `~/.lens/config.toml` or by running `lens`. 

To change your default, run: 

``` 
lens chains set-default <chain_name>
```

#### Keys

Lens uses the keyring from the Cosmos-sdk. You can read more about it [here](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keyring/doc.go). To add your key to lens run:

``` 
lens keys restore <key_name> '<mnemonic>'
```

After this when running `lens keys list` you should see your defaulted chains address, if this is not changed you will see the cosmos-hub address. To see your key encoded for use on other chains run `lens keys enumerate <key_name>`. 
