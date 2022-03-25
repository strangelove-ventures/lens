package cmd

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// appState is the modifiable state of the application.
type appState struct {
	// Log is the root logger of the application.
	// Consumers are expected to store and use local copies of the logger
	// after modifying with the .With method.
	Log *zap.Logger

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
	a.Log.Info("Updated lens configuration", zap.String("path", cfgPath))
	return nil
}
