package cmd

import (
	"log/slog"
	"os"

	"github.com/benwsapp/rlgl/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initConfig(cmd *cobra.Command) (string, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("site")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		return "", err
	}

	return viper.ConfigFileUsed(), nil
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the rlgl web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}

		configPath, err := initConfig(cmd)
		if err != nil {
			return err
		}

		slog.Info("starting server", "addr", addr, "config", configPath)
		return server.Run(addr, configPath)
	},
}

func init() {
	runCmd.Flags().String("addr", ":8080", "address to bind the server to")
	runCmd.Flags().String("config", "", "path to site configuration file (defaults to site.yaml)")
	rootCmd.AddCommand(runCmd)
}
