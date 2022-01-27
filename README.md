# `lens`

**Lens provides a cmd line tool to  interact with any cosmos chain supporting the core [Cosmos-SDK modules](https://github.com/cosmos/cosmos-sdk/tree/master/x).**

---

`lens` is your lens to view the Cosmos :atom:. `lens` packs all the best practices in golang cosmos client development into one place and provides a simple and easy to use APIs provided by standard Cosmos chains. The `cmd` package implements the `lens` command line tool while the `client` package contains all the building blocks to build your own, complex, feature rich, Cosmos client in go.

Intended use cases:
- Trading Bots
- Server side query integrations for alerting or other automation
- Indexers and search engines
- Transaction automation (`x/authz` for security) for farming applications
- ...:atom::rocket::moon:

This is the start of ideas around how to implement the cosmos client libraries in a seperate repo.

## **--INSTALL--**
---
```bash
git clone https://github.com/strangelove-ventures/lens.git

cd lens

make install
```
Now run:
```bash
lens
```
You should see:

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

## **--CONFIG--**
---
The config file tells lens how to interact with blockchains. This is where information such as grpc address's and chain-ids are held.

Config File Location: `~/.lens/config.yaml` 

> NOTE: The config file is not created at install, it is created the first time lens needs to query your config. Just to get it created, you can run something like:
>```
>lens chains show-default
>```

### **CHAINS**
Lens comes with two defaulted chains. Cosmos Hub and Osmosis.

To interact with other chains, you need to add them to your config, run:

```
lens chains add <chain_name>
#Example:
lens chains add juno
```

To view all possible chain names, run:
```
lens chains registry-list
```
> NOTE: These two commands check the chain registry located [here](https://github.com/cosmos/chain-registry), for the requested chain.

When running a command, it will run the command for the defaulted chain.

To view your default chain, run:

```
 lens chains show-default
```

To change your default, run: 

``` 
lens chains set-default <chain_name>
```

An example config file:

```
default_chain: cosmoshub
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
    key-directory: /Users/dan/.lens/keys
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
    key-directory: /Users/dan/.lens/keys
    debug: false
    timeout: 20s
    output-format: json
    sign-mode: direct
```

### **Keys**
Lens uses the keyring from the Cosmos-sdk. There is more information about it keyring [here](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keyring/doc.go). 

To add a key to lens you have two options:

* `lens keys add` - This will add a key to you default chain and name it "default". You can optionally add a name as an argument. 
* `lens keys restore <name>` - This will restore a key to your default chain. Replace '\<name\>' with a key name. This command will THEN ask for your mnemonic which is needed to restore and use it to boradcast transactions. 

>❗️ NOTE: IF you name your key anything other than "default", you will need to manually change the `key:` value in your config to link key with chain.  
>```bash
>default_chain: cosmoshub
>chains:
>  cosmoshub:
>    key: default #CHANGE THIS NAME
>    chain-id: cosmoshub-4
>    rpc-addr: https://cosmoshub-4.technofractal.com:443
>    ...
>```

After adding or restoring a key, it should appear in your list by running: `lens keys list`, by default it will show the Cosmos Hub address. 

To see the key encoded for use on other chains run `lens keys enumerate <key_name>`. 


## --QUICKSTART EXAMPLES--
---

