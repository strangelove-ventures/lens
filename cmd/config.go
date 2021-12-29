package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client"
	"gopkg.in/yaml.v2"
)

//createConfig idempotently creates the config.
func createConfig(home string, debug bool) error {
	cfgPath := path.Join(home, "config.yaml")

	// If the config doesn't exist...
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// And the config folder doesn't exist...
		// And the home folder doesn't exist
		if _, err := os.Stat(home); os.IsNotExist(err) {
			// Create the home folder
			if err = os.Mkdir(home, os.ModePerm); err != nil {
				return err
			}
		}
	}

	// Then create the file...
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// And write the default config to that location...
	if _, err = f.Write(defaultConfig(path.Join(home, "keys"), debug)); err != nil {
		return err
	}
	return nil
}

// Command for inititalizing an empty config at the --home location
func configInitCmd() *cobra.Command {
	// TODO: add a `--chain` flag here to specify which chain to initialize
	// this should reference standard configs in an `interchain` directory
	// similar to the one in the relayer
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "Creates a default home directory at path defined by --home",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}

			return createConfig(home, debug)
		},
	}
	return cmd
}

// Config represents the config file for the relayer
type Config struct {
	DefaultChain string                               `yaml:"default_chain" json:"default_chain"`
	Chains       map[string]*client.ChainClientConfig `yaml:"chains" json:"chains"`

	cl map[string]*client.ChainClient
}

func (c *Config) GetDefaultClient() *client.ChainClient {
	return c.GetClient(c.DefaultChain)
}

func (c *Config) GetClient(chainID string) *client.ChainClient {
	if v, ok := c.cl[chainID]; ok {
		return v
	}
	return nil
}

// Called to initialize the relayer.Chain types on Config
func validateConfig(c *Config) error {
	for _, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return err
		}
	}
	if c.GetDefaultClient() == nil {
		return fmt.Errorf("default chain (%s) configuration not found", c.DefaultChain)
	}
	return nil
}

// MustYAML returns the yaml string representation of the Paths
func (c Config) MustYAML() []byte {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return out
}

func defaultConfig(keyHome string, debug bool) []byte {
	modules := []module.AppModuleBasic{}
	for _, v := range simapp.ModuleBasics {
		modules = append(modules, v)
	}
	cfg := Config{
		DefaultChain: "cosmoshub",
		Chains: map[string]*client.ChainClientConfig{
			"cosmoshub": {
				Key:            "default",
				ChainID:        "cosmoshub-4",
				RPCAddr:        "https://cosmoshub-4.technofractal.com:443",
				GRPCAddr:       "https://gprc.cosmoshub-4.technofractal.com:443",
				AccountPrefix:  "cosmos",
				KeyringBackend: "test",
				GasAdjustment:  1.2,
				GasPrices:      "0.01uatom",
				KeyDirectory:   keyHome,
				Debug:          debug,
				Timeout:        "20s",
				OutputFormat:   "json",
				BroadcastMode:  "block",
				SignModeStr:    "direct",
				Modules:        modules,
			},
			"osmosis": {
				Key:            "default",
				ChainID:        "osmosis-1",
				RPCAddr:        "https://osmosis-1.technofractal.com:443",
				GRPCAddr:       "https://gprc.osmosis-1.technofractal.com:443",
				AccountPrefix:  "osmo",
				KeyringBackend: "test",
				GasAdjustment:  1.2,
				GasPrices:      "0.01uosmo",
				KeyDirectory:   keyHome,
				Debug:          debug,
				Timeout:        "20s",
				OutputFormat:   "json",
				BroadcastMode:  "block",
				SignModeStr:    "direct",
				Modules:        modules,
			},
		},
	}
	return cfg.MustYAML()
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}

	config = &Config{}
	cfgPath := path.Join(home, "config.yaml")
	_, err = os.Stat(cfgPath)
	if err != nil {
		err = createConfig(home, debug)
		if err != nil {
			return err
		}
	}
	viper.SetConfigFile(cfgPath)
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read in config:", err)
		os.Exit(1)
	}

	// read the config file bytes
	file, err := ioutil.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// unmarshall them into the struct
	if err = yaml.Unmarshal(file, config); err != nil {
		fmt.Println("Error unmarshalling config:", err)
		os.Exit(1)
	}

	// instantiate chain client
	// TODO: this is a bit of a hack, we should probably have a
	// better way to inject modules into the client
	config.cl = make(map[string]*client.ChainClient)
	modules := []module.AppModuleBasic{}
	for _, v := range simapp.ModuleBasics {
		modules = append(modules, v)
	}
	for name, chain := range config.Chains {
		chain.Modules = modules
		cl, err := client.NewChainClient(chain, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Println("Error creating chain client:", err)
			os.Exit(1)
		}
		config.cl[name] = cl
	}

	// validate configuration
	if err = validateConfig(config); err != nil {
		fmt.Println("Error parsing chain config:", err)
		os.Exit(1)
	}
	return nil
}
