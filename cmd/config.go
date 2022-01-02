package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
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

func overwriteConfig(home string, cfg *Config) error {
	cfgPath := path.Join(home, "config.yaml")
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(cfg.MustYAML()); err != nil {
		return err
	}
	return nil
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
	return Config{
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
				SignModeStr:    "direct",
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
				SignModeStr:    "direct",
			},
		},
	}.MustYAML()
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
	for name, chain := range config.Chains {
		chain.Modules = append([]module.AppModuleBasic{}, ModuleBasics...)
		cl, err := client.NewChainClient(chain, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Println("Error creating chain client:", err)
			os.Exit(1)
		}
		config.cl[name] = cl
	}

	// override chain if needed

	if cmd.PersistentFlags().Changed("chain") {
		defaultChain, err := cmd.PersistentFlags().GetString("chain")
		if err != nil {
			return err
		}

		config.DefaultChain = defaultChain
	}

	// validate configuration
	if err = validateConfig(config); err != nil {
		fmt.Println("Error parsing chain config:", err)
		os.Exit(1)
	}
	return nil
}
