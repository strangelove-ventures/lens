package cmd

import (
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

// appState is the modifiable state of the application.
type appState struct {
	Viper *viper.Viper

	HomePath        string
	OverriddenChain string
	Debug           bool
	Config          *Config
}

// OverwriteConfig overwrites the config files on disk with the serialization of cfg,
// and it replaces a.Config with cfg.
//
// It is possible to use a brand new Config argument,
// but typically the argument is a.Config.
func (a *appState) OverwriteConfig(cfg *Config) error {
	home := a.Viper.GetString("home")
	cfgPath := path.Join(home, "config.yaml")
	if err := os.WriteFile(cfgPath, cfg.MustYAML(), 0600); err != nil {
		return err
	}

	a.Config = cfg
	log.Printf("updated lens configuration at %s", cfgPath)
	return nil
}
